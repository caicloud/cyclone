package workflowrun

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/caicloud/cyclone/pkg/workflow/controller"
)

func TestNewParallelismController(t *testing.T) {
	assert.NotNil(t, NewParallelismController(nil))
	assert.NotNil(t, NewParallelismController(&controller.ParallelismConfig{}))
	assert.NotNil(t, NewParallelismController(&controller.ParallelismConfig{
		Overall: controller.ParallelismConstraint{
			MaxParallel:  2,
			MaxQueueSize: 5,
		},
	}))
	assert.NotNil(t, NewParallelismController(&controller.ParallelismConfig{
		SingleWorkflow: controller.ParallelismConstraint{
			MaxParallel:  2,
			MaxQueueSize: 5,
		},
	}))
}

func TestAttemptNew(t *testing.T) {
	pc := NewParallelismController(&controller.ParallelismConfig{
		Overall: controller.ParallelismConstraint{
			MaxParallel:  2,
			MaxQueueSize: 2,
		},
	})

	assert.Equal(t, AttemptActionStart, pc.AttemptNew("ns1", "wf1", "wfr1"))
	assert.Equal(t, AttemptActionStart, pc.AttemptNew("ns1", "wf1", "wfr2"))
	assert.Equal(t, AttemptActionQueued, pc.AttemptNew("ns1", "wf1", "wfr3"))
	assert.Equal(t, AttemptActionQueued, pc.AttemptNew("ns1", "wf1", "wfr4"))
	assert.Equal(t, AttemptActionFailed, pc.AttemptNew("ns1", "wf1", "wfr5"))

	pc = NewParallelismController(&controller.ParallelismConfig{
		SingleWorkflow: controller.ParallelismConstraint{
			MaxParallel:  1,
			MaxQueueSize: 2,
		},
	})

	assert.Equal(t, AttemptActionStart, pc.AttemptNew("ns1", "wf1", "wfr1"))
	assert.Equal(t, AttemptActionQueued, pc.AttemptNew("ns1", "wf1", "wfr2"))
	assert.Equal(t, AttemptActionQueued, pc.AttemptNew("ns1", "wf1", "wfr3"))
	assert.Equal(t, AttemptActionFailed, pc.AttemptNew("ns1", "wf1", "wfr4"))

	pc = NewParallelismController(&controller.ParallelismConfig{
		Overall: controller.ParallelismConstraint{
			MaxParallel:  3,
			MaxQueueSize: 3,
		},
		SingleWorkflow: controller.ParallelismConstraint{
			MaxParallel:  2,
			MaxQueueSize: 2,
		},
	})

	assert.Equal(t, AttemptActionStart, pc.AttemptNew("ns1", "wf1", "wfr1"))
	assert.Equal(t, AttemptActionStart, pc.AttemptNew("ns1", "wf1", "wfr2"))
	assert.Equal(t, AttemptActionQueued, pc.AttemptNew("ns1", "wf1", "wfr3"))
	assert.Equal(t, AttemptActionQueued, pc.AttemptNew("ns1", "wf1", "wfr4"))
	assert.Equal(t, AttemptActionFailed, pc.AttemptNew("ns1", "wf1", "wfr5"))

	assert.Equal(t, AttemptActionStart, pc.AttemptNew("ns1", "wf2", "wfr1"))
	assert.Equal(t, AttemptActionQueued, pc.AttemptNew("ns1", "wf2", "wfr2"))
	assert.Equal(t, AttemptActionFailed, pc.AttemptNew("ns1", "wf2", "wfr3"))
	assert.Equal(t, AttemptActionFailed, pc.AttemptNew("ns1", "wf2", "wfr4"))
	assert.Equal(t, AttemptActionFailed, pc.AttemptNew("ns1", "wf2", "wfr5"))
}

func TestMarkFinished(t *testing.T) {
	pc := NewParallelismController(&controller.ParallelismConfig{
		Overall: controller.ParallelismConstraint{
			MaxParallel:  2,
			MaxQueueSize: 2,
		},
	})

	assert.Equal(t, AttemptActionStart, pc.AttemptNew("ns1", "wf1", "wfr1"))
	assert.Equal(t, AttemptActionStart, pc.AttemptNew("ns1", "wf1", "wfr2"))
	assert.Equal(t, AttemptActionQueued, pc.AttemptNew("ns1", "wf1", "wfr3"))
	assert.Equal(t, AttemptActionQueued, pc.AttemptNew("ns1", "wf1", "wfr4"))
	assert.Equal(t, AttemptActionFailed, pc.AttemptNew("ns1", "wf1", "wfr5"))
	pc.MarkFinished("ns1", "wf1", "wfr1")
	assert.Equal(t, AttemptActionStart, pc.AttemptNew("ns1", "wf1", "wfr3"))
	assert.Equal(t, AttemptActionQueued, pc.AttemptNew("ns1", "wf1", "wfr4"))
	assert.Equal(t, AttemptActionQueued, pc.AttemptNew("ns1", "wf1", "wfr4"))
	assert.Equal(t, AttemptActionQueued, pc.AttemptNew("ns1", "wf1", "wfr5"))
	assert.Equal(t, AttemptActionFailed, pc.AttemptNew("ns1", "wf1", "wfr6"))
	pc.MarkFinished("ns1", "wf1", "wfr2")
	assert.Equal(t, AttemptActionStart, pc.AttemptNew("ns1", "wf1", "wfr4"))
	assert.Equal(t, AttemptActionQueued, pc.AttemptNew("ns1", "wf1", "wfr5"))
}
