package workflowrun

import (
	"fmt"
	"reflect"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
)

// resolveStatus determines the final status from two given status, one is latest status, and
// another one is the new status reported.
func resolveStatus(latest, update *v1alpha1.Status) *v1alpha1.Status {
	// If the latest status is already a terminated status (Completed, Failed, Cancelled), no need to
	// update it, we just return the latest status.
	if latest.Phase == v1alpha1.StatusSucceeded || latest.Phase == v1alpha1.StatusFailed || latest.Phase == v1alpha1.StatusCancelled {
		return latest
	}

	// If the latest status is not a terminated status, but the reported status is, then we
	// apply the reported status.
	if update.Phase == v1alpha1.StatusSucceeded || update.Phase == v1alpha1.StatusFailed || latest.Phase == v1alpha1.StatusCancelled {
		return update
	}

	// If both statuses are not terminated statuses, we select the one that with latest transition time.
	if update.LastTransitionTime.After(latest.LastTransitionTime.Time) {
		return update
	}

	return latest
}

// NextStages returns next stages that can be started to execute(stages that
// are not started yet but have all depended stages finished.) and stages that
// retry chances exhausted(Pending phase and retryStatus is not compliant with
// retry configuration).
func NextStages(duration time.Duration, times int, wf *v1alpha1.Workflow, wfr *v1alpha1.WorkflowRun) ([]string, []string) {
	var nextStages []string
	var nonRetryableStages []string
	for _, stage := range wf.Spec.Stages {

		if s, ok := wfr.Status.Stages[stage.Name]; ok {
			// If this stage already have status set and not pending, it means it's already been started, skip it.
			if s.Status.Phase != v1alpha1.StatusPending {
				continue
			}

			// If this stage already have status set pending and retryStatus is not nil, determine whether should
			// continue to execute it.
			if s.Status.RetryStatus != nil && (times <= 0 || duration.String() == "0s" ||
				s.Status.RetryStatus.Times > times ||
				time.Now().After(s.Status.RetryStatus.StartTime.Add(time.Second*duration))) {
				nonRetryableStages = append(nonRetryableStages, stage.Name)
				continue
			}
		}

		// All depended stages must have been successfully finished, otherwise this
		// stage would be skipped.
		safeToRun := true
		for _, d := range stage.Depends {
			status, ook := wfr.Status.Stages[d]
			if !(ook && (status.Status.Phase == v1alpha1.StatusSucceeded || (status.Status.Phase == v1alpha1.StatusFailed && IsTrivial(wf, d)))) {
				safeToRun = false
				break
			}
		}

		if safeToRun {
			nextStages = append(nextStages, stage.Name)
		}
	}

	return nextStages, nonRetryableStages
}

// IsTrivial returns whether a stage is trivial in a workflow
func IsTrivial(wf *v1alpha1.Workflow, stage string) bool {
	for _, s := range wf.Spec.Stages {
		if s.Name == stage {
			return s.Trivial
		}
	}

	return false
}

// staticStatus masks timestamp in status, safe for comparison of status.
func staticStatus(status *v1alpha1.WorkflowRunStatus) *v1alpha1.WorkflowRunStatus {
	t := metav1.Time{Time: time.Unix(0, 0)}
	copy := status.DeepCopy()
	copy.Overall.LastTransitionTime = t
	for k := range status.Stages {
		copy.Stages[k].Status.LastTransitionTime = t
	}
	return copy
}

// workflowRunItem keeps track of a WorkflowRun object (by name and namespace), and a time
// to take an action on the WorkflowRun object, such as GC, timeout handling. Retry is also
// supported, 'retry' indicates the remaining retry count, 0 means no retry.
type workflowRunItem struct {
	// Name of the WorkflowRun object
	name string
	// Namespace of the WorkflowRun object
	namespace string
	// The time to take the action (GC or timeout handling)
	expireTime time.Time
	// If the action taken failed, how many times to retry, 0 means no retry
	retry int
}

func (i *workflowRunItem) String() string {
	return fmt.Sprintf("%s:%s", i.namespace, i.name)
}

// EnsureOwner ensures WorkflowRun's owner is set to the referred Workflow.
// So that when Workflow is deleted, related WorkflowRun would also be deleted.
func ensureOwner(client clientset.Interface, wf *v1alpha1.Workflow, wfr *v1alpha1.WorkflowRun) error {
	// If owner of Workflow already set, skip it.
	for _, owner := range wfr.OwnerReferences {
		if owner.Kind == reflect.TypeOf(v1alpha1.Workflow{}).Name() {
			return nil
		}
	}

	// Get Workflow if not available.
	if wf == nil {
		f, err := client.CycloneV1alpha1().Workflows(wfr.Namespace).Get(wfr.Spec.WorkflowRef.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		wf = f
	}

	// Append Workflow as owner of WorkflowRun.
	wfr.OwnerReferences = append(wfr.OwnerReferences, metav1.OwnerReference{
		APIVersion: v1alpha1.APIVersion,
		Kind:       reflect.TypeOf(v1alpha1.Workflow{}).Name(),
		Name:       wf.Name,
		UID:        wf.UID,
	})

	return nil
}

// IsWorkflowRunTerminated judges whether the WorkflowRun has be terminated.
// Return true if terminated, otherwise return false.
func IsWorkflowRunTerminated(wfr *v1alpha1.WorkflowRun) bool {
	if wfr.Status.Overall.Phase == v1alpha1.StatusSucceeded ||
		wfr.Status.Overall.Phase == v1alpha1.StatusFailed ||
		wfr.Status.Overall.Phase == v1alpha1.StatusCancelled {
		return true
	}

	return false
}

// GCPodName generates a pod name for GC pod
func GCPodName(wfr string) string {
	return fmt.Sprintf("wfrgc--%s", wfr)
}

// NextStageStatus returns the status of next to execute stage.
// returns a new Running status if the stage is first time to execute,
// and returns a Running status with old retryStatus if the stage is
// retryable.
func NextStageStatus(stagesStatus map[string]*v1alpha1.StageStatus, stage string) *v1alpha1.Status {
	if stageStatus, ok := stagesStatus[stage]; ok {
		status := stageStatus.Status.DeepCopy()
		if status.RetryStatus != nil {
			return &v1alpha1.Status{
				Phase:              v1alpha1.StatusRunning,
				Reason:             "StageInitialized",
				LastTransitionTime: metav1.Time{Time: time.Now()},
				StartTime:          metav1.Time{Time: time.Now()},
				RetryStatus:        status.RetryStatus,
			}
		}
	}

	return &v1alpha1.Status{
		Phase:              v1alpha1.StatusRunning,
		Reason:             "StageInitialized",
		LastTransitionTime: metav1.Time{Time: time.Now()},
		StartTime:          metav1.Time{Time: time.Now()},
	}
}
