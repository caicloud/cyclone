package common

import "fmt"

const (
	// StageMountPath is path that we will mount PV on in container.
	StageMountPath = "/__cyclone__workspace"

	// CoordinatorWorkspacePath is path of artifacts in coordinator container
	CoordinatorWorkspacePath = "/workspace/"
)

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
