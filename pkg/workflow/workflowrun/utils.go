package workflowrun

import (
	"fmt"
	"reflect"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/util"
)

// resolveStatus determines the final status from two given status, one is latest status, and
// another one is the new status reported.
func resolveStatus(latest, update *v1alpha1.Status) *v1alpha1.Status {
	// If the latest status is already a terminated status (Completed, Failed, Cancelled), no need to
	// update it, we just return the latest status.
	if util.IsPhaseTerminated(latest.Phase) {
		return latest
	}

	// If the latest status is not a terminated status, but the reported status is, then we
	// apply the reported status.
	if util.IsPhaseTerminated(update.Phase) {
		return update
	}

	// If both statuses are not terminated statuses, we select the one that with latest transition time.
	if update.LastTransitionTime.After(latest.LastTransitionTime.Time) {
		return update
	}

	return latest
}

// NextStages determine next stages that can be started to execute. It returns
// stages that are not started yet but have all depended stages finished.
func NextStages(wf *v1alpha1.Workflow, wfr *v1alpha1.WorkflowRun) []string {
	var nextStages []string
	for _, stage := range wf.Spec.Stages {
		// If this stage already have status set and not pending, it means it's already been started, skip it.
		if s, ok := wfr.Status.Stages[stage.Name]; ok && s.Status.Phase != v1alpha1.StatusPending {
			continue
		}

		// All depended stages must have been successfully finished, otherwise this
		// stage would be skipped.
		safeToRun := true
		for _, d := range stage.Depends {
			status, ok := wfr.Status.Stages[d]
			if !(ok && (status.Status.Phase == v1alpha1.StatusSucceeded || (status.Status.Phase == v1alpha1.StatusFailed && IsTrivial(wf, d)) ||
				(status.Status.Phase == v1alpha1.StatusCancelled && IsTrivial(wf, d)))) {
				safeToRun = false
				break
			}
		}

		if safeToRun {
			nextStages = append(nextStages, stage.Name)
		}
	}

	return nextStages
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

// GCPodName generates a pod name for GC pod
func GCPodName(wfr string) string {
	return fmt.Sprintf("wfrgc--%s", wfr)
}
