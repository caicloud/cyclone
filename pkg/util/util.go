package util

import "github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"

// IsWorkflowRunTerminated judges whether the WorkflowRun has be terminated.
// Return true if terminated, otherwise return false.
func IsWorkflowRunTerminated(wfr *v1alpha1.WorkflowRun) bool {
	return IsPhaseTerminated(wfr.Status.Overall.Phase)
}

// IsPhaseTerminated judges whether the phase is terminated
func IsPhaseTerminated(phase v1alpha1.StatusPhase) bool {
	if phase == v1alpha1.StatusSucceeded ||
		phase == v1alpha1.StatusFailed ||
		phase == v1alpha1.StatusCancelled {
		return true
	}

	return false
}
