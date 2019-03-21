package workflowrun

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/caicloud/cyclone/pkg/workflow/common"
)

// PodName generates a pod name from Workflow name and Stage name
func PodName(wf, stg string) string {
	return fmt.Sprintf("%s-%s-%s", wf, stg, rand.String(5))
}

// GCPodName generates a pod name for GC pod
func GCPodName(wfr string) string {
	return fmt.Sprintf("wfrgc--%s", wfr)
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
