package pod

import (
	"github.com/caicloud/cyclone/pkg/k8s/clientset"

	"fmt"
	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/workflow"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers/workflowrun"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
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
	wfr, ok := annotations[workflow.WorkflowrunAnnotationName]
	if !ok {
		return nil, fmt.Errorf("invalid workflow pod, without annotation %s", workflow.WorkflowrunAnnotationName)
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
			p.wfrOperator.SetStageStatus(wfr, p.stage, v1alpha1.Status{
				Status:             v1alpha1.StatusError,
				LastTransitionTime: metav1.Time{time.Now()},
				Reason:             "PodFailed",
			})
		}
	case corev1.PodSucceeded:
		if !ok || status.Status.Status != v1alpha1.StatusCompleted {
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
// When the main working containers have been finished (no matter Succeed or Failed), we need
// to update stage status, and take necessary actions to stop the pod.
func (p *Operator) DetermineStatus(wfr *v1alpha1.WorkflowRun) {
	// If there are containers that haven't report status, no need to judge pod status.
	if len(p.pod.Status.ContainerStatuses) != len(p.pod.Spec.Containers) {
		return
	}

	// If any workload containers (non-sidecars) not terminated, we regard the pod not completed.
	anyError := false
	for _, containerStatus := range p.pod.Status.ContainerStatuses {
		if !strings.HasPrefix(containerStatus.Name, workflow.SidecarContainerPrefix) {
			if containerStatus.State.Terminated == nil {
				log.WithField("container", containerStatus.Name).Debug("Container not terminated")
				return
			}
			if containerStatus.State.Terminated.ExitCode != 0 {
				anyError = true
			}
		}
	}

	// Now the workload containers have all been finished. We then:
	// - Update the stage status in WorkflowRun
	// - TODO(ChenDe): Notify sidecar to process artifacts or output resoruces
	// - TODO(ChenDe): Stop containers or delete pod

	if anyError {
		log.WithField("stage", p.stage).Debug("Pod status judged to be error")
		p.wfrOperator.SetStageStatus(wfr, p.stage, v1alpha1.Status{
			Status:             v1alpha1.StatusError,
			LastTransitionTime: metav1.Time{time.Now()},
			Reason:             "PayloadContainersError",
			Message:            "Some workload containers failed",
		})
	} else {
		log.WithField("stage", p.stage).Debug("Pod status judged to be completed")
		p.wfrOperator.SetStageStatus(wfr, p.stage, v1alpha1.Status{
			Status:             v1alpha1.StatusCompleted,
			LastTransitionTime: metav1.Time{time.Now()},
			Reason:             "PayloadContainersCompleted",
			Message:            "All workload containers succeeded",
		})
	}
}
