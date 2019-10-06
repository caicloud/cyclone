package v1alpha1

import (
	"context"
	"sort"
	"strings"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/meta"
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/statistic"
	"github.com/caicloud/cyclone/pkg/server/biz/utils"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/handler/v1alpha1/sorter"
	"github.com/caicloud/cyclone/pkg/server/types"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// ListProjects list projects the given tenant has access to.
// - ctx Context of the reqeust
// - tenant Tenant
// - query Query params includes start, limit and filter.
func ListProjects(ctx context.Context, tenant string, query *types.QueryParams) (*types.ListResponse, error) {
	// TODO(ChenDe): Need a more efficient way to get paged items.
	projects, err := handler.K8sClient.CycloneV1alpha1().Projects(common.TenantNamespace(tenant)).List(metav1.ListOptions{})
	if err != nil {
		log.Errorf("Get project from k8s with tenant %s error: %v", tenant, err)
		return nil, err
	}

	items := projects.Items
	var results []v1alpha1.Project
	if query.Filter == "" {
		results = items
	} else {
		results, err = filterProjects(items, query.Filter)
		if err != nil {
			return nil, err
		}
	}

	size := uint64(len(results))
	if query.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.Project{}), nil
	}

	end := query.Start + query.Limit
	if end > size {
		end = size
	}

	if query.Sort {
		sort.Sort(sorter.NewProjectSorter(results, query.Ascending))
	}

	items = results[query.Start:end]
	if query.Detail {
		for i := range items {
			workflows, err := handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).List(metav1.ListOptions{
				LabelSelector: meta.ProjectSelector(items[i].Name),
			})
			if err != nil {
				log.Warningf("Get workflows from k8s with tenant %s, project %s error: %v", tenant, items[i].Name, err)
				continue
			}
			items[i].Status = &v1alpha1.ProjectStatus{WorkflowCount: len(workflows.Items)}
		}
	}

	return types.NewListResponse(int(size), items), nil
}

func filterProjects(projects []v1alpha1.Project, filter string) ([]v1alpha1.Project, error) {
	if filter == "" {
		return projects, nil
	}

	var results []v1alpha1.Project
	kv := strings.Split(filter, "=")
	if len(kv) != 2 {
		return nil, cerr.ErrorQueryParamNotCorrect.Error(filter)
	}
	value := strings.ToLower(kv[1])

	if kv[0] == queryFilterName {
		for _, item := range projects {
			if strings.Contains(item.Name, value) {
				results = append(results, item)
			}
		}
	} else if kv[0] == queryFilterAlias {
		for _, item := range projects {
			if item.Annotations != nil {
				if alias, ok := item.Annotations[meta.AnnotationAlias]; ok {
					if strings.Contains(alias, value) {
						results = append(results, item)
					}
				}
			}
		}
	} else {
		results = projects
	}

	return results, nil
}

// CreateProject creates a project for the tenant.
func CreateProject(ctx context.Context, tenant string, project *v1alpha1.Project) (*v1alpha1.Project, error) {
	modifiers := []CreationModifier{GenerateNameModifier}
	for _, modifier := range modifiers {
		err := modifier(tenant, "", "", project)
		if err != nil {
			return nil, err
		}
	}

	return handler.K8sClient.CycloneV1alpha1().Projects(common.TenantNamespace(tenant)).Create(project)
}

// GetProject gets a project with the given project name under given tenant.
func GetProject(ctx context.Context, tenant, name string) (*v1alpha1.Project, error) {
	project, err := handler.K8sClient.CycloneV1alpha1().Projects(common.TenantNamespace(tenant)).Get(name, metav1.GetOptions{})
	if err != nil {
		log.Infof("get project %v of tenant %v error %v", name, tenant, err)
		return nil, cerr.ConvertK8sError(err)
	}
	return project, nil
}

// UpdateProject updates a project with the given tenant name and project name. If updated successfully, return
// the updated project.
func UpdateProject(ctx context.Context, tenant, pName string, project *v1alpha1.Project) (*v1alpha1.Project, error) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := handler.K8sClient.CycloneV1alpha1().Projects(common.TenantNamespace(tenant)).Get(pName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		newProject := origin.DeepCopy()
		newProject.Spec = project.Spec
		newProject.Annotations = utils.MergeMap(project.Annotations, newProject.Annotations)
		newProject.Labels = utils.MergeMap(project.Labels, newProject.Labels)
		_, err = handler.K8sClient.CycloneV1alpha1().Projects(common.TenantNamespace(tenant)).Update(newProject)
		return err
	})

	if err != nil {
		return nil, cerr.ConvertK8sError(err)
	}

	return project, nil
}

// DeleteProject deletes a project with the given tenant and project name.
func DeleteProject(ctx context.Context, tenant, project string) error {
	err := deleteCollections(tenant, project)
	if err != nil {
		return err
	}

	err = handler.K8sClient.CycloneV1alpha1().Projects(common.TenantNamespace(tenant)).Delete(project, &metav1.DeleteOptions{})
	return cerr.ConvertK8sError(err)
}

// GetProjectStatistics handles the request to get a project's statistics.
func GetProjectStatistics(ctx context.Context, tenant, project, start, end string) (*api.Statistic, error) {
	wfrs, err := handler.K8sClient.CycloneV1alpha1().WorkflowRuns(common.TenantNamespace(tenant)).List(metav1.ListOptions{
		LabelSelector: meta.ProjectSelector(project),
	})
	if err != nil {
		return nil, err
	}

	return statistic.Stats(wfrs, start, end)
}
