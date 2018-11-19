package common

import "fmt"

const (
	// The path that we will mount PV on in container.
	StageMountPath = "/__cyclone__workspace"

	// The mount path of the emptyDir shared by workload containers and coordinate container.
	// It's used to transfer output artifacts.
	StageEmptyDirMounthPath = "/__cyclone__emptydir"
)

// StagePath gets the path of a stage in PV
func StagePath(wfr, stage string) string {
	return fmt.Sprintf("workflowruns/%s/stages/%s", wfr, stage)
}

// ArtifactPath gets the path of a artifact in PV.
func ArtifactPath(wfr, stage, artifact string) string {
	return fmt.Sprintf("workflowruns/%s/stages/%s/artifacts/%s", wfr, stage, artifact)
}
