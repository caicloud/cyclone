package workflowrun

import (
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	ccommon "github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/common/values"
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
	// Update stage outputs, they are key-value results from stage execution
	UpdateStageOutputs(stage string, keyValues []v1alpha1.KeyValue)
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
	// ResolveGlobalVariables resolves global variables from workflow.
	ResolveGlobalVariables()
}

type operator struct {
	clusterClient kubernetes.Interface
	client        clientset.Interface
	recorder      record.EventRecorder
	wf            *v1alpha1.Workflow
	wfr           *v1alpha1.WorkflowRun
}

// Ensure *operator has implemented Operator interface.
var _ Operator = (*operator)(nil)

// NewOperator create a new operator.
func NewOperator(clusterClient kubernetes.Interface, client clientset.Interface, wfr interface{}, namespace string) (Operator, error) {
	if w, ok := wfr.(string); ok {
		return newFromName(clusterClient, client, w, namespace)
	}

	if w, ok := wfr.(*v1alpha1.WorkflowRun); ok {
		return newFromValue(clusterClient, client, w, namespace)
	}

	return nil, fmt.Errorf("invalid parameter 'wfr' provided: %v", wfr)
}

// When create Operator from WorkflowRun name, we only get WorkflowRun value, but not for
// Workflow.
func newFromName(clusterClient kubernetes.Interface, client clientset.Interface, wfr, namespace string) (Operator, error) {
	w, err := client.CycloneV1alpha1().WorkflowRuns(namespace).Get(wfr, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &operator{
		clusterClient: clusterClient,
		client:        client,
		recorder:      common.GetEventRecorder(client, common.EventSourceWfrController),
		wfr:           w,
	}, nil
}

// When create Operator from WorkflowRun value, we will also get Workflow value.
func newFromValue(clusterClient kubernetes.Interface, client clientset.Interface, wfr *v1alpha1.WorkflowRun, namespace string) (Operator, error) {
	f, err := client.CycloneV1alpha1().Workflows(namespace).Get(wfr.Spec.WorkflowRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &operator{
		clusterClient: clusterClient,
		client:        client,
		recorder:      common.GetEventRecorder(client, common.EventSourceWfrController),
		wf:            f,
		wfr:           wfr,
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

// InitStagesStatus initializes all missing stages' status to pending, and record workflow topology at this time to workflowRun status.
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
				Depends: stg.Depends,
				Trivial: stg.Trivial,
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
			if len(s.Depends) == 0 {
				combined.Status.Stages[stage].Depends = status.Depends
			}
			combined.Status.Stages[stage].Trivial = status.Trivial
		}

		// Update golbal variables to resolved values
		combined.Spec.GlobalVariables = o.wfr.Spec.GlobalVariables

		if !reflect.DeepEqual(staticStatus(&latest.Status), staticStatus(&combined.Status)) ||
			len(latest.OwnerReferences) != len(combined.OwnerReferences) {

			// If status has any change, the overall last transition time need to update
			combined.Status.Overall.LastTransitionTime = metav1.Time{Time: time.Now()}

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

// UpdateStageOutputs updates stage outputs, they are key-value results from stage execution
func (o *operator) UpdateStageOutputs(stage string, keyValues []v1alpha1.KeyValue) {
	if len(keyValues) == 0 {
		return
	}

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

	o.wfr.Status.Stages[stage].Outputs = keyValues
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
			err = err || !IsTrivial(o.wf, stage)
		case v1alpha1.StatusPending:
			pending = true
		case v1alpha1.StatusSucceeded:
		case v1alpha1.StatusCancelled:
			err = err || !IsTrivial(o.wf, stage)
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
			log.WithField("stg", stage).Error("Get stage error: ", err)
			continue
		}

		err = NewWorkloadProcessor(o.clusterClient, o.client, o.wf, o.wfr, stg, o).Process()
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
// GC would performed silently, without event recording and status updating.
func (o *operator) GC(lastTry, wfrDeletion bool) error {
	wg := sync.WaitGroup{}
	allPodsFinished := true
	// For each pod created, delete it.
	for stg, status := range o.wfr.Status.Stages {
		// For non-terminated stage, update status to cancelled.
		if status.Status.Phase == v1alpha1.StatusPending ||
			status.Status.Phase == v1alpha1.StatusRunning ||
			status.Status.Phase == v1alpha1.StatusWaiting {
			o.UpdateStageStatus(stg, &v1alpha1.Status{
				Phase:              v1alpha1.StatusCancelled,
				Reason:             "GC",
				LastTransitionTime: metav1.Time{Time: time.Now()},
			})
		}

		if status.Pod == nil {
			log.WithField("wfr", o.wfr.Name).
				WithField("stg", stg).
				Warn("Pod information is missing, can't clean the pod.")
			continue
		}
		err := o.clusterClient.CoreV1().Pods(status.Pod.Namespace).Delete(status.Pod.Name, &metav1.DeleteOptions{})
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
		} else {
			log.WithField("ns", status.Pod.Namespace).WithField("pod", status.Pod.Name).Info("Start to delete pod")

			wg.Add(1)
			go func(namespace, podName string) {
				defer wg.Done()

				timeout := time.After(5 * time.Minute)
				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-timeout:
						allPodsFinished = false
						log.WithField("ns", namespace).WithField("pod", podName).Warn("Pod deletion timeout")
						return
					case <-ticker.C:
						_, err := o.clusterClient.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
						if err != nil && errors.IsNotFound(err) {
							log.WithField("ns", namespace).WithField("pod", podName).Info("Pod deleted")
							return
						}
					}
				}
			}(status.Pod.Namespace, status.Pod.Name)
		}
	}

	// Wait all workflowRun related workload pods deleting completed, then start gc pod to clean data on PV.
	// Otherwise, if the path which is used by workload pods in the PV is deleted before workload pods deletion,
	// the pod deletion process will get stuck on Terminating status.
	wg.Wait()

	// If there are pods not finished and this is not the last gc try process, we will not start gc pod to clean
	// data on PV. The last gc try process will ensure data could be cleaned.
	if !allPodsFinished && !lastTry {
		if !wfrDeletion {
			o.recorder.Eventf(o.wfr, corev1.EventTypeWarning, "GC", "There are stage pods not Finished")
		}
		return nil
	}

	// Get execution context of the WorkflowRun, namespace and PVC are defined in the context.
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
				Annotations: map[string]string{
					meta.AnnotationIstioInject: meta.AnnotationValueFalse,
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
						Resources: controller.Config.GC.ResourceRequirements,
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

		// If controller instance name is set, add label to the pod created.
		if instance := os.Getenv(ccommon.ControllerInstanceEnvName); len(instance) != 0 {
			gcPod.ObjectMeta.Labels[meta.LabelControllerInstance] = instance
		}

		_, err := o.clusterClient.CoreV1().Pods(executionContext.Namespace).Create(gcPod)
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
		err := o.Update()
		if err != nil {
			log.WithField("wfr", o.wfr.Name).Warn("Update wfr error: ", err)
		}
	}

	return nil
}

// ResolveGlobalVariables will resolve global variables in workflowrun For example, generate final value for generation
// type value defined in workflow. For example, $(random:5) --> 'axyps'
func (o *operator) ResolveGlobalVariables() {
	if o.wf == nil || o.wfr == nil {
		return
	}

	var appendVariables []v1alpha1.GlobalVariable
	for _, wfVariable := range o.wf.Spec.GlobalVariables {
		var found bool
		for _, variable := range o.wfr.Spec.GlobalVariables {
			if variable.Name == wfVariable.Name {
				found = true
				break
			}
		}

		if !found {
			appendVariables = append(appendVariables, v1alpha1.GlobalVariable{
				Name:  wfVariable.Name,
				Value: values.GenerateValue(wfVariable.Value),
			})
		}
	}

	o.wfr.Spec.GlobalVariables = append(o.wfr.Spec.GlobalVariables, appendVariables...)

	if len(appendVariables) > 0 {
		log.WithField("variables", appendVariables).Info("Append variables from wf to wfr")
	}
}
