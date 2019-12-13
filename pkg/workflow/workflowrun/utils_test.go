package workflowrun

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
)

func TestResolveStatus(t *testing.T) {
	latest := &v1alpha1.Status{
		Phase: v1alpha1.StatusSucceeded,
	}
	update := &v1alpha1.Status{
		Phase: v1alpha1.StatusRunning,
	}
	expected := &v1alpha1.Status{
		Phase: v1alpha1.StatusSucceeded,
	}
	result := resolveStatus(latest, update)
	assert.Equal(t, expected, result)

	latest = &v1alpha1.Status{
		Phase: v1alpha1.StatusRunning,
	}
	update = &v1alpha1.Status{
		Phase: v1alpha1.StatusSucceeded,
	}
	expected = &v1alpha1.Status{
		Phase: v1alpha1.StatusSucceeded,
	}
	result = resolveStatus(latest, update)
	assert.Equal(t, expected, result)

	latest = &v1alpha1.Status{
		Phase: v1alpha1.StatusFailed,
	}
	update = &v1alpha1.Status{
		Phase: v1alpha1.StatusSucceeded,
	}
	expected = &v1alpha1.Status{
		Phase: v1alpha1.StatusFailed,
	}
	result = resolveStatus(latest, update)
	assert.Equal(t, expected, result)

	now := metav1.Time{Time: time.Now()}
	old := metav1.Time{Time: time.Now().Add(-time.Second * 10)}
	latest = &v1alpha1.Status{
		Phase:              v1alpha1.StatusRunning,
		LastTransitionTime: now,
	}
	update = &v1alpha1.Status{
		Phase:              v1alpha1.StatusRunning,
		LastTransitionTime: old,
	}
	expected = &v1alpha1.Status{
		Phase:              v1alpha1.StatusRunning,
		LastTransitionTime: now,
	}
	result = resolveStatus(latest, update)
	assert.Equal(t, expected, result)
}

func TestIsTrivial(t *testing.T) {
	wf := &v1alpha1.Workflow{
		Spec: v1alpha1.WorkflowSpec{
			Stages: []v1alpha1.StageItem{
				{
					Name:    "A",
					Trivial: false,
				},
				{
					Name:    "B",
					Trivial: false,
					Depends: []string{"A"},
				},
				{
					Name:    "C",
					Trivial: true,
				},
			},
		},
	}
	assert.Equal(t, false, IsTrivial(wf, "A"))
	assert.Equal(t, false, IsTrivial(wf, "B"))
	assert.Equal(t, true, IsTrivial(wf, "C"))
}

func TestNextStages(t *testing.T) {
	wf := &v1alpha1.Workflow{
		Spec: v1alpha1.WorkflowSpec{
			Stages: []v1alpha1.StageItem{
				{
					Name: "A",
				},
				{
					Name:    "B",
					Depends: []string{"A"},
				},
				{
					Name: "C",
				},
			},
		},
	}
	wfr := &v1alpha1.WorkflowRun{
		Status: v1alpha1.WorkflowRunStatus{
			Stages: map[string]*v1alpha1.StageStatus{
				"A": {
					Status: v1alpha1.Status{Phase: v1alpha1.StatusSucceeded},
				},
			},
		},
	}
	expected := []string{"B", "C"}
	nexts, nonRetryStages := NextStages(0, 1, wf, wfr)
	assert.Equal(t, expected, nexts)
	assert.Nil(t, nonRetryStages)

	wf = &v1alpha1.Workflow{
		Spec: v1alpha1.WorkflowSpec{
			Stages: []v1alpha1.StageItem{
				{
					Name: "A",
				},
				{
					Name:    "B",
					Depends: []string{"A"},
				},
				{
					Name: "C",
				},
			},
		},
	}
	wfr = &v1alpha1.WorkflowRun{
		Status: v1alpha1.WorkflowRunStatus{
			Stages: map[string]*v1alpha1.StageStatus{
				"A": {
					Status: v1alpha1.Status{Phase: v1alpha1.StatusFailed},
				},
			},
		},
	}

	expected = []string{"C"}
	nexts, nonRetryStages = NextStages(0, 1, wf, wfr)
	assert.Equal(t, expected, nexts)
	assert.Nil(t, nonRetryStages)

	wf = &v1alpha1.Workflow{
		Spec: v1alpha1.WorkflowSpec{
			Stages: []v1alpha1.StageItem{
				{
					Name: "A",
				},
				{
					Name:    "B",
					Depends: []string{"A"},
				},
				{
					Name: "C",
				},
			},
		},
	}
	wfr = &v1alpha1.WorkflowRun{
		Status: v1alpha1.WorkflowRunStatus{
			Stages: map[string]*v1alpha1.StageStatus{
				"A": {
					Status: v1alpha1.Status{Phase: v1alpha1.StatusRunning},
				},
			},
		},
	}

	expected = []string{"C"}
	nexts, nonRetryStages = NextStages(0, 1, wf, wfr)
	assert.Equal(t, expected, nexts)
	assert.Nil(t, nonRetryStages)

	wf = &v1alpha1.Workflow{
		Spec: v1alpha1.WorkflowSpec{
			Stages: []v1alpha1.StageItem{
				{
					Name: "A",
				},
				{
					Name:    "B",
					Depends: []string{"A"},
				},
				{
					Name: "C",
				},
			},
		},
	}
	wfr = &v1alpha1.WorkflowRun{
		Status: v1alpha1.WorkflowRunStatus{
			Stages: map[string]*v1alpha1.StageStatus{
				"B": {
					Status: v1alpha1.Status{
						Phase: v1alpha1.StatusPending,
						RetryStatus: &v1alpha1.RetryStatus{
							Times:     3,
							StartTime: metav1.Time{Time: time.Now()},
						},
					},
				},
				"A": {
					Status: v1alpha1.Status{Phase: v1alpha1.StatusSucceeded},
				},
			},
		},
	}

	expected = []string{"B", "C"}
	nexts, nonRetryStages = NextStages(300, 4, wf, wfr)
	assert.Equal(t, expected, nexts)
	assert.Nil(t, nonRetryStages)

	wf = &v1alpha1.Workflow{
		Spec: v1alpha1.WorkflowSpec{
			Stages: []v1alpha1.StageItem{
				{
					Name: "A",
				},
				{
					Name:    "B",
					Depends: []string{"A"},
				},
				{
					Name: "C",
				},
			},
		},
	}
	wfr = &v1alpha1.WorkflowRun{
		Status: v1alpha1.WorkflowRunStatus{
			Stages: map[string]*v1alpha1.StageStatus{
				"B": {
					Status: v1alpha1.Status{
						Phase: v1alpha1.StatusPending,
						RetryStatus: &v1alpha1.RetryStatus{
							Times:     3,
							StartTime: metav1.Time{Time: time.Now()},
						},
					},
				},
				"A": {
					Status: v1alpha1.Status{Phase: v1alpha1.StatusSucceeded},
				},
			},
		},
	}

	expected = []string{"C"}
	expectedNonRetry := []string{"B"}
	nexts, nonRetryStages = NextStages(300, 2, wf, wfr)
	assert.Equal(t, expected, nexts)
	assert.Equal(t, expectedNonRetry, nonRetryStages)
}

func TestStaticStatus(t *testing.T) {
	now := metav1.Time{Time: time.Now()}
	zero := metav1.Time{Time: time.Unix(0, 0)}
	status := &v1alpha1.WorkflowRunStatus{
		Stages: map[string]*v1alpha1.StageStatus{
			"A": {
				Status: v1alpha1.Status{
					Phase:              v1alpha1.StatusRunning,
					LastTransitionTime: now,
				},
			},
		},
		Overall: v1alpha1.Status{
			Phase:              v1alpha1.StatusRunning,
			LastTransitionTime: now,
		},
	}
	actual := staticStatus(status)
	expected := &v1alpha1.WorkflowRunStatus{
		Stages: map[string]*v1alpha1.StageStatus{
			"A": {
				Status: v1alpha1.Status{
					Phase:              v1alpha1.StatusRunning,
					LastTransitionTime: zero,
				},
			},
		},
		Overall: v1alpha1.Status{
			Phase:              v1alpha1.StatusRunning,
			LastTransitionTime: zero,
		},
	}
	assert.Equal(t, expected, actual)
}

func TestString(t *testing.T) {
	item := workflowRunItem{
		name:      "test",
		namespace: "default",
	}
	assert.Equal(t, "default:test", item.String())
}

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

func TestGCPodName(t *testing.T) {
	assert.Equal(t, GCPodName("wfr"), "wfrgc--wfr")
}

func TestNextStageStatus(t *testing.T) {
	retryStatus := &v1alpha1.RetryStatus{
		Times:     2,
		StartTime: metav1.Time{Time: time.Now()},
	}

	statusPendingRetry := &v1alpha1.StageStatus{
		Status: v1alpha1.Status{
			Phase:       v1alpha1.StatusPending,
			RetryStatus: retryStatus,
		},
	}

	statusPending := &v1alpha1.StageStatus{
		Status: v1alpha1.Status{
			Phase: v1alpha1.StatusPending,
		},
	}

	stagesStatus := make(map[string]*v1alpha1.StageStatus)
	stagesStatus["retry"] = statusPendingRetry
	stagesStatus["pending"] = statusPending

	testCases := map[string]struct {
		stage               string
		expectedPhase       v1alpha1.StatusPhase
		expectedRetryStatus *v1alpha1.RetryStatus
	}{
		"retry": {
			stage:               "retry",
			expectedPhase:       v1alpha1.StatusRunning,
			expectedRetryStatus: retryStatus,
		},
		"normal": {
			stage:               "pending",
			expectedPhase:       v1alpha1.StatusRunning,
			expectedRetryStatus: nil,
		},
	}

	for d, tc := range testCases {
		result := NextStageStatus(stagesStatus, tc.stage)
		if result.Phase != tc.expectedPhase {
			t.Errorf("Fail to determine next stage status for %s: expect phase %s, but got %s", d, tc.expectedPhase, result.Phase)
		}
		if !assert.Equal(t, tc.expectedRetryStatus, result.RetryStatus) {
			t.Errorf("Fail to determine next status for %s: expect retry status %v, but got %v", d, tc.expectedRetryStatus, result.RetryStatus)
		}
	}
}
