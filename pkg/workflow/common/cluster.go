package common

import (
	"k8s.io/client-go/kubernetes"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/workflow/controller/store"
)

// GetExecutionClusterClient gets execution cluster client with the WorkflowRun
func GetExecutionClusterClient(wfr *v1alpha1.WorkflowRun) kubernetes.Interface {
	cluster := common.ControlClusterName
	if wfr.Spec.ExecutionContext != nil && wfr.Spec.ExecutionContext.Cluster != "" {
		cluster = wfr.Spec.ExecutionContext.Cluster
	}

	return store.GetClusterClient(cluster)
}

// IsMaster judge weather this wkf in master cluster or not
func IsMaster(wfr *v1alpha1.WorkflowRun) bool {

	if wfr.Spec.ExecutionContext != nil {
		name := wfr.Spec.ExecutionContext.Cluster

		return name == "" || name == common.ControlClusterName
	}

	return false
}
