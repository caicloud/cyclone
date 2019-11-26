package workflowrun

import (
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/caicloud/cyclone/pkg/workflow/controller"
)

type AttemptAction string

const (
	AttemptActionStart  AttemptAction = "Start"
	AttemptActionQueued AttemptAction = "Queued"
	AttemptActionFailed AttemptAction = "Failed"
)

// ParallelismController is interface to manage parallelism of WorkflowRun executions
type ParallelismController interface {
	// AttemptNew tries to run a new WorkflowRun, and returning the corresponding action.
	// AttemptActionStart -- The WorkflowRun can start to run
	// AttemptActionQueued -- The WorkflowRun is queued
	// AttemptActionFailed -- The WorkflowRun will fail directly due to queue full
	AttemptNew(ns, wf, wfr string) AttemptAction
	// MarkFinished mark a WorkflowRun execution finished
	MarkFinished(ns, wf, wfr string)
}

type wfStatus struct {
	// WorkflowRuns that are waiting for the Workflow
	waitingWfrMap map[string]struct{}
	// WorkflowRuns that are running
	runningWfrMap map[string]struct{}
}

type overallStatus struct {
	// How many WorkflowRun are running in total
	total int64
	// How many WorkflowRun are waiting in total
	waiting int64
}

type parallelismController struct {
	config  *controller.ParallelismConfig
	wfMap   map[string]*wfStatus
	overall overallStatus
	lock    *sync.Mutex
}

func NewParallelismController(parallelismConfig *controller.ParallelismConfig) ParallelismController {
	return &parallelismController{
		config: parallelismConfig,
		wfMap:  make(map[string]*wfStatus),
		lock:   &sync.Mutex{},
	}
}

// AttemptNew tries to run a new WorkflowRun, and returning the corresponding action.
// AttemptActionStart -- The WorkflowRun can start to run
// AttemptActionQueued -- The WorkflowRun is queued
// AttemptActionFailed -- The WorkflowRun will fail directly due to queue full
func (c *parallelismController) AttemptNew(ns, wf, wfr string) AttemptAction {
	// If no parallelism constraint configured, always allow WorkflowRun to execute
	if c.config == nil {
		return AttemptActionStart
	}
	log.Infof("Attempt to run '%s'", wfr)

	c.lock.Lock()
	defer c.lock.Unlock()

	// If the WorkflowRun already in running, return action directly
	var alreadyQueued bool
	if m, ok := c.wfMap[wf]; ok {
		if _, ook := m.runningWfrMap[wfr]; ook {
			return AttemptActionStart
		}
		if _, ook := m.waitingWfrMap[wfr]; ook {
			alreadyQueued = true
		}
	}

	// Check the overall constraints.
	// Configured value 0 or negative indicates no constraints
	var overallVote = AttemptActionStart
	if c.config.Overall.MaxParallel > 0 {
		if c.overall.total >= c.config.Overall.MaxParallel {
			// Configured value 0 or negative indicates no constraints
			if c.config.Overall.MaxQueueSize > 0 {
				if c.overall.waiting >= c.config.Overall.MaxQueueSize {
					overallVote = AttemptActionFailed
				} else {
					overallVote = AttemptActionQueued
				}
			}
		}
	}

	// Check the Workflow level constraints.
	// Configured value 0 or negative indicates no constraints
	var workflowVote = AttemptActionStart
	if c.config.SingleWorkflow.MaxParallel > 0 {
		if m, ok := c.wfMap[wf]; ok {
			if int64(len(m.runningWfrMap)) >= c.config.SingleWorkflow.MaxParallel {
				// Configured value 0 or negative indicates no constraints
				if c.config.SingleWorkflow.MaxQueueSize > 0 {
					if int64(len(m.waitingWfrMap)) >= c.config.SingleWorkflow.MaxQueueSize {
						workflowVote = AttemptActionFailed
					} else {
						workflowVote = AttemptActionQueued
					}
				}
			}
		}
	}

	if overallVote == AttemptActionFailed || workflowVote == AttemptActionFailed {
		if alreadyQueued {
			return AttemptActionQueued
		}
		return AttemptActionFailed
	}

	if overallVote == AttemptActionQueued || workflowVote == AttemptActionQueued {
		// Register the WorkflowRun in waiting queue
		if _, ok := c.wfMap[wf]; !ok {
			c.wfMap[wf] = &wfStatus{
				runningWfrMap: make(map[string]struct{}),
				waitingWfrMap: make(map[string]struct{}),
			}
		}

		if _, ook := c.wfMap[wf].waitingWfrMap[wfr]; !ook {
			c.overall.waiting++
			c.wfMap[wf].waitingWfrMap[wfr] = struct{}{}
		}

		return AttemptActionQueued
	}

	// Register the WorkflowRun to be ready for running
	if _, ok := c.wfMap[wf]; !ok {
		c.wfMap[wf] = &wfStatus{
			runningWfrMap: make(map[string]struct{}),
			waitingWfrMap: make(map[string]struct{}),
		}
	}
	if _, ook := c.wfMap[wf].runningWfrMap[wfr]; !ook {
		c.overall.total++
		c.wfMap[wf].runningWfrMap[wfr] = struct{}{}
	}

	// If the WorkflowRun is previously in waiting queue, remove it from the queue
	if _, ok := c.wfMap[wf].waitingWfrMap[wfr]; ok {
		delete(c.wfMap[wf].waitingWfrMap, wfr)
		c.overall.waiting--
	}

	return AttemptActionStart
}

// MarkFinished mark a WorkflowRun execution finished
func (c *parallelismController) MarkFinished(ns, wf, wfr string) {
	// If no parallelism constraint configured, no action to take
	if c.config == nil {
		return
	}
	log.Infof("To mark '%s' finished", wfr)

	c.lock.Lock()
	defer c.lock.Unlock()

	if m, ok := c.wfMap[wf]; ok {
		if _, ook := m.runningWfrMap[wfr]; ook {
			c.overall.total--
			delete(m.runningWfrMap, wfr)
			log.Infof("Delete %s from running map, %d remained", wfr, len(m.runningWfrMap))
		}
	}
}
