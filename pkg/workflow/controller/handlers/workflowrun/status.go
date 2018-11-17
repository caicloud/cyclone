package workflowrun

import (
	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
)

func SetStageStatus(wfr *v1alpha1.WorkflowRun, stage string, status v1alpha1.Status) {
	_, ok := wfr.Status.Stages[stage]
	if !ok {
		wfr.Status.Stages[stage] = &v1alpha1.StageStatus{}
	}
	wfr.Status.Stages[stage].Status = status
}

func OverallStatus(wfr *v1alpha1.WorkflowRun) string {
	var running, waiting, err bool
	for _, stage := range wfr.Status.Stages {
		switch stage.Status.Status {
		case v1alpha1.StatusRunning:
			running = true
		case v1alpha1.StatusWaiting:
			waiting = true
		case v1alpha1.StatusError:
			err = true
		}
	}

	if running {
		return v1alpha1.StatusRunning
	}

	if waiting {
		return v1alpha1.StatusWaiting
	}

	if err {
		return v1alpha1.StatusError
	}

	return v1alpha1.StatusCompleted
}
