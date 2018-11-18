package workflowrun

import (
	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
)

// NextStages determine next stages that can be started to execute. It returns
// stages that are not started yet but have all depended stages finished.
func NextStages(wf *v1alpha1.Workflow, wfr *v1alpha1.WorkflowRun) []string {
	var nextStages []string
	for _, stage := range wf.Spec.Stages {
		// If this stage already have status set, it means it's already been started, skip it.
		_, ok := wfr.Status.Stages[stage.Name]
		if ok {
			continue
		}

		// All depended stages must have been successfully finished, otherwise this
		// stage would be skipped.
		safeToRun := true
		for _, d := range stage.Depends {
			status, ok := wfr.Status.Stages[d]
			if !(ok && status.Status.Status == v1alpha1.StatusCompleted) {
				safeToRun = false
				break
			}
		}

		if safeToRun {
			nextStages = append(nextStages, stage.Name)
		}
	}

	return nextStages
}
