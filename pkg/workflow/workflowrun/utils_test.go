package workflowrun

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset/fake"
)

func TestGetResourceVolumeName(t *testing.T) {
	assert.Equal(t, "rsc-git", GetResourceVolumeName("git"))
}

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
	nexts := NextStages(wf, wfr)
	assert.Equal(t, expected, nexts)

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
	nexts = NextStages(wf, wfr)
	assert.Equal(t, expected, nexts)

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
	nexts = NextStages(wf, wfr)
	assert.Equal(t, expected, nexts)
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

func TestResolveRefStringValue(t *testing.T) {
	jsonData := "{\"user\": {\"name\":\"cyclone\", \"languages\": [\"java\", \"go\"]}}"
	client := fake.NewSimpleClientset()
	client.PrependReactor("get", "secrets", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret",
				Namespace: "ns",
			},
			Data: map[string][]byte{
				"key":  []byte("key1"),
				"json": []byte(jsonData),
			},
		}, nil
	})

	assert := assert.New(t)
	v, err := ResolveRefStringValue("value1", client)
	assert.Nil(err)
	assert.Equal("value1", v)
	v, err = ResolveRefStringValue("111", client)
	assert.Nil(err)
	assert.Equal("111", v)
	v, err = ResolveRefStringValue("ns.secret", client)
	assert.Nil(err)
	assert.Equal("ns.secret", v)
	v, err = ResolveRefStringValue("$.ns.secret/data.key", client)
	assert.Nil(err)
	assert.Equal("key1", v)
	_, err = ResolveRefStringValue("$.ns.secret/data.non-exist-key", client)
	assert.Error(err)
	_, err = ResolveRefStringValue("$.ns.secret/data.json/non-exist-key", client)
	assert.Error(err)
	v, err = ResolveRefStringValue("$.ns.secret/data.json/user.name", client)
	assert.Nil(err)
	assert.Equal("cyclone", v)
}

func TestIsWorkflowRunTerminated(t *testing.T) {
	wfr = &v1alpha1.WorkflowRun{
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
