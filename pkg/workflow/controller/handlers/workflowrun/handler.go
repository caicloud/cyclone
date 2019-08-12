package workflowrun

import (
	"bytes"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/meta"
	utilhttp "github.com/caicloud/cyclone/pkg/util/http"
	"github.com/caicloud/cyclone/pkg/util/kmutex"
	"github.com/caicloud/cyclone/pkg/workflow/common"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"
	"github.com/caicloud/cyclone/pkg/workflow/workflowrun"
)

// Handler handles changes of WorkflowRun CR.
type Handler struct {
	Client           clientset.Interface
	TimeoutProcessor *workflowrun.TimeoutProcessor
	GCProcessor      *workflowrun.GCProcessor
	LimitedQueues    *workflowrun.LimitedQueues
	notifyLock       *kmutex.Kmutex
}

// Ensure *Handler has implemented handlers.Interface interface.
var _ handlers.Interface = (*Handler)(nil)

// New creates a workflowrun Handler
func New(client clientset.Interface) *Handler {
	return &Handler{
		Client:           client,
		TimeoutProcessor: workflowrun.NewTimeoutProcessor(client),
		GCProcessor:      workflowrun.NewGCProcessor(client, controller.Config.GC.Enabled),
		LimitedQueues:    workflowrun.NewLimitedQueues(client, controller.Config.Limits.MaxWorkflowRuns),
		notifyLock:       kmutex.New(),
	}
}

// ObjectCreated handles a newly created WorkflowRun
func (h *Handler) ObjectCreated(obj interface{}) {
	originWfr, ok := obj.(*v1alpha1.WorkflowRun)
	if !ok {
		log.Warning("unknown resource type")
		return
	}
	log.WithField("name", originWfr.Name).Debug("Start to process WorkflowRun create")

	if !validate(originWfr) {
		return
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
		return
	}

	// Add this WorkflowRun to timeout processor, so that it would be cleaned up when time expired.
	err := h.TimeoutProcessor.Add(originWfr)
	if err != nil {
		log.WithField("wfr", originWfr.Name).Warn("add wfr to timeout processor failed", err)
	}

	wfr := originWfr.DeepCopy()
	clusterClient := common.GetExecutionClusterClient(wfr)
	if clusterClient == nil {
		log.WithField("wfr", wfr.Name).Error("Execution cluster client not found")
		return
	}

	operator, err := workflowrun.NewOperator(clusterClient, h.Client, wfr, wfr.Namespace)
	if err != nil {
		log.WithField("wfr", wfr.Name).Error("Failed to create workflowrun operator: ", err)
		return
	}

	operator.ResolveGlobalVariables()
	if err := operator.Reconcile(); err != nil {
		log.WithField("wfr", wfr.Name).Error("Reconcile error: ", err)
	}
}

// ObjectUpdated handles a updated WorkflowRun
func (h *Handler) ObjectUpdated(obj interface{}) {
	originWfr, ok := obj.(*v1alpha1.WorkflowRun)
	if !ok {
		log.Warning("unknown resource type")
		return
	}
	log.WithField("name", originWfr.Name).Debug("Start to process WorkflowRun update")

	if !validate(originWfr) {
		return
	}

	// Refresh updates 'refresh' time field of the WorkflowRun in the queue.
	h.LimitedQueues.Refresh(originWfr)

	// Add the WorkflowRun object to GC processor, it will be checked before actually added to
	// the GC queue.
	h.GCProcessor.Add(originWfr)

	// If the WorkflowRun has already been terminated(Completed, Failed, Cancelled), send notifications if necessary,
	// otherwise directly skip it.
	if workflowrun.IsWorkflowRunTerminated(originWfr) {
		// Send notification after workflowrun terminated.
		err := h.sendNotification(originWfr)
		if err != nil {
			log.WithField("wfr", originWfr.Name).Warn("send notification failed", err)
		}
		return
	}

	// If the WorkflowRun has already been terminated or waiting for external events, skip it.
	if originWfr.Status.Overall.Phase == v1alpha1.StatusSucceeded ||
		originWfr.Status.Overall.Phase == v1alpha1.StatusFailed ||
		originWfr.Status.Overall.Phase == v1alpha1.StatusWaiting {
		return
	}

	wfr := originWfr.DeepCopy()
	clusterClient := common.GetExecutionClusterClient(wfr)
	if clusterClient == nil {
		log.WithField("wfr", wfr.Name).Error("Execution cluster client not found")
		return
	}

	operator, err := workflowrun.NewOperator(clusterClient, h.Client, wfr, wfr.Namespace)
	if err != nil {
		log.WithField("wfr", wfr.Name).Error("Failed to create workflowrun operator: ", err)
		return
	}

	if err := operator.Reconcile(); err != nil {
		log.WithField("wfr", wfr.Name).Error("Reconcile error: ", err)
	}
}

// ObjectDeleted handles the case when a WorkflowRun get deleted. It will perform GC immediately for this WorkflowRun.
func (h *Handler) ObjectDeleted(obj interface{}) {
	originWfr, ok := obj.(*v1alpha1.WorkflowRun)
	if !ok {
		log.Warning("unknown resource type")
		return
	}
	log.WithField("name", originWfr.Name).Debug("Start to GC for WorkflowRun delete")

	wfr := originWfr.DeepCopy()
	clusterClient := common.GetExecutionClusterClient(wfr)
	if clusterClient == nil {
		log.WithField("wfr", wfr.Name).Error("Execution cluster client not found")
		return
	}

	operator, err := workflowrun.NewOperator(clusterClient, h.Client, wfr, wfr.Namespace)
	if err != nil {
		log.WithField("wfr", wfr.Name).Error("Failed to create workflowrun operator: ", err)
		return
	}

	err = operator.GC(true, true)
	if err != nil {
		log.WithField("wfr", wfr.Name).Warn("GC failed", err)
	}
}

// sendNotification sends notifications for workflowruns when:
// * notification endpoint is configured
// * without notification sent label
func (h *Handler) sendNotification(wfr *v1alpha1.WorkflowRun) error {
	h.notifyLock.Lock(wfr.Name)
	defer h.notifyLock.Unlock(wfr.Name)

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

// validate workflow run
func validate(wfr *v1alpha1.WorkflowRun) bool {
	// check workflowRef can not be nil
	if wfr.Spec.WorkflowRef == nil {
		log.WithField("name", wfr.Name).Error("WorkflowRef is nil")
		return false
	}

	return true
}
