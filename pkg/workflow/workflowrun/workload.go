package workflowrun

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/util/retry"
	"github.com/caicloud/cyclone/pkg/workflow/workload/delegation"
	"github.com/caicloud/cyclone/pkg/workflow/workload/pod"
)

// WorkloadProcessor processes stage workload. There are kinds of workload supported: pod, delegation.
// With pod, Cyclone would create a pod to run the stage. With delegation, Cyclone would send
// a POST request to the given URL in the workload spec.
type WorkloadProcessor struct {
	clusterClient kubernetes.Interface
	client        clientset.Interface
	wf            *v1alpha1.Workflow
	wfr           *v1alpha1.WorkflowRun
	stg           *v1alpha1.Stage
	wfrOper       Operator
}

// NewWorkloadProcessor ...
func NewWorkloadProcessor(clusterClient kubernetes.Interface, client clientset.Interface, wf *v1alpha1.Workflow, wfr *v1alpha1.WorkflowRun, stage *v1alpha1.Stage, wfrOperator Operator) *WorkloadProcessor {
	return &WorkloadProcessor{
		client:        client,
		clusterClient: clusterClient,
		wf:            wf,
		wfr:           wfr,
		stg:           stage,
		wfrOper:       wfrOperator,
	}
}

// Process processes the stage according to workload type.
func (p *WorkloadProcessor) Process() error {
	if p.stg.Spec.Pod != nil && p.stg.Spec.Delegation != nil {
		return fmt.Errorf("exact 1 workload (pod or delegation) expected in stage '%s/%s', but got both", p.stg.Namespace, p.stg.Name)
	}
	if p.stg.Spec.Pod == nil && p.stg.Spec.Delegation == nil {
		return fmt.Errorf("exact 1 workload (pod or delegation) expected in stage '%s/%s', but got none", p.stg.Namespace, p.stg.Name)
	}

	if p.stg.Spec.Pod != nil {
		return p.processPod()
	}

	if p.stg.Spec.Delegation != nil {
		return p.processDelegation()
	}

	return nil
}

func (p *WorkloadProcessor) processPod() error {
	// Generate pod for this stage.
	po, err := pod.NewBuilder(p.client, p.wf, p.wfr, p.stg).Build()
	if err != nil {
		log.WithField("wfr", p.wfr.Name).WithField("stg", p.stg.Name).Error("Create pod manifest for stage error: ", err)
		p.wfrOper.GetRecorder().Eventf(p.wfr, corev1.EventTypeWarning, "GeneratePodSpecError", "Generate pod for stage '%s' error: %v", p.stg.Name, err)
		p.wfrOper.UpdateStageStatus(p.stg.Name, &v1alpha1.Status{
			Phase:              v1alpha1.StatusFailed,
			Reason:             "GeneratePodError",
			LastTransitionTime: metav1.Time{Time: time.Now()},
			Message:            fmt.Sprintf("Failed to generate pod: %v", err),
		})
		return err
	}
	log.WithField("stg", p.stg.Name).Debug("Pod manifest created")

	// Create the generated pod with retry on exceeded quota.
	// Here is a litter tricky. Cyclone will delete stage related pod to release cpu/memory resource when stage have
	// been finished, but pod deletion needs some time, so retry on exceeded quota gives the time to waiting previous
	// stage pod deletion.
	backoff := wait.Backoff{
		Steps:    3,
		Duration: 5 * time.Second,
		Factor:   1.5,
		Jitter:   0.1,
	}
	origin := po.DeepCopy()
	err = retry.OnExceededQuota(backoff, func() error {
		po, err = p.clusterClient.CoreV1().Pods(pod.GetExecutionContext(p.wfr).Namespace).Create(origin)
		return err
	})
	if err != nil {
		log.WithField("wfr", p.wfr.Name).WithField("stg", p.stg.Name).Error("Create pod for stage error: ", err)
		p.wfrOper.GetRecorder().Eventf(p.wfr, corev1.EventTypeWarning, "StagePodCreated", "Create pod for stage '%s' error: %v", p.stg.Name, err)
		p.wfrOper.UpdateStageStatus(p.stg.Name, &v1alpha1.Status{
			Phase:              v1alpha1.StatusFailed,
			Reason:             "CreatePodError",
			LastTransitionTime: metav1.Time{Time: time.Now()},
			Message:            fmt.Sprintf("Failed to create pod: %v", err),
		})
		return err
	}

	log.WithField("wfr", p.wfr.Name).WithField("stg", p.stg.Name).Debug("Create pod for stage succeeded")
	p.wfrOper.GetRecorder().Eventf(p.wfr, corev1.EventTypeNormal, "StagePodCreated", "Create pod for stage '%s' succeeded", p.stg.Name)
	p.wfrOper.UpdateStageStatus(p.stg.Name, &v1alpha1.Status{
		Phase:              v1alpha1.StatusRunning,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             "StagePodCreated",
	})

	p.wfrOper.UpdateStagePodInfo(p.stg.Name, &v1alpha1.PodInfo{
		Name:      po.Name,
		Namespace: po.Namespace,
	})

	return nil
}

func (p *WorkloadProcessor) processDelegation() error {
	err := delegation.Delegate(&delegation.Request{
		Stage:       p.stg,
		Workflow:    p.wf,
		WorkflowRun: p.wfr,
	})

	if err != nil {
		p.wfrOper.GetRecorder().Eventf(p.wfr, corev1.EventTypeWarning, "DelegationFailure", "Delegate stage %s to %s error: %v", p.stg.Name, p.stg.Spec.Delegation.URL, err)
		p.wfrOper.UpdateStageStatus(p.stg.Name, &v1alpha1.Status{
			Phase:              v1alpha1.StatusFailed,
			Reason:             "DelegationFailure",
			LastTransitionTime: metav1.Time{Time: time.Now()},
			Message:            fmt.Sprintf("Delegate error: %v", err),
		})
		return err
	}

	p.wfrOper.GetRecorder().Eventf(p.wfr, corev1.EventTypeNormal, "DelegationSucceed", "Delegate stage %s to %s succeeded", p.stg.Name, p.stg.Spec.Delegation.URL)
	p.wfrOper.UpdateStageStatus(p.stg.Name, &v1alpha1.Status{
		Phase:              v1alpha1.StatusWaiting,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             "DelegationSucceed",
	})

	return nil
}
