package pod

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
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

// ResolveRefStringValue resolves the given secret ref value, if it's not a ref value, return the origin value.
// Ref value is in format of '$.<ns>.<secret>/<jsonpath>/...' to refer value in a secret.
func ResolveRefStringValue(ref string, client clientset.Interface) (string, error) {
	refValue := NewSecretRefValue()

	// Return the origin value if not a valid ref
	if err := refValue.Parse(ref); err != nil {
		return ref, nil
	}

	v, err := refValue.Resolve(client)
	if err != nil {
		return "", err
	}

	strV, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("expect string value: %v", v)
	}

	return strV, nil
}
