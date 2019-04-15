package workflowrun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

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
	if originWfr.Status.Overall.Phase == v1alpha1.StatusSucceeded ||
		originWfr.Status.Overall.Phase == v1alpha1.StatusFailed ||
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

	// If the WorkflowRun has already been terminated(Completed, Failed, Cancelled), send notifications if necessary,
	// otherwise directly skip it.
	if workflowrun.IsWorkflowRunTerminated(originWfr) {
		// Send notification after workflowrun terminated.
		status, err := h.sendNotifications(originWfr)
		if err != nil {
			log.WithField("name", originWfr.Name).Error("Send notification error: ", err)
			return
		}
		// If notification status is nil, then no notification is sent.
		if status == nil {
			return
		}

		// Update WorkflowRun notification status with retry.
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			// Get latest WorkflowRun.
			latest, err := h.Client.CycloneV1alpha1().WorkflowRuns(originWfr.Namespace).Get(originWfr.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			if latest.Status.Notifications == nil {
				latest.Status.Notifications = status
				_, err = h.Client.CycloneV1alpha1().WorkflowRuns(originWfr.Namespace).Update(latest)
				return err
			}

			return nil
		})

		if err != nil {
			log.WithField("name", originWfr.Name).Error("Update workflowrun notification status error: ", err)
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

// sendNotifications send notifications for workflowruns when:
// * its workflow has notification config
// * finish time after workflow controller starts
// * notification status of workflowrun is nil
// If the returned notification status is nil, it means that there is no need to send notification.
func (h *Handler) sendNotifications(wfr *v1alpha1.WorkflowRun) (map[string]v1alpha1.NotificationStatus, error) {
	if wfr.Status.Notifications != nil ||
		wfr.Status.Overall.LastTransitionTime.Before(controllerStartTime) {
		return nil, nil
	}

	wfRef := wfr.Spec.WorkflowRef
	if wfRef == nil {
		return nil, fmt.Errorf("Workflow reference of workflow run %s/%s is empty", wfr.Namespace, wfr.Name)
	}
	wf, err := h.Client.CycloneV1alpha1().Workflows(wfRef.Namespace).Get(wfRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if len(wf.Spec.Notification.Receivers) == 0 {
		return nil, nil
	}

	// Send notifications with workflowrun.
	bodyBytes, err := json.Marshal(wfr)
	if err != nil {
		log.WithField("wfr", wfr.Name).Error("Failed to marshal workflowrun: ", err)
		return nil, err
	}
	body := bytes.NewReader(bodyBytes)

	status := make(map[string]v1alpha1.NotificationStatus)
	for _, endpoint := range controller.Config.Notifications {
		req, err := http.NewRequest(http.MethodPost, endpoint.URL, body)
		if err != nil {
			err = fmt.Errorf("failed to new notification request: %v", err)
			log.WithField("wfr", wfr.Name).Error(err)
			status[endpoint.Name] = v1alpha1.NotificationStatus{
				Result:  v1alpha1.NotificationResultFailed,
				Message: err.Error(),
			}
			continue
		}
		// Set Json content type in Http header.
		req.Header.Set(contentType, contentJSON)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			s := v1alpha1.NotificationStatus{
				Result: v1alpha1.NotificationResultFailed,
			}

			log.WithField("wfr", wfr.Name).Errorf("Failed to send notification for %s: %v", endpoint.Name, err)
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error(err)
				s.Message = err.Error()
			} else {
				s.Message = fmt.Sprintf("Status code: %d, error: %s", resp.StatusCode, body)
			}

			status[endpoint.Name] = s
			continue
		}

		log.WithField("wfr", wfr.Name).Infof("Status code of notification for %s: %d", endpoint.Name, resp.StatusCode)
		status[endpoint.Name] = v1alpha1.NotificationStatus{
			Result:  v1alpha1.NotificationResultSucceeded,
			Message: fmt.Sprintf("Status code: %d", resp.StatusCode),
		}
	}

	return status, nil
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
