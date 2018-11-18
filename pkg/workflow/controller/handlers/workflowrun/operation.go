package workflowrun

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var lock sync.Mutex

type Operator struct {
	client clientset.Interface
}

func NewOperator(client clientset.Interface) *Operator {
	return &Operator{
		client: client,
	}
}

// TODO(ChenDe): Need a more decent way to handle this.
func combineStatus(origin, modifed *v1alpha1.WorkflowRun) *v1alpha1.WorkflowRun {
	combined := origin.DeepCopy()
	if combined.Status.Overall.Status != v1alpha1.StatusError && combined.Status.Overall.Status != v1alpha1.StatusCompleted {
		combined.Status.Overall = modifed.Status.Overall
	}
	if combined.Status.Stages == nil {
		combined.Status.Stages = modifed.Status.Stages
	} else if modifed.Status.Stages == nil {
		return combined
	}

	modifedStages := modifed.Status.Stages
	for stage, status := range modifedStages {
		s, ok := combined.Status.Stages[stage]
		if !ok {
			combined.Status.Stages[stage] = status
			continue
		}
		if s.Pod == nil {
			combined.Status.Stages[stage].Pod = status.Pod
		}

		if s.Status.Status != v1alpha1.StatusError && s.Status.Status != v1alpha1.StatusCompleted {
			combined.Status.Stages[stage].Status = status.Status
		}

		combined.Status.Stages[stage].Outputs = status.Outputs
	}

	return combined
}

func (p *Operator) UpdateStatus(wfr *v1alpha1.WorkflowRun) error {
	lock.Lock()
	defer lock.Unlock()

	latest, err := p.client.CycloneV1alpha1().WorkflowRuns(wfr.Namespace).Get(wfr.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	combined := combineStatus(latest, wfr)

	log.WithField("workflowrun", wfr.Name).Info("Update WorkflowRun status")
	if !reflect.DeepEqual(staticStatus(&latest.Status), staticStatus(&combined.Status)) {
		_, err = p.client.CycloneV1alpha1().WorkflowRuns(wfr.Namespace).Update(combined)
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

func (p *Operator) OverallStatus(wfr *v1alpha1.WorkflowRun, wf *v1alpha1.Workflow) string {
	if wfr.Status.Stages == nil || len(wfr.Status.Stages) == 0 {
		return v1alpha1.StatusRunning
	}

	var running, waiting, err bool
	for _, stage := range wfr.Status.Stages {
		if stage.Status.Status == "" {
			log.Error("Empty status")
			return v1alpha1.StatusRunning
		}

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

	next := len(NextStages(wf, wfr))
	if next > 0 {
		return v1alpha1.StatusRunning
	}
	return v1alpha1.StatusCompleted
}

func (p *Operator) SetStagePodInfo(wfr *v1alpha1.WorkflowRun, stage string, podInfo *v1alpha1.PodInfo) {
	_, ok := wfr.Status.Stages[stage]
	if !ok {
		wfr.Status.Stages[stage] = &v1alpha1.StageStatus{}
	}
	wfr.Status.Stages[stage].Pod = podInfo
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
		p.SetStageStatus(wfr, stage, v1alpha1.Status{
			Status:             v1alpha1.StatusRunning,
			Reason:             "StageInitialized",
			LastTransitionTime: metav1.Time{time.Now()},
		})
	}
	wfr.Status.Overall.Status = p.OverallStatus(wfr, len(wf.Spec.Stages))
	err = p.UpdateStatus(wfr)
	if err != nil {
		log.WithField("WorkflowRun", wfr.Name).Error("Update status error: ", err)
		return err
	}

	// Return if no stages need to run.
	if len(nextStages) == 0 {
		return nil
	}

	// Create pod to run stages.
	for _, stage := range nextStages {
		log.WithField("stage", stage).Info("Start to run stage")

		// Generate pod for this stage.
		pod, err := NewPodMaker(p.client, wf, wfr, stage).MakePod()
		if err != nil {
			log.WithField("WorkflowRun", wfr.Name).WithField("Stage", stage).Error("Create pod manifest for stage error: ", err)
			p.SetStageStatus(wfr, stage, v1alpha1.Status{
				Status:             v1alpha1.StatusError,
				Reason:             "GeneratePodError",
				LastTransitionTime: metav1.Time{time.Now()},
				Message:            fmt.Sprintf("Failed to generate pod: %v", err),
			})
			continue
		}
		log.WithField("Stage", stage).Debug("Pod manifest created")

		// Create the generated pod.
		pod, err = p.client.CoreV1().Pods(wfr.Namespace).Create(pod)
		if err != nil {
			log.WithField("WorkflowRun", wfr.Name).WithField("Stage", stage).Error("Create pod for stage error: ", err)
			p.SetStageStatus(wfr, stage, v1alpha1.Status{
				Status:             v1alpha1.StatusError,
				Reason:             "CreatePodError",
				LastTransitionTime: metav1.Time{time.Now()},
				Message:            fmt.Sprintf("Failed to create pod: %v", err),
			})
			continue
		}

		p.SetStageStatus(wfr, stage, v1alpha1.Status{
			Status:             v1alpha1.StatusRunning,
			LastTransitionTime: metav1.Time{time.Now()},
			Reason:             "StageInitialized",
		})

		p.SetStagePodInfo(wfr, stage, &v1alpha1.PodInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		})
	}

	wfr.Status.Overall.Status = p.OverallStatus(wfr, len(wf.Spec.Stages))
	err = p.UpdateStatus(wfr)
	if err != nil {
		log.WithField("WorkflowRun", wfr.Name).Error("Update status error: ", err)
		return err
	}

	return nil
}
