package v1alpha1

import (
	"context"
	"sort"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/server/biz/utils"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/handler/v1alpha1/sorter"
	"github.com/caicloud/cyclone/pkg/server/types"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// CreateStage ...
func CreateStage(ctx context.Context, project, tenant string, stg *v1alpha1.Stage) (*v1alpha1.Stage, error) {
	modifiers := []CreationModifier{GenerateNameModifier, InjectProjectLabelModifier, InjectProjectOwnerRefModifier}
	for _, modifier := range modifiers {
		err := modifier(tenant, project, "", stg)
		if err != nil {
			return nil, err
		}
	}

	return handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Create(stg)
}

// ListStages ...
func ListStages(ctx context.Context, project, tenant string, query *types.QueryParams) (*types.ListResponse, error) {
	stages, err := handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).List(metav1.ListOptions{
		LabelSelector: meta.ProjectSelector(project),
	})
	if err != nil {
		log.Errorf("Get stages from k8s with tenant %s, project %s error: %v", tenant, project, err)
		return nil, err
	}

	items := stages.Items
	size := uint64(len(items))
	if query.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.Stage{}), nil
	}

	end := query.Start + query.Limit
	if end > size {
		end = size
	}

	if query.Sort {
		sort.Sort(sorter.NewStageSorter(items, query.Ascending))
	}

	return types.NewListResponse(int(size), items[query.Start:end]), nil
}

// GetStage ...
func GetStage(ctx context.Context, project, stage, tenant string) (*v1alpha1.Stage, error) {
	stg, err := handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Get(stage, metav1.GetOptions{})

	return stg, cerr.ConvertK8sError(err)
}

// UpdateStage ...
func UpdateStage(ctx context.Context, project, stage, tenant string, stg *v1alpha1.Stage) (*v1alpha1.Stage, error) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Get(stage, metav1.GetOptions{})
		if err != nil {
			return err
		}
		newStg := origin.DeepCopy()
		newStg.Spec = stg.Spec
		newStg.Annotations = utils.MergeMap(stg.Annotations, newStg.Annotations)
		newStg.Labels = utils.MergeMap(stg.Labels, newStg.Labels)
		_, err = handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Update(newStg)
		return err
	})

	if err != nil {
		return nil, cerr.ConvertK8sError(err)
	}

	return stg, nil
}

// DeleteStage ...
func DeleteStage(ctx context.Context, project, stage, tenant string) error {
	err := handler.K8sClient.CycloneV1alpha1().Stages(common.TenantNamespace(tenant)).Delete(stage, nil)
	return cerr.ConvertK8sError(err)
}
