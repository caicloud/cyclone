package cluster

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	cycloneV1alpha1 "github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	cycloneCommon "github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// ReconcileTriggerExecutionContext reconciles execution context for workflow triggers.
// Update WorkflowTriggers' unavailable execution context (related cluster closed or
// deleted) with an available one (related cluster opened).
func ReconcileTriggerExecutionContext(tenant string) error {
	// List all available worker clusters
	availableExecutionContexts, err := listAvailableExecutionContext(tenant)
	if err != nil {
		return err
	}

	log.Infof("Available execution context: %v", availableExecutionContexts)
	if len(availableExecutionContexts) == 0 {
		return cerr.ErrorUnknownInternal.Error(fmt.Sprintf("No execution context available for tenant: %s", tenant))
	}

	// List all WorkflowTriggers of the tenant
	workflowTriggers, err := handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("List workflowtriggers from k8s for tenant %s with error: %v", tenant, err)
		return err
	}

	var goerr error
	wg := sync.WaitGroup{}
	for _, workflowTrigger := range workflowTriggers.Items {
		wg.Add(1)

		go func(wft cycloneV1alpha1.WorkflowTrigger) {
			defer wg.Done()

			var needUpdate = false
			if wft.Spec.WorkflowRunSpec.ExecutionContext == nil {
				needUpdate = true
			} else if _, ok := availableExecutionContexts[wft.Spec.WorkflowRunSpec.ExecutionContext.Cluster]; !ok {
				needUpdate = true
			}
			if !needUpdate {
				return
			}

			executionContext := randExecutionContext(availableExecutionContexts)
			if executionContext == nil {
				goerr = fmt.Errorf("get random execution context error")
				return
			}
			log.Infof("WorkflowTrigger %s's execution context need to be updated to cluster %s", wft.Name, executionContext.Cluster)

			newWft := wft.DeepCopy()
			newWft.Spec.WorkflowRunSpec.ExecutionContext = executionContext
			_, err := handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Update(context.TODO(), newWft, metav1.UpdateOptions{})
			if err == nil {
				// Update wft succeeded, return directly.
				return
			}

			// Update workflowTrigger failed, start to retry.
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				origin, err := handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Get(context.TODO(), wft.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}
				newWft := origin.DeepCopy()
				newWft.Spec.WorkflowRunSpec.ExecutionContext = executionContext
				_, err = handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Update(context.TODO(), newWft, metav1.UpdateOptions{})
				return err
			})
			if err != nil {
				goerr = cerr.ConvertK8sError(err)
				return
			}
		}(workflowTrigger)
	}

	wg.Wait()
	if goerr != nil {
		return goerr
	}
	return nil
}

func listAvailableExecutionContext(tenant string) (map[string]*cycloneV1alpha1.ExecutionContext, error) {
	executionContexts := make(map[string]*cycloneV1alpha1.ExecutionContext)

	integrations, err := GetSchedulableClusters(handler.K8sClient, tenant)
	if err != nil {
		return executionContexts, err
	}

	for _, integration := range integrations {
		cluster := integration.Spec.Cluster
		if cluster == nil {
			continue
		}

		if !cluster.IsWorkerCluster {
			continue
		}

		executionContexts[executionContextClusterName(cluster.ClusterName, cluster.IsControlCluster)] =
			construstExecutionContext(cluster, tenant)
	}

	return executionContexts, nil
}

func executionContextClusterName(originClusterName string, isControlCluster bool) string {
	if isControlCluster {
		return cycloneCommon.ControlClusterName
	}

	return originClusterName
}

func construstExecutionContext(clusterSource *v1alpha1.ClusterSource, tenant string) *cycloneV1alpha1.ExecutionContext {
	executionContext := &cycloneV1alpha1.ExecutionContext{
		Cluster:   executionContextClusterName(clusterSource.ClusterName, clusterSource.IsControlCluster),
		Namespace: clusterSource.Namespace,
		PVC:       clusterSource.PVC,
	}

	if executionContext.Namespace == "" {
		executionContext.Namespace = common.TenantNamespace(tenant)
	}
	if executionContext.PVC == "" {
		executionContext.PVC = common.TenantPVC(tenant)
	}

	return executionContext
}

func randExecutionContext(clusters map[string]*cycloneV1alpha1.ExecutionContext) *cycloneV1alpha1.ExecutionContext {
	if clusters == nil {
		return nil
	}

	i := rand.Intn(len(clusters))
	for k := range clusters {
		if i == 0 {
			return clusters[k]
		}
		i--
	}

	return nil
}
