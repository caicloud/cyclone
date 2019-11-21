package v1alpha1

import (
	"context"

	"github.com/caicloud/nirvana/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/common"
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/integration/cluster"
	"github.com/caicloud/cyclone/pkg/server/biz/usage"
	svrcommon "github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/config"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/types"
)

// ListExecutionContexts list execution contexts of a tenant
func ListExecutionContexts(ctx context.Context, tenant string) (*types.ListResponse, error) {
	var executionContexts []api.ExecutionContext

	integrations, err := cluster.GetSchedulableClusters(handler.K8sClient, tenant)
	if err != nil {
		return nil, err
	}

	for _, integration := range integrations {
		cluster := integration.Spec.Cluster
		if cluster == nil {
			log.Warningf("Cluster of integration %s is nil", integration.Name)
			continue
		}

		pvcName := cluster.PVC
		if pvcName == "" {
			pvcName = svrcommon.TenantPVC(tenant)
		}

		executionContext := api.ExecutionContext{
			Spec: api.ExecutionContextSpec{
				Cluster:   cluster.ClusterName,
				Namespace: cluster.Namespace,
				PVC:       pvcName,
			},
			Status: api.ExecutionContextStatus{
				Phase:             api.ExecutionContextNotUnknown,
				ReservedResources: config.Config.StorageUsageWatcher.ResourceRequirements,
			},
		}

		if cluster.IsWorkerCluster {
			overallStatus, err := getPVCOverallStatus(tenant, pvcName, cluster)
			if err != nil {
				log.Warningf("Get PVC status in %s/%s", cluster.ClusterName, cluster.Namespace)
			} else {
				executionContext.Status.PVC = overallStatus
				if overallStatus.Status != nil && overallStatus.Status.Phase == corev1.ClaimBound {
					executionContext.Status.Phase = api.ExecutionContextReady
				} else {
					executionContext.Status.Phase = api.ExecutionContextNotReady
				}
			}
		} else {
			executionContext.Status.Phase = api.ExecutionContextClosed
		}

		executionContexts = append(executionContexts, executionContext)
	}

	return types.NewListResponse(len(executionContexts), executionContexts), nil
}

func getPVCOverallStatus(tenant, pvcName string, cluster *api.ClusterSource) (overallStatus api.PVCOverallStatus, err error) {
	overallStatus.Status, err = getPVCStatus(tenant, pvcName, cluster)
	if err != nil {
		return overallStatus, err
	}

	reporter, err := usage.NewPVCReporter(handler.K8sClient, tenant)
	if err != nil {
		log.Warningf("Create pvc reporter for tenant %s error: %v", tenant, err)
	} else {
		overallStatus.Usage = &api.PVCUsageStatus{
			PVCUsage:       reporter.ReadableUsage().PVCUsage,
			UsedPercentage: reporter.OverallUsedPercentage(),
		}
	}

	return overallStatus, nil
}

func getPVCStatus(tenant, pvcName string, cluster *api.ClusterSource) (*corev1.PersistentVolumeClaimStatus, error) {
	client, err := common.NewClusterClient(&cluster.Credential, cluster.IsControlCluster)
	if err != nil {
		return nil, err
	}

	pvc, err := client.CoreV1().PersistentVolumeClaims(cluster.Namespace).Get(pvcName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &pvc.Status, nil
}
