package workflowrun

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/meta"
	utilhttp "github.com/caicloud/cyclone/pkg/util/http"
	"github.com/caicloud/cyclone/pkg/workflow/common"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"
	"github.com/caicloud/cyclone/pkg/workflow/workflowrun"
)

// Handler handles changes of WorkflowRun CR.
type Handler struct {
	Client                clientset.Interface
	TimeoutProcessor      *workflowrun.TimeoutProcessor
	GCProcessor           *workflowrun.GCProcessor
	LimitedQueues         *workflowrun.LimitedQueues
	ParallelismController workflowrun.ParallelismController
}

// Ensure *Handler has implemented handlers.Interface interface.
var _ handlers.Interface = (*Handler)(nil)

// ObjectCreated handles a newly created WorkflowRun
func (h *Handler) ObjectCreated(obj interface{}) error {
	originWfr, ok := obj.(*v1alpha1.WorkflowRun)
	if !ok {
		log.Warning("unknown resource type")
		return nil
	}
	log.WithField("name", originWfr.Name).Debug("Start to process WorkflowRun create")

	if !validate(originWfr) {
		return nil
	}

	// AddOrRefresh adds a WorkflowRun to its corresponding queue, if the queue size exceed the
	// maximum size, the oldest one would be deleted. And if the WorkflowRun already exists in
	// the queue, its 'refresh' time field would be refreshed.
	h.LimitedQueues.AddOrRefresh(originWfr)

	// Add the WorkflowRun object to GC processor, it will be checked before actually added to
	// the GC queue.
	h.GCProcessor.Add(originWfr)

	// If the WorkflowRun has already been terminated or waiting for external events, skip it.
	if originWfr.Status.Overall.Phase == v1alpha1.StatusSucceeded ||
		originWfr.Status.Overall.Phase == v1alpha1.StatusFailed ||
		originWfr.Status.Overall.Phase == v1alpha1.StatusWaiting {
		return nil
	}

	// If the WorkflowRun has not yet be started to execute, check the parallelism constraints to determine
	// whether to execute it.
	if originWfr.Status.Overall.Phase == "" {
		switch h.ParallelismController.AttemptNew(originWfr.Namespace, originWfr.Spec.WorkflowRef.Name, originWfr.Name) {
		case workflowrun.AttemptActionQueued:
			log.WithField("wfr", originWfr.Name).Infof("Too many WorkflowRun are running, stay pending in queue, will retry in %s", common.ResyncPeriod.String())
			return nil
		case workflowrun.AttemptActionFailed:
			if err := h.SetStatus(originWfr.Namespace, originWfr.Name, &v1alpha1.Status{
				Phase:              v1alpha1.StatusFailed,
				Reason:             "TooManyWaiting",
				LastTransitionTime: metav1.Time{Time: time.Now()},
			}); err != nil {
				log.WithField("wfr", originWfr.Name).Error("Set status to Failed error, ", err)
				return nil
			}
		}
	}

	// Add this WorkflowRun to timeout processor, so that it would be cleaned up when time expired.
	err := h.TimeoutProcessor.Add(originWfr)
	if err != nil {
		log.WithField("wfr", originWfr.Name).Warn("add wfr to timeout processor failed: ", err)
	}

	wfr := originWfr.DeepCopy()
	clusterClient := common.GetExecutionClusterClient(wfr)
	if clusterClient == nil {
		log.WithField("wfr", wfr.Name).Error("Execution cluster client not found")
		return nil
	}

	operator, err := workflowrun.NewOperator(clusterClient, h.Client, wfr, wfr.Namespace)
	if err != nil {
		log.WithField("wfr", wfr.Name).Error("Failed to create workflowrun operator: ", err)
		return nil
	}

	operator.ResolveGlobalVariables()
	if err := operator.Reconcile(); err != nil {
		log.WithField("wfr", wfr.Name).Error("Reconcile error: ", err)
	}

	// TODO: return retryable error
	return nil
}

// ObjectUpdated handles a updated WorkflowRun
func (h *Handler) ObjectUpdated(old, new interface{}) error {
	originWfr, ok := new.(*v1alpha1.WorkflowRun)
	if !ok {
		log.Warning("unknown resource type")
		return nil
	}

	if !validate(originWfr) {
		log.WithField("wfr", originWfr.Name).Warning("Invalid wfr")
		return nil
	}

	// Refresh updates 'refresh' time field of the WorkflowRun in the queue.
	h.LimitedQueues.Refresh(originWfr)

	// If the WorkflowRun has not yet be started to execute, check the parallelism constraints to determine
	// whether to execute it.
	var toRun bool
	if originWfr.Status.Overall.Phase == "" {
		log.WithField("wfr", originWfr.Name).Info("Attempt to run WorkflowRun")
		switch h.ParallelismController.AttemptNew(originWfr.Namespace, originWfr.Spec.WorkflowRef.Name, originWfr.Name) {
		case workflowrun.AttemptActionQueued:
			log.WithField("wfr", originWfr.Name).Infof("Too many WorkflowRun are running, stay pending in queue, will retry in %s", common.ResyncPeriod.String())
			return nil
		case workflowrun.AttemptActionFailed:
			if err := h.SetStatus(originWfr.Namespace, originWfr.Name, &v1alpha1.Status{
				Phase:              v1alpha1.StatusFailed,
				Reason:             "TooManyWaiting",
				LastTransitionTime: metav1.Time{Time: time.Now()},
			}); err != nil {
				log.WithField("wfr", originWfr.Name).Error("Set status to Failed error, ", err)
				return nil
			}
		case workflowrun.AttemptActionStart:
			toRun = true
		}
	}

	if !toRun && reflect.DeepEqual(old, new) {
		return nil
	}
	log.WithField("name", originWfr.Name).Debug("Start to process WorkflowRun update")

	// Add the WorkflowRun object to GC processor, it will be checked before actually added to
	// the GC queue.
	h.GCProcessor.Add(originWfr)

	// If the WorkflowRun has already been terminated(Completed, Failed, Cancelled), send notifications if necessary,
	// otherwise directly skip it.
	if workflowrun.IsWorkflowRunTerminated(originWfr) {
		// If the WorkflowRun already terminated, mark it in the ParallelismController.
		h.ParallelismController.MarkFinished(originWfr.Namespace, originWfr.Spec.WorkflowRef.Name, originWfr.Name)

		// Send notification after workflowrun terminated.
		err := h.sendNotification(originWfr)
		if err != nil {
			log.WithField("wfr", originWfr.Name).Warn("send notification failed", err)
		}
		return nil
	}

	// If the WorkflowRun has already been terminated or waiting for external events, skip it.
	if originWfr.Status.Overall.Phase == v1alpha1.StatusSucceeded ||
		originWfr.Status.Overall.Phase == v1alpha1.StatusFailed ||
		originWfr.Status.Overall.Phase == v1alpha1.StatusWaiting {
		return nil
	}

	wfr := originWfr.DeepCopy()
	clusterClient := common.GetExecutionClusterClient(wfr)
	if clusterClient == nil {
		log.WithField("wfr", wfr.Name).Error("Execution cluster client not found")
		return nil
	}

	operator, err := workflowrun.NewOperator(clusterClient, h.Client, wfr, wfr.Namespace)
	if err != nil {
		log.WithField("wfr", wfr.Name).Error("Failed to create workflowrun operator: ", err)
		return nil
	}

	if err := operator.Reconcile(); err != nil {
		log.WithField("wfr", wfr.Name).Error("Reconcile error: ", err)
	}

	// TODO: return retryable error
	return nil
}

// ObjectDeleted handles the case when a WorkflowRun get deleted. It will perform GC immediately for this WorkflowRun.
func (h *Handler) ObjectDeleted(obj interface{}) error {
	originWfr, ok := obj.(*v1alpha1.WorkflowRun)
	if !ok {
		log.Warning("unknown resource type")
		return nil
	}
	log.WithField("name", originWfr.Name).Debug("Start to GC for WorkflowRun delete")

	// Mark the WorkflowRun terminated in ParallelismController
	defer func() {
		h.ParallelismController.MarkFinished(originWfr.Namespace, originWfr.Name, originWfr.Spec.WorkflowRef.Name)
	}()

	wfr := originWfr.DeepCopy()
	clusterClient := common.GetExecutionClusterClient(wfr)
	if clusterClient == nil {
		log.WithField("wfr", wfr.Name).Error("Execution cluster client not found")
		return nil
	}

	operator, err := workflowrun.NewOperator(clusterClient, h.Client, wfr, wfr.Namespace)
	if err != nil {
		log.WithField("wfr", wfr.Name).Error("Failed to create workflowrun operator: ", err)
		return nil
	}

	err = operator.GC(true, true)
	if err != nil {
		log.WithField("wfr", wfr.Name).Warn("GC failed", err)
	}

	return nil
}

// sendNotification sends notifications for workflowruns when:
// * notification endpoint is configured
// * without notification sent label
func (h *Handler) sendNotification(wfr *v1alpha1.WorkflowRun) error {
	// Get latest WorkflowRun.
	latestWfr, err := h.Client.CycloneV1alpha1().WorkflowRuns(wfr.Namespace).Get(wfr.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if meta.LabelExists(latestWfr.Labels, meta.LabelWorkflowRunNotificationSent) {
		return nil
	}

	sent := false
	defer func() {
		// Update WorkflowRun notification status with retry.
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			// Get latest WorkflowRun.
			latest, err := h.Client.CycloneV1alpha1().WorkflowRuns(wfr.Namespace).Get(wfr.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			latest.Labels = meta.AddNotificationSentLabel(latest.Labels, sent)
			_, err = h.Client.CycloneV1alpha1().WorkflowRuns(wfr.Namespace).Update(latest)
			return err
		})

		if err != nil {
			log.WithField("name", wfr.Name).Error("Update workflowrun notification sent label error: ", err)
		}
	}()

	url := controller.Config.NotificationURL
	if url != "" {
		// Send notifications with workflowrun.
		bodyBytes, err := json.Marshal(wfr)
		if err != nil {
			log.WithField("wfr", wfr.Name).Errorf("Failed to marshal workflowrun: %v", err)
			return err
		}
		body := bytes.NewReader(bodyBytes)

		req, err := http.NewRequest(http.MethodPost, url, body)
		if err != nil {
			log.WithField("wfr", wfr.Name).Errorf("Failed to new notification request: %v", err)
			return err
		}
		// Set Json content type in Http header.
		req.Header.Set(utilhttp.HeaderContentType, utilhttp.HeaderContentTypeJSON)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.WithField("wfr", wfr.Name).Errorf("Failed to send notification: %v", err)
			return err
		}

		log.WithField("wfr", wfr.Name).Infof("Status code of notification: %d", resp.StatusCode)
		sent = true
	}

	return nil
}

// SetStatus sets workflowRun status
func (h *Handler) SetStatus(ns, wfr string, status *v1alpha1.Status) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := h.Client.CycloneV1alpha1().WorkflowRuns(ns).Get(wfr, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if latest.Status.Overall.Phase == status.Phase {
			return nil
		}

		toUpdate := latest.DeepCopy()
		toUpdate.Status.Overall = *status
		_, err = h.Client.CycloneV1alpha1().WorkflowRuns(latest.Namespace).Update(toUpdate)
		if err == nil {
			log.WithField("wfr", toUpdate.Name).
				WithField("status", status.Phase).
				Info("WorkflowRun status updated successfully.")
		}

		return err
	})
}

// validate workflow run
func validate(wfr *v1alpha1.WorkflowRun) bool {
	// check workflowRef can not be nil
	if wfr.Spec.WorkflowRef == nil {
		log.WithField("name", wfr.Name).Error("WorkflowRef is nil")
		return false
	}

	return true
}
