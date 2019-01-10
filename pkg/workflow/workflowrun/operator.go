package workflowrun

import (
	"fmt"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/common"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
)

// Operator is used to perform operations on a WorkflowRun instance, such
// as update status, run next stages, garbage collection, etc.
type Operator interface {
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
	// 'lastTry' indicates whether this is the last time to perform GC,
	// if set to true, the WorkflowRun status will be marked as cleaned regardless
	// whether the GC action succeeded or not.
	GC(lastTry bool) error
	// Run next stages in the Workflow and resolve overall status.
	Reconcile() error
}

type operator struct {
	client   clientset.Interface
	recorder record.EventRecorder
	wf       *v1alpha1.Workflow
	wfr      *v1alpha1.WorkflowRun
}

// Ensure *Handler has implemented handlers.Interface interface.
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
					WithField("status", combined.Status.Overall.Status).
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
		o.wfr.Status.Stages[stage] = &v1alpha1.StageStatus{}
	}

	o.wfr.Status.Stages[stage].Status = *status
}

// UpdateStagePodInfo updates stage pod information to WorkflowRun.
func (o *operator) UpdateStagePodInfo(stage string, podInfo *v1alpha1.PodInfo) {
	if o.wfr.Status.Stages == nil {
		o.wfr.Status.Stages = make(map[string]*v1alpha1.StageStatus)
	}

	if _, ok := o.wfr.Status.Stages[stage]; !ok {
		o.wfr.Status.Stages[stage] = &v1alpha1.StageStatus{
			Status: v1alpha1.Status{
				Status: v1alpha1.StatusRunning,
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
	// If the WorkflowRun has no stage status recorded yet, we resolve the overall status as pending.
	if o.wfr.Status.Stages == nil || len(o.wfr.Status.Stages) == 0 {
		return &v1alpha1.Status{
			Status:             v1alpha1.StatusPending,
			LastTransitionTime: metav1.Time{Time: time.Now()},
		}, nil
	}

	var running, waiting, err bool
	for stage, status := range o.wfr.Status.Stages {
		switch status.Status.Status {
		case v1alpha1.StatusPending:
			log.WithField("stage", stage).Warn("Pending stage should not occur.")
		case v1alpha1.StatusRunning:
			running = true
		case v1alpha1.StatusWaiting:
			waiting = true
		case v1alpha1.StatusError:
			err = true
		case v1alpha1.StatusCompleted:
		default:
			log.WithField("stg", stage).
				WithField("status", status.Status.Status).
				Error("Unknown stage status observed.")
			err = true
		}
	}

	// If there are running stages, resolve the overall status as running.
	if running {
		return &v1alpha1.Status{
			Status:             v1alpha1.StatusRunning,
			LastTransitionTime: metav1.Time{Time: time.Now()},
		}, nil
	}

	// Then if there are waiting stages, resolve the overall status as waiting.
	if waiting {
		return &v1alpha1.Status{
			Status:             v1alpha1.StatusWaiting,
			LastTransitionTime: metav1.Time{Time: time.Now()},
		}, nil
	}

	// Then if there are failed stages, resolve the overall status as failed.
	if err {
		return &v1alpha1.Status{
			Status:             v1alpha1.StatusError,
			LastTransitionTime: metav1.Time{Time: time.Now()},
		}, nil
	}

	// If all recorded stages are in completed status, we still need to check whether there other
	// stages that are not executed yet.
	var e error
	if o.wf == nil {
		o.wf, e = o.client.CycloneV1alpha1().Workflows(o.wfr.Namespace).Get(o.wfr.Spec.WorkflowRef.Name, metav1.GetOptions{})
		if e != nil {
			return nil, e
		}
	}
	next := len(NextStages(o.wf, o.wfr))
	if next > 0 {
		return &v1alpha1.Status{
			Status:             v1alpha1.StatusRunning,
			LastTransitionTime: metav1.Time{Time: time.Now()},
		}, nil
	}

	// Finally, all stages have been completed and no more stages to run. We mark the WorkflowRun
	// overall stage as Completed.
	return &v1alpha1.Status{
		Status:             v1alpha1.StatusCompleted,
		LastTransitionTime: metav1.Time{Time: time.Now()},
	}, nil
}

// Reconcile finds next stages in the workflow to run and resolve WorkflowRun's overall status.
func (o *operator) Reconcile() error {
	if o.wfr.Status.Stages == nil {
		o.wfr.Status.Stages = make(map[string]*v1alpha1.StageStatus)
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
			Status:             v1alpha1.StatusRunning,
			Reason:             "StageInitialized",
			LastTransitionTime: metav1.Time{Time: time.Now()},
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

		// Generate pod for this stage.
		pod, err := NewPodBuilder(o.client, o.wf, o.wfr, stage).Build()
		if err != nil {
			log.WithField("wfr", o.wfr.Name).WithField("stg", stage).Error("Create pod manifest for stage error: ", err)
			o.recorder.Eventf(o.wfr, corev1.EventTypeWarning, "GeneratePodSpecError", "Generate pod for stage '%s' error: %v", stage, err)
			o.UpdateStageStatus(stage, &v1alpha1.Status{
				Status:             v1alpha1.StatusError,
				Reason:             "GeneratePodError",
				LastTransitionTime: metav1.Time{Time: time.Now()},
				Message:            fmt.Sprintf("Failed to generate pod: %v", err),
			})
			continue
		}
		log.WithField("stg", stage).Debug("Pod manifest created")

		// Create the generated pod.
		pod, err = o.client.CoreV1().Pods(o.wfr.Namespace).Create(pod)
		if err != nil {
			log.WithField("wfr", o.wfr.Name).WithField("stg", stage).Error("Create pod for stage error: ", err)
			o.recorder.Eventf(o.wfr, corev1.EventTypeWarning, "StagePodCreated", "Create pod for stage '%s' error: %v", stage, err)
			o.UpdateStageStatus(stage, &v1alpha1.Status{
				Status:             v1alpha1.StatusError,
				Reason:             "CreatePodError",
				LastTransitionTime: metav1.Time{Time: time.Now()},
				Message:            fmt.Sprintf("Failed to create pod: %v", err),
			})
			continue
		}

		o.recorder.Eventf(o.wfr, corev1.EventTypeNormal, "StagePodCreated", "Create pod for stage '%s' succeeded", stage)
		o.UpdateStageStatus(stage, &v1alpha1.Status{
			Status:             v1alpha1.StatusRunning,
			LastTransitionTime: metav1.Time{Time: time.Now()},
			Reason:             "StagePodCreated",
		})

		o.UpdateStagePodInfo(stage, &v1alpha1.PodInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		})
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
// 'lastTry' indicates whether this is the last try to perform GC on this WorkflowRun object,
// if set to true, the WorkflowRun would be marked as cleaned regardless whether the GC succeeded or not.
func (o *operator) GC(lastTry bool) error {
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
			o.recorder.Eventf(o.wfr, corev1.EventTypeWarning, "GC", "Delete pod '%s' error: %v", status.Pod.Name, err)
		}
	}

	// Create a gc pod to clean data on tmp PV.
	gcPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("wfrgc--%s", o.wfr.Name),
			Namespace: o.wfr.Namespace,
			Labels: map[string]string{
				common.WorkflowLabelName: "true",
			},
			Annotations: map[string]string{
				common.WorkflowRunAnnotationName: o.wfr.Name,
				common.GCAnnotationName:          "true",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: v1alpha1.APIVersion,
					Kind:       reflect.TypeOf(v1alpha1.WorkflowRun{}).Name(),
					Name:       o.wfr.Name,
					UID:        o.wfr.UID,
				},
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
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: common.DefaultPvVolumeName,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: controller.Config.PVC,
							ReadOnly:  false,
						},
					},
				},
			},
		},
	}

	_, err := o.client.CoreV1().Pods(o.wfr.Namespace).Create(gcPod)
	if err != nil {
		log.WithField("wfr", o.wfr.Name).Warn("Create GC pod failed")
		if !lastTry {
			return err
		}
		o.recorder.Eventf(o.wfr, corev1.EventTypeWarning, "GC", "Create GC pod error: %v", err)
	}

	o.recorder.Event(o.wfr, corev1.EventTypeNormal, "GC", "GC is performed succeed.")

	o.wfr.Status.Cleaned = true
	o.Update()

	return nil
}
