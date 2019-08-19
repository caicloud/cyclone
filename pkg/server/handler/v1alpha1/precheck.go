package v1alpha1

import (
	"context"
	"strings"

	"github.com/caicloud/nirvana/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/server/biz/integration/cluster"
	svrcommon "github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/config"
	"github.com/caicloud/cyclone/pkg/server/handler"
)

// Precheck checks worker cluster status
func Precheck(ctx context.Context, tenant string, checklist string) (*WorkersContextStatus, error) {
	lists := strings.Split(checklist, ",")
	workersContext := &WorkersContextStatus{
		Workers: []*WorkerContextStatus{},
	}

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

		workerContext := &WorkerContextStatus{
			Integration: integration.Name,
			Cluster:     cluster.ClusterName,
			Namespace:   cluster.Namespace,
			Opened:      cluster.IsWorkerCluster,
		}
		workersContext.Workers = append(workersContext.Workers, workerContext)

		if !cluster.IsWorkerCluster {
			log.Infof("Cluster %s is not a worker cluster, skip", integration.Name)
			continue
		}

		var client *kubernetes.Clientset
		var err error
	check:
		for _, item := range lists {
			switch item {
			case "ReservedResources":
				workerContext.ReservedResources = config.Config.StorageUsageWatcher.ResourceRequirements
			case "PVC":
				if client == nil {
					client, err = common.NewClusterClient(&cluster.Credential, cluster.IsControlCluster)
					if err != nil {
						log.Warningf("New cluster client for integration %s error %v", integration.Name, err)
						continue check
					}
				}

				pvcName := cluster.PVC
				if pvcName == "" {
					pvcName = svrcommon.TenantPVC(tenant)
				}
				pvc, err := client.CoreV1().PersistentVolumeClaims(cluster.Namespace).Get(pvcName, metav1.GetOptions{})
				if err != nil {
					log.Warningf("Get PVC %s in namespace %s", pvcName, cluster.Namespace)
					continue check
				}
				workerContext.PVC = &pvc.Status

			}
		}
	}

	return workersContext, nil
}

// WorkerContextStatus describes the status of worker clusters, it contains information that affects
// pipeline execution, like reserved resources, pvc status.
type WorkerContextStatus struct {
	Cluster           string                              `json:"cluster"`
	Namespace         string                              `json:"namespace"`
	Integration       string                              `json:"integration"`
	Opened            bool                                `json:"opened"`
	ReservedResources map[corev1.ResourceName]string      `json:"reservedResources"`
	PVC               *corev1.PersistentVolumeClaimStatus `json:"pvc"`
}

// WorkersContextStatus contains a set of WorkerContextStatus–
type WorkersContextStatus struct {
	Workers []*WorkerContextStatus `json:"workers"`
}
