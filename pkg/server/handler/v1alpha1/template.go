package v1alpha1

import (
	"context"
	"strings"

	"github.com/caicloud/nirvana/log"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/config"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/types"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

const (
	// TemplateTypeBuildin represents the type of buildin templates.
	TemplateTypeBuildin = "buildin"

	// TemplateTypeCustom represents the type of custom templates.
	TemplateTypeCustom = "custom"
)

// ListTemplates get templates the given tenant has access to.
// - ctx Context of the reqeust
// - tenant Tenant
// - includePublic Whether to include system level stage templates, default to true
// - query Query params includes start, limit and filter.
func ListTemplates(ctx context.Context, tenant string, includePublic bool, query *types.QueryParams) (*types.ListResponse, error) {
	// TODO(ChenDe): Need a more efficient way to get paged items.
	templates, err := handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).List(metav1.ListOptions{
		LabelSelector: meta.StageTemplateSelector(),
	})
	if err != nil {
		log.Errorf("Get templates from k8s with tenant %s error: %v", tenant, err)
		return nil, err
	}

	items := templates.Items
	if includePublic {
		publicTemplates, err := handler.K8sClient.CycloneV1alpha1().Stages(config.GetSystemNamespace()).List(metav1.ListOptions{
			LabelSelector: meta.StageTemplateSelector() + "," + meta.BuiltinLabelSelector(),
		})
		if err != nil {
			log.Errorf("Get templates from system namespace %s error: %v", config.GetSystemNamespace(), err)
			return nil, err
		}

		items = append(items, publicTemplates.Items...)
	}

	results := []v1alpha1.Stage{}
	if query.Filter == "" {
		results = items
	} else {
		// Support multiple filters rules: name and type, separated with comma.
		filterParts := strings.Split(query.Filter, ",")
		filters := make(map[string]string)
		for _, part := range filterParts {
			kv := strings.Split(part, "=")
			if len(kv) != 2 {
				return nil, cerr.ErrorQueryParamNotCorrect.Error(query.Filter)
			}

			filters[kv[0]] = kv[1]
		}

		var selected bool
		for _, item := range items {
			selected = true
			for key, value := range filters {
				switch key {
				case "name":
					if !strings.Contains(item.Name, strings.ToLower(value)) {
						selected = false
					}
				case "type":
					if item.Labels != nil {
						// Templates will be skipped when meet one of the conditions:
						// * there is buildin label, and the query type is custom
						// * there is no buildin label, but the query type is buildin
						label, ok := item.Labels[meta.LabelBuiltin]
						log.Infof("label: %s, value: %s", label, value)
						if (ok && value == TemplateTypeCustom) || (!ok && value == TemplateTypeBuildin) {
							selected = false
							log.Infof("selected: %t", selected)
						}
					}
				}
			}

			if selected {
				results = append(results, item)
			}
		}
	}

	size := int64(len(results))
	if query.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.Stage{}), nil
	}

	end := query.Start + query.Limit
	if end > size {
		end = size
	}

	return types.NewListResponse(int(size), results[query.Start:end]), nil
}

// CreateTemplate creates a stage template for the tenant. 'stage' describe the template to create. Stage templates
// are special stages, with 'stage.cyclone.dev/template' label. If created successfully, return the create template.
func CreateTemplate(ctx context.Context, tenant string, stage *v1alpha1.Stage) (*v1alpha1.Stage, error) {
	modifiers := []CreationModifier{GenerateNameModifier}
	for _, modifier := range modifiers {
		err := modifier(tenant, "", "", stage)
		if err != nil {
			return nil, err
		}
	}

	stage.Labels = meta.AddStageTemplateLabel(stage.Labels)
	return handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Create(stage)
}

// GetTemplate gets a stage template with the given template name under given tenant.
func GetTemplate(ctx context.Context, tenant, template string, includePublic bool) (stage *v1alpha1.Stage, err error) {
	// Convert the returned error if it is a k8s error.
	defer func() {
		cerr.ConvertK8sError(err)
	}()

	stage, err = handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Get(template, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}

		if includePublic {
			publicTemplate, err := handler.K8sClient.CycloneV1alpha1().Stages(config.GetSystemNamespace()).Get(template, metav1.GetOptions{})
			if err != nil {
				log.Errorf("Get templates from system namespace %s error: %v", config.GetSystemNamespace(), err)
				return nil, err
			}
			return publicTemplate, nil
		}

		return nil, err
	}

	return stage, nil
}

// UpdateTemplate updates a stage templates with the given tenant name and template name. If updated successfully, return
// the updated template.
func UpdateTemplate(ctx context.Context, tenant, template string, stage *v1alpha1.Stage) (*v1alpha1.Stage, error) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Get(template, metav1.GetOptions{})
		if err != nil {
			return err
		}
		newStage := origin.DeepCopy()
		newStage.Spec = stage.Spec
		newStage.Annotations = MergeMap(stage.Annotations, newStage.Annotations)
		newStage.Labels = MergeMap(stage.Labels, newStage.Labels)
		_, err = handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Update(newStage)
		return err
	})

	if err != nil {
		return nil, cerr.ConvertK8sError(err)
	}

	return stage, nil
}

// DeleteTemplate deletes a stage template with the given tenant and template name.
func DeleteTemplate(ctx context.Context, tenant, template string) error {
	err := handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Delete(template, &metav1.DeleteOptions{})

	return cerr.ConvertK8sError(err)
}
