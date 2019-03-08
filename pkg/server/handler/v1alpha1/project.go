package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/nirvana/log"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/types"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// ListProjects list projects the given tenant has access to.
// - ctx Context of the reqeust
// - tenant Tenant
// - pagination Pagination with page and limit.
func ListProjects(ctx context.Context, tenant string, pagination *types.Pagination) (*types.ListResponse, error) {
	// TODO(ChenDe): Need a more efficient way to get paged items.
	projects, err := handler.K8sClient.CycloneV1alpha1().Projects(common.TenantNamespace(tenant)).List(metav1.ListOptions{})
	if err != nil {
		log.Errorf("Get project from k8s with tenant %s error: %v", tenant, err)
		return nil, err
	}

	items := projects.Items

	size := int64(len(items))
	if pagination.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.Project{}), nil
	}

	end := pagination.Start + pagination.Limit
	if end > size {
		end = size
	}
	return types.NewListResponse(int(size), items[pagination.Start:end]), nil
}

// CreateProject creates a project for the tenant.
func CreateProject(ctx context.Context, tenant string, project *v1alpha1.Project) (*v1alpha1.Project, error) {
	modifiers := []CreationModifier{GenerateNameModifier}
	for _, modifier := range modifiers {
		err := modifier("", tenant, project)
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
		newProject.Annotations = UpdateAnnotations(project.Annotations, newProject.Annotations)
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
	err := handler.K8sClient.CycloneV1alpha1().Projects(common.TenantNamespace(tenant)).Delete(project, &metav1.DeleteOptions{})

	return cerr.ConvertK8sError(err)
}
