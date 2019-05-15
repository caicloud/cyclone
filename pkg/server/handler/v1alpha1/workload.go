package v1alpha1

import (
	"context"
	"fmt"

	"github.com/caicloud/nirvana/log"
	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/server/biz/integration/cluster"
	svrcommon "github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/types"
)

// ListWorkingPods lists all pods of workflowruns.
func ListWorkingPods(ctx context.Context, tenant string, query *types.QueryParams) (*types.ListResponse, error) {
	integrations, err := cluster.GetSchedulableClusters(handler.K8sClient, tenant)
	if err != nil {
		return nil, err
	}
	if len(integrations) != 1 {
		return nil, fmt.Errorf("expect one schedulable cluster, but %d found", len(integrations))
	}
	cluster := integrations[0].Spec.Cluster
	if cluster == nil {
		return nil, fmt.Errorf("schedulable cluster info is empty for tenant: %s", tenant)
	}

	client, err := common.NewClusterClient(&cluster.Credential, cluster.IsControlCluster)
	if err != nil {
		return nil, fmt.Errorf("create cluster client error: %v", err)
	}

	pods, err := client.CoreV1().Pods(svrcommon.TenantNamespace(tenant)).List(metav1.ListOptions{
		LabelSelector: meta.LabelExistsSelector(meta.LabelWorkflowRunName),
	})
	if err != nil {
		log.Errorf("Failed to list pods for tenant %s as error: %v", tenant, err)
		return nil, err
	}

	items := pods.Items
	size := int64(len(items))
	if query.Start >= size {
		return types.NewListResponse(int(size), []core_v1.Pod{}), nil
	}

	end := query.Start + query.Limit
	if end > size {
		end = size
	}

	return types.NewListResponse(int(size), items[query.Start:end]), nil
}
