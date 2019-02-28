package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/nirvana/log"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/types"
	wfcommon "github.com/caicloud/cyclone/pkg/workflow/common"
)

// ListTemplates get templates the given tenant has access to.
// - ctx Context of the reqeust
// - tenant Tenant
// - includePublic Whether to include system level stage templates, default to true
// - pagination Pagination with page and limit.
func ListTemplates(ctx context.Context, tenant string, includePublic bool, pagination *types.Pagination) (*types.ListResponse, error) {
	// TODO(ChenDe): Need a more efficient way to get paged items.
	templates, err := handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).List(metav1.ListOptions{
		LabelSelector: wfcommon.StageTemplateLabelSelector,
	})
	if err != nil {
		log.Errorf("Get templates from k8s with tenant %s error: %v", tenant, err)
		return nil, err
	}

	items := templates.Items
	if tenant != common.AdminTenant && includePublic {
		publicTemplates, err := handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(common.AdminTenant)).List(metav1.ListOptions{
			LabelSelector: wfcommon.StageTemplateLabelSelector,
		})
		if err != nil {
			log.Errorf("Get templates from k8s with tenant %s error: %v", common.AdminTenant, err)
			return nil, err
		}

		items = append(items, publicTemplates.Items...)
	}

	size := int64(len(items))
	if pagination.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.Stage{}), nil
	}

	end := pagination.Start + pagination.Limit
	if end > size {
		end = size
	}
	return types.NewListResponse(int(size), items[pagination.Start:end]), nil
}

// CreateTemplate creates a stage template for the tenant. 'stage' describe the template to create. Stage templates
// are special stages, with 'cyclone.io/stage-template' label. If created successfully, return the create template.
func CreateTemplate(ctx context.Context, tenant string, stage *v1alpha1.Stage) (*v1alpha1.Stage, error) {
	modifiers := []CreationModifier{GenerateNameModifier}
	for _, modifier := range modifiers {
		err := modifier("", tenant, stage)
		if err != nil {
			return nil, err
		}
	}

	LabelStageTemplate(stage)
	LabelCustomizedTemplate(stage)
	return handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Create(stage)
}

// GetTemplate gets a stage template with the given template name under given tenant.
func GetTemplate(ctx context.Context, tenant, template string, includePublic bool) (*v1alpha1.Stage, error) {
	stage, err := handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Get(template, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}

		if tenant != common.AdminTenant && includePublic {
			publicTemplate, err := handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(common.AdminTenant)).Get(template, metav1.GetOptions{})
			if err != nil {
				log.Errorf("Get templates from k8s with tenant %s error: %v", common.AdminTenant, err)
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
		newStage.Annotations = UpdateAnnotations(stage.Annotations, newStage.Annotations)
		_, err = handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Update(newStage)
		return err
	})

	if err != nil {
		return nil, err
	}

	return stage, nil
}

// DeleteTemplate deletes a stage template with the given tenant and template name.
func DeleteTemplate(ctx context.Context, tenant, template string) error {
	return handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Delete(template, &metav1.DeleteOptions{})
}
