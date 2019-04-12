package workflowrun

import (
	"fmt"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/workflow/common"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
)

// Operator is used to perform operations on a WorkflowRun instance, such
// as update status, run next stages, garbage collection, etc.
type Operator interface {
	// Get Recorder
	GetRecorder() record.EventRecorder
	// Get WorkflowRun instance.
	GetWorkflowRun() *v1alpha1.WorkflowRun
	// Update WorkflowRun, mainly the status.
	Update() error
	// Update stage status.
	UpdateStageStatus(stage string, status *v1alpha1.Status)
	// Update stage pod info.
	UpdateStagePodInfo(stage string, podInfo *v1alpha1.PodInfo)
	// Decide overall status of the WorkflowRun from stage status.
	OverallStatus() (*v1alpha1.Status, error)
	// Garbage collection on the WorkflowRun based on GC policy configured
	// in Workflow Controller. Pod and data on PV would be cleaned.
	// - 'lastTry' indicates whether this is the last time to perform GC,
	// if set to true, the WorkflowRun status will be marked as cleaned regardless
	// whether the GC action succeeded or not.
	// - 'wfrDeletion' indicates whether the GC is performed because of WorkflowRun deleted.
	GC(lastTry, wfrDeletion bool) error
	// Run next stages in the Workflow and resolve overall status.
	Reconcile() error
}

type operator struct {
	client   clientset.Interface
	recorder record.EventRecorder
	wf       *v1alpha1.Workflow
	wfr      *v1alpha1.WorkflowRun
}

// Ensure *operator has implemented Operator interface.
var _ Operator = (*operator)(nil)

// NewOperator create a new operator.
func NewOperator(client clientset.Interface, wfr interface{}, namespace string) (Operator, error) {
	if w, ok := wfr.(string); ok {
		return newFromName(client, w, namespace)
	}

	if w, ok := wfr.(*v1alpha1.WorkflowRun); ok {
		return newFromValue(client, w, namespace)
	}

	return nil, fmt.Errorf("invalid parameter 'wfr' provided: %v", wfr)
}

// When create Operator from WorkflowRun name, we only get WorkflowRun value, but not for
// Workflow.
func newFromName(client clientset.Interface, wfr, namespace string) (Operator, error) {
	w, err := client.CycloneV1alpha1().WorkflowRuns(namespace).Get(wfr, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &operator{
		client:   client,
		recorder: common.GetEventRecorder(client, common.EventSourceWfrController),
		wfr:      w,
	}, nil
}

// When create Operator from WorkflowRun value, we will also get Workflow value.
func newFromValue(client clientset.Interface, wfr *v1alpha1.WorkflowRun, namespace string) (Operator, error) {
	f, err := client.CycloneV1alpha1().Workflows(namespace).Get(wfr.Spec.WorkflowRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &operator{
		client:   client,
		recorder: common.GetEventRecorder(client, common.EventSourceWfrController),
		wf:       f,
		wfr:      wfr,
	}, nil
}

// GetWorkflowRun returns the WorkflowRun object.
func (o *operator) GetWorkflowRun() *v1alpha1.WorkflowRun {
	return o.wfr
}

// GetRecorder returns the event recorder.
func (o *operator) GetRecorder() record.EventRecorder {
	return o.recorder
}

// InitStagesStatus initializes all missing stages' status to pending.
func (o *operator) InitStagesStatus() {
	if o.wfr.Status.Stages == nil {
		o.wfr.Status.Stages = make(map[string]*v1alpha1.StageStatus)
	}

	for _, stg := range o.wf.Spec.Stages {
		if _, ok := o.wfr.Status.Stages[stg.Name]; !ok {
			o.wfr.Status.Stages[stg.Name] = &v1alpha1.StageStatus{
				Status: v1alpha1.Status{
					Phase: v1alpha1.StatusPending,
				},
			}
		}
	}
}

// Update the WorkflowRun status, it retrieves the latest WorkflowRun and apply changes to
// it, then update it with retry.
func (o *operator) Update() error {
	if o.wfr == nil {
		return nil
	}

	// Update WorkflowRun status with retry.
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get latest WorkflowRun.
		latest, err := o.client.CycloneV1alpha1().WorkflowRuns(o.wfr.Namespace).Get(o.wfr.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		combined := latest.DeepCopy()
		if combined.Status.Stages == nil {
			combined.Status.Stages = make(map[string]*v1alpha1.StageStatus)
		}

		// Ensure it has owner reference to related Workflow.
		if err := ensureOwner(o.client, o.wf, combined); err != nil {
			log.WithField("wfr", combined.Name).Warn("Ensure owner error: ", err)
		}

		// Apply changes to latest WorkflowRun
		combined.Status.Cleaned = combined.Status.Cleaned || o.wfr.Status.Cleaned
		combined.Status.Overall = *resolveStatus(&combined.Status.Overall, &o.wfr.Status.Overall)
		for stage, status := range o.wfr.Status.Stages {
			s, ok := combined.Status.Stages[stage]
			if !ok {
				combined.Status.Stages[stage] = status
				continue
			}

			combined.Status.Stages[stage].Status = *resolveStatus(&s.Status, &status.Status)
			if s.Pod == nil {
				combined.Status.Stages[stage].Pod = status.Pod
			}
			if len(s.Outputs) == 0 {
				combined.Status.Stages[stage].Outputs = status.Outputs
			}
		}

		if !reflect.DeepEqual(staticStatus(&latest.Status), staticStatus(&combined.Status)) ||
			len(latest.OwnerReferences) != len(combined.OwnerReferences) {
			_, err = o.client.CycloneV1alpha1().WorkflowRuns(latest.Namespace).Update(combined)
			if err == nil {
				log.WithField("wfr", latest.Name).
					WithField("status", combined.Status.Overall.Phase).
					WithField("cleaned", combined.Status.Cleaned).
					Info("WorkflowRun status updated successfully.")
			}
			return err
		}

		return nil
	})
}

// UpdateStageStatus updates status of a stage in WorkflowRun status part.
func (o *operator) UpdateStageStatus(stage string, status *v1alpha1.Status) {
	if o.wfr.Status.Stages == nil {
		o.wfr.Status.Stages = make(map[string]*v1alpha1.StageStatus)
	}

	if _, ok := o.wfr.Status.Stages[stage]; !ok {
		o.wfr.Status.Stages[stage] = &v1alpha1.StageStatus{
			Status: *status,
		}
	} else {
		// keep startTime unchanged
		originStatus := o.wfr.Status.Stages[stage].Status
		o.wfr.Status.Stages[stage].Status = *status
		if originStatus.Phase != v1alpha1.StatusPending {
			o.wfr.Status.Stages[stage].Status.StartTime = originStatus.StartTime
		}
	}
}

// UpdateStagePodInfo updates stage pod information to WorkflowRun.
func (o *operator) UpdateStagePodInfo(stage string, podInfo *v1alpha1.PodInfo) {
	if o.wfr.Status.Stages == nil {
		o.wfr.Status.Stages = make(map[string]*v1alpha1.StageStatus)
	}

	if _, ok := o.wfr.Status.Stages[stage]; !ok {
		o.wfr.Status.Stages[stage] = &v1alpha1.StageStatus{
			Status: v1alpha1.Status{
				Phase: v1alpha1.StatusRunning,
			},
		}
	}

	o.wfr.Status.Stages[stage].Pod = podInfo
}

// OverallStatus calculates the overall status of the WorkflowRun. When a stage has its status
// changed, the change will be updated in WorkflowRun stage status, but the overall status is
// not calculated. So when we observed a WorkflowRun updated, we need to calculate its overall
// status and update it if changed.
func (o *operator) OverallStatus() (*v1alpha1.Status, error) {
	startTime := o.wfr.ObjectMeta.CreationTimestamp
	// If the WorkflowRun has no stage status recorded yet, we resolve the overall status as pending.
	if o.wfr.Status.Stages == nil || len(o.wfr.Status.Stages) == 0 {
		return &v1alpha1.Status{
			Phase:              v1alpha1.StatusPending,
			LastTransitionTime: metav1.Time{Time: time.Now()},
			StartTime:          startTime,
		}, nil
	}

	var running, waiting, pending, err bool
	for stage, status := range o.wfr.Status.Stages {
		switch status.Status.Phase {
		case v1alpha1.StatusRunning:
			running = true
		case v1alpha1.StatusWaiting:
			waiting = true
		case v1alpha1.StatusFailed:
			err = !IsTrivial(o.wf, stage)
		case v1alpha1.StatusPending:
			pending = true
		case v1alpha1.StatusSucceeded:
		default:
			log.WithField("stg", stage).
				WithField("status", status.Status.Phase).
				Error("Unknown stage status observed.")
			err = true
		}
	}

	// If there are running stages, resolve the overall status as running.
	if running {
		return &v1alpha1.Status{
			Phase:              v1alpha1.StatusRunning,
			LastTransitionTime: metav1.Time{Time: time.Now()},
			StartTime:          startTime,
		}, nil
	}

	// Then if there are waiting stages, resolve the overall status as waiting.
	if waiting {
		return &v1alpha1.Status{
			Phase:              v1alpha1.StatusWaiting,
			LastTransitionTime: metav1.Time{Time: time.Now()},
			StartTime:          startTime,
		}, nil
	}

	// Then if there are failed stages, resolve the overall status as failed.
	if err {
		return &v1alpha1.Status{
			Phase:              v1alpha1.StatusFailed,
			LastTransitionTime: metav1.Time{Time: time.Now()},
			StartTime:          startTime,
		}, nil
	}

	// If there are still stages waiting for running, we set status to Running.
	// Here we assumed all stage statues have be initialized to Pending before wfr execution.
	if pending {
		return &v1alpha1.Status{
			Phase:              v1alpha1.StatusRunning,
			LastTransitionTime: metav1.Time{Time: time.Now()},
			StartTime:          startTime,
		}, nil
	}

	// Finally, all stages have been completed and no more stages to run. We mark the WorkflowRun
	// overall stage as Completed.
	return &v1alpha1.Status{
		Phase:              v1alpha1.StatusSucceeded,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		StartTime:          startTime,
	}, nil
}

// Reconcile finds next stages in the workflow to run and resolve WorkflowRun's overall status.
func (o *operator) Reconcile() error {
	if o.wfr.Status.Stages == nil {
		o.InitStagesStatus()
	}

	// Get next stages that need to be run.
	nextStages := NextStages(o.wf, o.wfr)
	if len(nextStages) == 0 {
		log.WithField("wfr", o.wfr.Name).Debug("No next stages to run")
	} else {
		log.WithField("stg", nextStages).Info("Next stages to run")
	}

	for _, stage := range nextStages {
		o.UpdateStageStatus(stage, &v1alpha1.Status{
			Phase:              v1alpha1.StatusRunning,
			Reason:             "StageInitialized",
			LastTransitionTime: metav1.Time{Time: time.Now()},
			StartTime:          metav1.Time{Time: time.Now()},
		})
	}
	overall, err := o.OverallStatus()
	if err != nil {
		return fmt.Errorf("resolve overall status error: %v", err)
	}
	o.wfr.Status.Overall = *overall
	err = o.Update()
	if err != nil {
		log.WithField("wfr", o.wfr.Name).Error("Update status error: ", err)
		return err
	}

	// Return if no stages need to run.
	if len(nextStages) == 0 {
		return nil
	}

	// Create pod to run stages.
	for _, stage := range nextStages {
		log.WithField("stg", stage).Info("Start to run stage")

		stg, err := o.client.CycloneV1alpha1().Stages(o.wfr.Namespace).Get(stage, metav1.GetOptions{})
		if err != nil {
			continue
		}

		err = NewWorkloadProcessor(o.client, o.wf, o.wfr, stg, o).Process()
		if err != nil {
			log.WithField("stg", stage).Error("Process workload error: ", err)
			continue
		}
	}

	overall, err = o.OverallStatus()
	if err != nil {
		return fmt.Errorf("resolve overall status error: %v", err)
	}
	o.wfr.Status.Overall = *overall
	err = o.Update()
	if err != nil {
		log.WithField("wfr", o.wfr.Name).Error("Update status error: ", err)
		return err
	}

	return nil
}

// Garbage collection of WorkflowRun. When it's terminated, we will cleanup the pods created by it.
// - 'lastTry' indicates whether this is the last try to perform GC on this WorkflowRun object,
// if set to true, the WorkflowRun would be marked as cleaned regardless whether the GC succeeded or not.
// - 'wfrDeletion' indicates whether the GC is performed because of WorkflowRun deleted. In this case,
// GC would performed silently, without event recorded, withoug status update.
func (o *operator) GC(lastTry, wfrDeletion bool) error {
	// For each pod created, delete it.
	for stg, status := range o.wfr.Status.Stages {
		if status.Pod == nil {
			log.WithField("wfr", o.wfr.Name).
				WithField("stg", stg).
				Warn("Pod information is missing, can't clean the pod.")
			continue
		}
		err := o.client.CoreV1().Pods(status.Pod.Namespace).Delete(status.Pod.Name, &metav1.DeleteOptions{})
		if err != nil {
			// If the pod not exist, just skip it without complain.
			if errors.IsNotFound(err) {
				continue
			}
			log.WithField("wfr", o.wfr.Name).
				WithField("stg", stg).
				WithField("pod", status.Pod.Name).
				Warn("Delete pod error: ", err)

			if !wfrDeletion {
				o.recorder.Eventf(o.wfr, corev1.EventTypeWarning, "GC", "Delete pod '%s' error: %v", status.Pod.Name, err)
			}
		}
		log.WithField("ns", status.Pod.Namespace).WithField("pod", status.Pod.Name).Info("Pod deleted")
	}

	// Get exeuction context of the WorkflowRun, namespace and PVC are defined in the context.
	executionContext := GetExecutionContext(o.wfr)

	// Create a gc pod to clean data on PV if PVC is configured.
	if executionContext.PVC != "" {
		gcPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      GCPodName(o.wfr.Name),
				Namespace: executionContext.Namespace,
				Labels: map[string]string{
					meta.LabelWorkflowRunName: o.wfr.Name,
					meta.LabelPodKind:         meta.PodKindGC.String(),
					meta.LabelPodCreatedBy:    meta.CycloneCreator,
				},
			},
			Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyNever,
				Containers: []corev1.Container{
					{
						Name:    common.GCContainerName,
						Image:   controller.Config.Images[controller.GCImage],
						Command: []string{"rm", "-rf", common.GCDataPath + "/" + o.wfr.Name},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      common.DefaultPvVolumeName,
								MountPath: common.GCDataPath,
								SubPath:   common.WorkflowRunsPath(),
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("64Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("64Mi"),
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: common.DefaultPvVolumeName,
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: executionContext.PVC,
								ReadOnly:  false,
							},
						},
					},
				},
			},
		}

		_, err := o.client.CoreV1().Pods(executionContext.Namespace).Create(gcPod)
		if err != nil {
			log.WithField("wfr", o.wfr.Name).Warn("Create GC pod error: ", err)
			if !lastTry {
				return err
			}

			if !wfrDeletion {
				o.recorder.Eventf(o.wfr, corev1.EventTypeWarning, "GC", "Create GC pod error: %v", err)
			}
		}
	}

	if !wfrDeletion {
		o.recorder.Event(o.wfr, corev1.EventTypeNormal, "GC", "GC is performed succeed.")

		o.wfr.Status.Cleaned = true
		o.Update()
	}

	return nil
}
