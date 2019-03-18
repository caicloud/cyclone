package v1alpha1

import (
	"context"

	"github.com/caicloud/nirvana/log"
	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/types"
)

// ListRunningPods lists running pods of workflowruns for one tenant.
func ListRunningPods(ctx context.Context, tenant string, pagination *types.Pagination) (*types.ListResponse, error) {
	pods, err := handler.K8sClient.CoreV1().Pods(common.TenantNamespace(tenant)).List(metav1.ListOptions{})
	if err != nil {
		log.Errorf("Failed to list pods for tenant %s as error: %v", tenant, err)
		return nil, err
	}

	items := pods.Items
	size := int64(len(items))
	if pagination.Start >= size {
		return types.NewListResponse(int(size), []core_v1.Pod{}), nil
	}

	end := pagination.Start + pagination.Limit
	if end > size {
		end = size
	}

	return types.NewListResponse(int(size), items[pagination.Start:end]), nil
}
