package util

import (
	"testing"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
)

func TestIsWorkflowRunTerminated(t *testing.T) {
	wfr := &v1alpha1.WorkflowRun{
		Status: v1alpha1.WorkflowRunStatus{
			Overall: v1alpha1.Status{},
		},
	}

	testCases := map[v1alpha1.StatusPhase]struct {
		phase    v1alpha1.StatusPhase
		expected bool
	}{
		v1alpha1.StatusPending: {
			v1alpha1.StatusPending,
			false,
		},
		v1alpha1.StatusRunning: {
			v1alpha1.StatusRunning,
			false,
		},
		v1alpha1.StatusWaiting: {
			v1alpha1.StatusWaiting,
			false,
		},
		v1alpha1.StatusSucceeded: {
			v1alpha1.StatusSucceeded,
			true,
		},
		v1alpha1.StatusFailed: {
			v1alpha1.StatusFailed,
			true,
		},
		v1alpha1.StatusCancelled: {
			v1alpha1.StatusCancelled,
			true,
		},
	}

	for d, tc := range testCases {
		wfr.Status.Overall.Phase = tc.phase
		result := IsWorkflowRunTerminated(wfr)
		if result != tc.expected {
			t.Errorf("Fail to judge the status for %s: expect %t, but got %t", d, tc.expected, result)
		}
	}
}
