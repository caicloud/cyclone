package pod

import (
	"fmt"
	"time"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers/workflowrun"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Operator struct {
	client      clientset.Interface
	workflowRun string
	stage       string
	pod         *corev1.Pod
	wfrOperator *workflowrun.Operator
}

func NewOperator(client clientset.Interface, pod *corev1.Pod) (*Operator, error) {
	annotations := pod.Annotations
	wfr, ok := annotations[workflow.WorkflowRunAnnotationName]
	if !ok {
		return nil, fmt.Errorf("invalid workflow pod, without annotation %s", workflow.WorkflowRunAnnotationName)
	}
	stage, ok := annotations[workflow.StageAnnotationName]
	if !ok {
		return nil, fmt.Errorf("invalid workflow pod, without annotation %s", workflow.StageAnnotationName)
	}

	return &Operator{
		client:      client,
		workflowRun: wfr,
		stage:       stage,
		pod:         pod,
		wfrOperator: workflowrun.NewOperator(client),
	}, nil
}

// OnDelete handles the situation when a stage pod gotten delete. It updates
// corresponding WorkflowRun's status.
func (p *Operator) OnDelete() error {
	origin, err := p.client.CycloneV1alpha1().WorkflowRuns(p.pod.Namespace).Get(p.workflowRun, metav1.GetOptions{})
	if err != nil {
		log.WithField("name", p.workflowRun).Error("Get WorkflowRun error: ", err)
		return err
	}

	wfr := origin.DeepCopy()
	status, ok := wfr.Status.Stages[p.stage]
	if !ok || status.Status.Status == v1alpha1.StatusRunning {
		p.wfrOperator.SetStageStatus(wfr, p.stage, v1alpha1.Status{
			Status:             "Error",
			LastTransitionTime: metav1.Time{time.Now()},
			Reason:             "PodDeleted",
		})
	}

	return p.wfrOperator.UpdateStatus(wfr)
}

func (p *Operator) OnUpdated() error {
	origin, err := p.client.CycloneV1alpha1().WorkflowRuns(p.pod.Namespace).Get(p.workflowRun, metav1.GetOptions{})
	if err != nil {
		log.WithField("name", p.workflowRun).Error("Get WorkflowRun error: ", err)
		return err
	}

	wfr := origin.DeepCopy()
	status, ok := wfr.Status.Stages[p.stage]

	switch p.pod.Status.Phase {
	case corev1.PodFailed:
		if !ok || status.Status.Status != v1alpha1.StatusError {
			log.WithField("workflowrun", wfr.Name).
				WithField("stage", p.stage).
				WithField("status", v1alpha1.StatusError).
				Info("To update stage status")
			p.wfrOperator.SetStageStatus(wfr, p.stage, v1alpha1.Status{
				Status:             v1alpha1.StatusError,
				LastTransitionTime: metav1.Time{time.Now()},
				Reason:             "PodFailed",
			})
		}
	case corev1.PodSucceeded:
		if !ok || status.Status.Status != v1alpha1.StatusCompleted {
			log.WithField("workflowrun", wfr.Name).
				WithField("stage", p.stage).
				WithField("status", v1alpha1.StatusCompleted).
				Info("To update stage status")
			p.wfrOperator.SetStageStatus(wfr, p.stage, v1alpha1.Status{
				Status:             v1alpha1.StatusCompleted,
				LastTransitionTime: metav1.Time{time.Now()},
				Reason:             "PodSucceed",
			})
		}
	default:
		p.DetermineStatus(wfr)
	}

	return p.wfrOperator.UpdateStatus(wfr)
}

// DetermineStatus determines status of a stage and update WorkflowRun status accordingly.
// Because coordinator container is the last container running in the pod (it performs collect
// logs, artifacts, notify resource resolver to push resource), when the coordinator container
// have been finished (no matter Succeed or Failed), we need to update stage status, and take
// necessary actions to stop the pod.
func (p *Operator) DetermineStatus(wfr *v1alpha1.WorkflowRun) {
	// If there are containers that haven't report status, no need to judge pod status.
	if len(p.pod.Status.ContainerStatuses) != len(p.pod.Spec.Containers) {
		return
	}

	// Check coordinator container's status, if it's terminated, we regard the pod completed.
	anyError := false
	for _, containerStatus := range p.pod.Status.ContainerStatuses {
		if containerStatus.Name == workflow.CoordinatorContainerName {
			if containerStatus.State.Terminated == nil {
				log.WithField("container", containerStatus.Name).Debug("Coordinator not terminated")
				return
			}
			if containerStatus.State.Terminated.ExitCode != 0 {
				anyError = true
			}
		}
	}

	// Now the workload containers and coordinator container have all been finished. We then:
	// - Update the stage status in WorkflowRun based on coordinator's exit code.
	// - TODO(ChenDe): Delete pod

	if anyError {
		log.WithField("workflowrun", wfr.Name).
			WithField("stage", p.stage).
			WithField("status", v1alpha1.StatusError).
			Info("To update stage status")
		p.wfrOperator.SetStageStatus(wfr, p.stage, v1alpha1.Status{
			Status:             v1alpha1.StatusError,
			LastTransitionTime: metav1.Time{time.Now()},
			Reason:             "CoordinatorError",
			Message:            "Coordinator exit with error",
		})
	} else {
		log.WithField("workflowrun", wfr.Name).
			WithField("stage", p.stage).
			WithField("status", v1alpha1.StatusCompleted).
			Info("To update stage status")
		p.wfrOperator.SetStageStatus(wfr, p.stage, v1alpha1.Status{
			Status:             v1alpha1.StatusCompleted,
			LastTransitionTime: metav1.Time{time.Now()},
			Reason:             "CoordinatorCompleted",
			Message:            "Coordinator completed",
		})
	}
}
