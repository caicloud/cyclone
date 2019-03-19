package common

import (
	"fmt"
	"strings"
)

// TenantNamespace gets namespace from given tenant name.
func TenantNamespace(tenant string) string {
	return fmt.Sprintf("%s%s", TenantNamespacePrefix, tenant)
}

// NamespaceTenant retrieves tenant from given namespace name.
func NamespaceTenant(n string) string {
	return strings.TrimPrefix(n, TenantNamespacePrefix)
}

// TenantPVC returns pvc name related to the tenant
func TenantPVC(tenant string) string {
	return TenantPVCPrefix + tenant
}

// TenantResourceQuota returns resource quota name related the tenant
func TenantResourceQuota(tenant string) string {
	return tenant
}

// LabelOwnerCyclone returns a label string describes resource belongs to cyclone
func LabelOwnerCyclone() string {
	return LabelOwner + "=" + OwnerCyclone
}

// IntegrationSecret returns secret name related to the integration
func IntegrationSecret(i string) string {
	return i
}

// SecretIntegration returns integration name related to the secret
func SecretIntegration(s string) string {
	return s
}

// WorkerClustersSelector is a selector for clusters which are use to perform workload
func WorkerClustersSelector() string {
	return LabelClusterOn + "=" + LabelTrueValue
}

// ProjectSelector is a selector for cyclone CRD resources which have corresponding project label
func ProjectSelector(project string) string {
	return LabelProjectName + "=" + project
}

// WorkflowSelector is a selector for cyclone CRD resources which have corresponding workflow label
func WorkflowSelector(workflow string) string {
	return LabelWorkflowName + "=" + workflow
}

// InputResourceVolumeName ...
func InputResourceVolumeName(name string) string {
	return "input-" + name
}

// OutputResourceVolumeName ...
func OutputResourceVolumeName(name string) string {
	return "output-" + name
}

// WorkflowRunsPath indicates WorkflowRuns data path in PV
func WorkflowRunsPath() string {
	return "workflowruns"
}

// StagePath gets the path of a stage in PV
func StagePath(wfr, stage string) string {
	return fmt.Sprintf("workflowruns/%s/stages/%s", wfr, stage)
}

// ArtifactsPath gets the path of artifacts in PV
func ArtifactsPath(wfr, stage string) string {
	return fmt.Sprintf("workflowruns/%s/stages/%s/artifacts/", wfr, stage)
}

// ArtifactPath gets the path of a artifact in PV.
func ArtifactPath(wfr, stage, artifact string) string {
	return fmt.Sprintf("workflowruns/%s/stages/%s/artifacts/%s", wfr, stage, artifact)
}

// ResourcePath gets the path of a resource in PV
func ResourcePath(wfr, resource string) string {
	return fmt.Sprintf("workflowruns/%s/resources/%s", wfr, resource)
}

// ContainerSelector is a function to select containers
type ContainerSelector func(name string) bool

// OnlyWorkload selects only workload containers.
func OnlyWorkload(name string) bool {
	if strings.HasPrefix(name, CycloneSidecarPrefix) {
		return false
	}

	if strings.HasPrefix(name, WorkloadSidecarPrefix) {
		return false
	}

	return true
}

// AllContainers selects all containers, it returns true regardless of the container name.
func AllContainers(string) bool {
	return true
}

// OnlyCustomContainer judges whether a container is a custom container based on container name.
// Containers added by Cyclone would have CycloneSidecarPrefix prefix in container names.
func OnlyCustomContainer(name string) bool {
	return !strings.HasPrefix(name, CycloneSidecarPrefix)
}

// NonWorkloadSidecar selects all containers except workload sidecars.
func NonWorkloadSidecar(name string) bool {
	if strings.HasPrefix(name, WorkloadSidecarPrefix) {
		return false
	}

	return true
}

// NonCoordinator selects all containers except coordinator.
func NonCoordinator(name string) bool {
	return name != CoordinatorSidecarName
}

// NonDockerInDocker selects all containers except docker:dind.
func NonDockerInDocker(name string) bool {
	return name != DockerInDockerSidecarName
}

// Pass check whether the given container name passes the given selectors.
func Pass(name string, selectors []ContainerSelector) bool {
	for _, s := range selectors {
		if !s(name) {
			return false
		}
	}
	return true
}
