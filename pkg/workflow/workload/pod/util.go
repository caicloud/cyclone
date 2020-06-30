package pod

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/workflow/common"

	"github.com/caicloud/cyclone/pkg/workflow/controller"
)

// Name generates a pod name from Workflow name and Stage name
func Name(wf, stg string) string {
	return fmt.Sprintf("%s-%s-%s", wf, stg, rand.String(5))
}

// InputContainerName generates a container name for input resolver container
func InputContainerName(index int) string {
	return fmt.Sprintf("i%d", index)
}

// OutputContainerName generates a container name for output resolver container
func OutputContainerName(index int) string {
	return fmt.Sprintf("%so%d", common.CycloneSidecarPrefix, index)
}

// ContainerName generate container names for pod.
func ContainerName(index int) string {
	return fmt.Sprintf("c%d", index)
}

// GetResourceVolumeName generates a volume name for a resource.
func GetResourceVolumeName(resourceName string) string {
	return fmt.Sprintf("rsc-%s", resourceName)
}

// GetExecutionContext gets execution context from WorkflowRun, if not found, use the
// default context in workflow controller configuration.
func GetExecutionContext(wfr *v1alpha1.WorkflowRun) *v1alpha1.ExecutionContext {
	if wfr.Spec.ExecutionContext != nil {
		return wfr.Spec.ExecutionContext
	}

	return &v1alpha1.ExecutionContext{
		Namespace: controller.Config.ExecutionContext.Namespace,
		PVC:       controller.Config.ExecutionContext.PVC,
	}
}

// MatchContainerGroup matches a container name against a ContainerGroup, if the container belongs to the container group,
// return true, otherwise false.  It only tests containers, init containers are not considered here. If input container
// group is empty or invalid, return true.
func MatchContainerGroup(group v1alpha1.ContainerGroup, container string) bool {
	switch group {
	case v1alpha1.ContainerGroupAll:
		return true
	case v1alpha1.ContainerGroupSidecar:
		return strings.HasPrefix(container, common.CycloneSidecarPrefix)
	case v1alpha1.ContainerGroupWorkload:
		return !strings.HasPrefix(container, common.CycloneSidecarPrefix)
	default:
		return true
	}
}
