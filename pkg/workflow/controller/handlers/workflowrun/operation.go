package workflowrun

import (
	"fmt"
	"reflect"
	"time"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Operator struct {
	client clientset.Interface
}

func NewOperator(client clientset.Interface) *Operator {
	return &Operator{
		client: client,
	}
}

func (p *Operator) UpdateStatus(wfr *v1alpha1.WorkflowRun) error {
	latest, err := p.client.CycloneV1alpha1().WorkflowRuns(wfr.Namespace).Get(wfr.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(staticStatus(&latest.Status), staticStatus(&wfr.Status)) {
		latest.Status = wfr.Status
		_, err := p.client.CycloneV1alpha1().WorkflowRuns(wfr.Namespace).Update(latest)
		return err
	}

	return nil
}

func (p *Operator) SetStageStatus(wfr *v1alpha1.WorkflowRun, stage string, status v1alpha1.Status) {
	if wfr.Status.Stages == nil {
		wfr.Status.Stages = make(map[string]*v1alpha1.StageStatus)
	}
	_, ok := wfr.Status.Stages[stage]
	if !ok {
		wfr.Status.Stages[stage] = &v1alpha1.StageStatus{}
	}
	wfr.Status.Stages[stage].Status = status
}

// staticStatus masks timestamp in status, safe for comparision of status.
func staticStatus(status *v1alpha1.WorkflowRunStatus) *v1alpha1.WorkflowRunStatus {
	t := metav1.Time{Time: time.Unix(0, 0)}
	copy := status.DeepCopy()
	copy.Overall.LastTransitionTime = t
	for k, _ := range status.Stages {
		copy.Stages[k].Status.LastTransitionTime = t
	}
	return copy
}

func (p *Operator) Reconcile(wfr *v1alpha1.WorkflowRun) error {
	if wfr.Status.Stages == nil {
		wfr.Status.Stages = make(map[string]*v1alpha1.StageStatus)
	}

	// Get the Workflow that this WorkflowRun referenced.
	wfRef := wfr.Spec.WorkflowRef
	wf, err := p.client.CycloneV1alpha1().Workflows(wfr.Namespace).Get(wfRef.Name, metav1.GetOptions{})
	if err != nil {
		log.WithField("workflow", wfRef.Name).Error("Get Workflow error: ", err)
		return err
	}

	nextStages := NextStages(wf, wfr)
	if len(nextStages) == 0 {
		log.WithField("workflowrun", wfr.Name).Info("No next stages to run")
	} else {
		log.WithField("stages", nextStages).Info("Next stages to run")
	}

	for _, stage := range nextStages {
		SetStageStatus(wfr, stage, v1alpha1.Status{
			Status:             v1alpha1.StatusRunning,
			Reason:             "StageInitialized",
			LastTransitionTime: metav1.Time{time.Now()},
		})
	}
	wfr.Status.Overall.Status = OverallStatus(wfr)
	err = p.UpdateStatus(wfr)
	if err != nil {
		log.WithField("WorkflowRun", wfr.Name).Error("Update status error: ", err)
		return err
	}

	// Return if no stages need to run.
	if len(nextStages) <= 0 {
		return nil
	}

	// Create pod to run stages.
	for _, stage := range nextStages {
		log.WithField("stage", stage).Info("Start to run stage")

		// Generate pod for this stage.
		pod, err := NewPodMaker(p.client, wf, wfr, stage).MakePod()
		if err != nil {
			log.WithField("WorkflowRun", wfr.Name).WithField("Stage", stage).Error("Create pod manifest for stage error: ", err)
			SetStageStatus(wfr, stage, v1alpha1.Status{
				Status:             v1alpha1.StatusError,
				Reason:             "GeneratePodError",
				LastTransitionTime: metav1.Time{time.Now()},
				Message:            fmt.Sprintf("Failed to generate pod: %v", err),
			})
			continue
		}
		log.WithField("Stage", stage).Debug("Pod manifest created")

		// Create the generated pod.
		_, err = p.client.CoreV1().Pods(wfr.Namespace).Create(pod)
		if err != nil {
			log.WithField("WorkflowRun", wfr.Name).WithField("Stage", stage).Error("Create pod for stage error: ", err)
			SetStageStatus(wfr, stage, v1alpha1.Status{
				Status:             v1alpha1.StatusError,
				Reason:             "CreatePodError",
				LastTransitionTime: metav1.Time{time.Now()},
				Message:            fmt.Sprintf("Failed to create pod: %v", err),
			})
			continue
		}

		SetStageStatus(wfr, stage, v1alpha1.Status{
			Status:             v1alpha1.StatusRunning,
			LastTransitionTime: metav1.Time{time.Now()},
			Reason:             "StageInitialized",
		})
	}

	wfr.Status.Overall.Status = OverallStatus(wfr)
	err = p.UpdateStatus(wfr)
	if err != nil {
		log.WithField("WorkflowRun", wfr.Name).Error("Update status error: ", err)
		return err
	}

	return nil
}
