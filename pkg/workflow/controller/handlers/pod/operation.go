package pod

import (
	"fmt"
	"time"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/workflow/workflowrun"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Operator ...
type Operator struct {
	client        clientset.Interface
	workflowRun   string
	stage         string
	metaNamespace string
	pod           *corev1.Pod
}

// NewOperator ...
func NewOperator(client clientset.Interface, pod *corev1.Pod) (*Operator, error) {
	annotations := pod.Annotations
	wfr, ok := annotations[common.WorkflowRunAnnotationName]
	if !ok {
		return nil, fmt.Errorf("invalid workflow pod, without annotation %s", common.WorkflowRunAnnotationName)
	}
	stage, ok := annotations[common.StageAnnotationName]
	if !ok {
		return nil, fmt.Errorf("invalid workflow pod, without annotation %s", common.StageAnnotationName)
	}
	metaNamespace, ok := annotations[common.MetaNamespaceAnnotationName]
	if !ok {
		return nil, fmt.Errorf("invalid workflow pod, without annotation %s", common.MetaNamespaceAnnotationName)
	}

	return &Operator{
		client:        client,
		workflowRun:   wfr,
		stage:         stage,
		metaNamespace: metaNamespace,
		pod:           pod,
	}, nil
}

// OnDelete handles the situation when a stage pod gotten delete. It updates
// corresponding WorkflowRun's status.
func (p *Operator) OnDelete() error {
	origin, err := p.client.CycloneV1alpha1().WorkflowRuns(p.metaNamespace).Get(p.workflowRun, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			log.WithField("name", p.workflowRun).Error("Get WorkflowRun error: ", err)
			return err
		}
		return nil
	}

	wfr := origin.DeepCopy()
	operator, err := workflowrun.NewOperator(p.client, wfr, origin.Namespace)
	if err != nil {
		return err
	}
	status, ok := wfr.Status.Stages[p.stage]
	if !ok || status.Status.Phase == v1alpha1.StatusRunning {
		operator.UpdateStageStatus(p.stage, &v1alpha1.Status{
			Phase:              v1alpha1.StatusFailed,
			LastTransitionTime: metav1.Time{Time: time.Now()},
			Reason:             "PodDeleted",
		})
	}

	return operator.Update()
}

// OnUpdated ...
func (p *Operator) OnUpdated() error {
	origin, err := p.client.CycloneV1alpha1().WorkflowRuns(p.metaNamespace).Get(p.workflowRun, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.WithField("wfr", p.workflowRun).WithField("ns", p.metaNamespace).Warn("wfr not found")
			// Delete the pod if WorkflowRun not exists any more, there is possible that the pod been deleted elsewhere on WorkflowRun deletion,
			// so if we delete pod failed here due to not found, just ignore it.
			if err := p.client.CoreV1().Pods(p.pod.Namespace).Delete(p.pod.Name, &metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
				log.WithField("ns", p.pod.Namespace).WithField("pod", p.pod.Name).Warn("Delete orphan pod error: ", err)
			} else {
				log.WithField("ns", p.pod.Namespace).WithField("pod", p.pod.Name).Info("Orphan pod deleted")
			}
			return nil
		}
		log.WithField("name", p.workflowRun).Error("Get WorkflowRun error: ", err)
		return err
	}

	// If the WorkflowRun has already been in terminated state, skip it.
	if origin.Status.Overall.Phase == v1alpha1.StatusSucceeded ||
		origin.Status.Overall.Phase == v1alpha1.StatusFailed {
		return nil
	}

	wfr := origin.DeepCopy()
	wfrOperator, err := workflowrun.NewOperator(p.client, wfr, origin.Namespace)
	if err != nil {
		return err
	}

	status, ok := wfr.Status.Stages[p.stage]

	switch p.pod.Status.Phase {
	case corev1.PodFailed:
		if !ok || status.Status.Phase != v1alpha1.StatusFailed {
			log.WithField("wfr", wfr.Name).
				WithField("stg", p.stage).
				WithField("status", v1alpha1.StatusFailed).
				Info("To update stage status")
			wfrOperator.UpdateStageStatus(p.stage, &v1alpha1.Status{
				Phase:              v1alpha1.StatusFailed,
				LastTransitionTime: metav1.Time{Time: time.Now()},
				Reason:             "PodFailed",
			})
		}
	case corev1.PodSucceeded:
		if !ok || status.Status.Phase != v1alpha1.StatusSucceeded {
			log.WithField("wfr", wfr.Name).
				WithField("stage", p.stage).
				WithField("status", v1alpha1.StatusSucceeded).
				Info("To update stage status")
			wfrOperator.UpdateStageStatus(p.stage, &v1alpha1.Status{
				Phase:              v1alpha1.StatusSucceeded,
				LastTransitionTime: metav1.Time{Time: time.Now()},
				Reason:             "PodSucceed",
			})
		}
	default:
		p.DetermineStatus(wfrOperator)
	}

	return wfrOperator.Update()
}

// DetermineStatus determines status of a stage and update WorkflowRun status accordingly.
// Because coordinator container is the last container running in the pod (it performs collect
// logs, artifacts, notify resource resolver to push resource), when the coordinator container
// have been finished (no matter Succeed or Failed), we need to update stage status, and take
// necessary actions to stop the pod.
func (p *Operator) DetermineStatus(wfrOperator workflowrun.Operator) {
	// If there are containers that haven't report status, no need to judge pod status.
	if len(p.pod.Status.ContainerStatuses) != len(p.pod.Spec.Containers) {
		return
	}

	// Check coordinator container's status, if it's terminated, we regard the pod completed.
	var terminatedCoordinatorState *corev1.ContainerStateTerminated
	for _, containerStatus := range p.pod.Status.ContainerStatuses {
		if containerStatus.Name == common.CoordinatorSidecarName {
			if containerStatus.State.Terminated == nil {
				log.WithField("container", containerStatus.Name).Debug("Coordinator not terminated")
				return
			}

			terminatedCoordinatorState = containerStatus.State.Terminated

			// There is only one coordinator container in each pod.
			break
		}
	}

	// Now the workload containers and coordinator container have all been finished. We then:
	// - Update the stage status in WorkflowRun based on coordinator's exit code.
	if terminatedCoordinatorState.ExitCode != 0 {
		log.WithField("wfr", wfrOperator.GetWorkflowRun().Name).
			WithField("stg", p.stage).
			WithField("status", v1alpha1.StatusFailed).
			Info("To update stage status")
		wfrOperator.UpdateStageStatus(p.stage, &v1alpha1.Status{
			Phase:              v1alpha1.StatusFailed,
			LastTransitionTime: metav1.Time{Time: time.Now()},
			Reason:             terminatedCoordinatorState.Reason,
			Message:            terminatedCoordinatorState.Message,
		})
	} else {
		log.WithField("wfr", wfrOperator.GetWorkflowRun().Name).
			WithField("stg", p.stage).
			WithField("status", v1alpha1.StatusSucceeded).
			Info("To update stage status")
		wfrOperator.UpdateStageStatus(p.stage, &v1alpha1.Status{
			Phase:              v1alpha1.StatusSucceeded,
			LastTransitionTime: metav1.Time{Time: time.Now()},
			Reason:             "CoordinatorCompleted",
			Message:            "Coordinator completed",
		})
	}
}
