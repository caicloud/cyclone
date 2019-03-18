package workflowrun

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"
	"github.com/caicloud/cyclone/pkg/workflow/workflowrun"
)

const (
	contentType string = "Content-Type"
	contentJSON string = "application/json"
)

// controllerStartTime represents the start time of workflow controller,
// use to avoid sending notifications for workflowruns finished before workflow controller starts.
var controllerStartTime *metav1.Time

func init() {
	controllerStartTime = &metav1.Time{Time: time.Now()}
}

// Handler handles changes of WorkflowRun CR.
type Handler struct {
	Client           clientset.Interface
	TimeoutProcessor *workflowrun.TimeoutProcessor
	GCProcessor      *workflowrun.GCProcessor
	LimitedQueues    *workflowrun.LimitedQueues
}

// Ensure *Handler has implemented handlers.Interface interface.
var _ handlers.Interface = (*Handler)(nil)

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
	if originWfr.Status.Overall.Phase == v1alpha1.StatusCompleted ||
		originWfr.Status.Overall.Phase == v1alpha1.StatusError ||
		originWfr.Status.Overall.Phase == v1alpha1.StatusWaiting {
		return
	}

	// Add this WorkflowRun to timeout processor, so that it would be cleaned up when time expired.
	h.TimeoutProcessor.Add(originWfr)

	wfr := originWfr.DeepCopy()
	operator, err := workflowrun.NewOperator(h.Client, wfr, wfr.Namespace)
	if err != nil {
		log.WithField("wfr", wfr.Name).Error("Failed to create workflowrun operator: ", err)
		return
	}

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

	// If the WorkflowRun has already been terminated(Completed, Error, Cancel), send notifications if necessary,
	// otherwise directly skip it.
	if workflowrun.IsWorkflowRunTerminated(originWfr) {
		// Send notification as workflowrun has been terminated.
		sendNotifications(originWfr)
		return
	}

	// If the WorkflowRun has already been waiting for external events, skip it.
	if originWfr.Status.Overall.Phase == v1alpha1.StatusWaiting {
		return
	}

	wfr := originWfr.DeepCopy()
	operator, err := workflowrun.NewOperator(h.Client, wfr, wfr.Namespace)
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
	operator, err := workflowrun.NewOperator(h.Client, wfr, wfr.Namespace)
	if err != nil {
		log.WithField("wfr", wfr.Name).Error("Failed to create workflowrun operator: ", err)
		return
	}

	operator.GC(true, true)
	return
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

// sendNotifications send notifications for workflowruns.
// Will skip workflowruns have finished before workflow controller starts.
func sendNotifications(wfr *v1alpha1.WorkflowRun) {
	// No need to send notifications for workflowruns finished before workflow controller starts.
	if wfr.Status.Overall.LastTransitionTime.Before(controllerStartTime) {
		return
	}

	// Send notifications with workflowrun.
	bodyBytes, err := json.Marshal(wfr)
	if err != nil {
		log.WithField("wfr", wfr.Name).Error("Failed to marshal workflowrun: ", err)
		return
	}
	body := bytes.NewReader(bodyBytes)

	for _, endpoint := range controller.Config.Notifications {
		req, err := http.NewRequest(http.MethodPost, endpoint.URL, body)
		if err != nil {
			log.WithField("wfr", wfr.Name).Error("Failed to new notification request: ", err)
			continue
		}
		// Set Json content type in Http header.
		req.Header.Set(contentType, contentJSON)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.WithField("wfr", wfr.Name).Errorf("Failed to send notification for %s: %v", endpoint.Name, err)
			continue
		}
		log.WithField("wfr", wfr.Name).Infof("Status code of notification for %s: %d", endpoint.Name, resp.StatusCode)
	}
}
