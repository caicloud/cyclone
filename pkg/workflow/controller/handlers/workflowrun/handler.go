package workflowrun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/util"
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
	Informer              cache.SharedIndexInformer
}

// Ensure *Handler has implemented handlers.Interface interface.
var _ handlers.Interface = (*Handler)(nil)

const (
	// finalizerWorkflowRun is the cyclone related finalizer key for workflow run.
	finalizerWorkflowRun string = "workflowrun.cyclone.dev/finalizer"
)

// NewHandler ...
func NewHandler(client clientset.Interface, gcEnable bool, maxWorkflowRuns int, parallelism *controller.ParallelismConfig) *Handler {
	return &Handler{
		Client:                client,
		TimeoutProcessor:      workflowrun.NewTimeoutProcessor(client),
		GCProcessor:           workflowrun.NewGCProcessor(client, gcEnable),
		LimitedQueues:         workflowrun.NewLimitedQueues(client, maxWorkflowRuns),
		ParallelismController: workflowrun.NewParallelismController(parallelism),
	}
}

// Reconcile compares the actual state with the desired, and attempts to
// converge the two.
func (h *Handler) Reconcile(obj interface{}) error {
	originWfr, ok := obj.(*v1alpha1.WorkflowRun)
	if !ok {
		log.WithField("obj", obj).Warning("Expect WorkflowRun, got unknown type resource")
		return fmt.Errorf("unknown resource type")
	}

	if !validate(originWfr) {
		log.WithField("wfr", originWfr.Name).Warning("Invalid wfr")
		return fmt.Errorf("invalid workflowRun")
	}

	// Refresh updates 'refresh' time field of the WorkflowRun in the queue.
	h.LimitedQueues.AddOrRefresh(originWfr)

	// Add the WorkflowRun object to GC processor, it will be checked before actually added to
	// the GC queue.
	h.GCProcessor.Add(originWfr)

	// If the WorkflowRun has not yet be started to execute, check the parallelism constraints to determine
	// whether to execute it.
	if originWfr.Status.Overall.Phase == "" {
		log.WithField("wfr", originWfr.Name).Info("Attempt to run WorkflowRun")
		attemptAction := h.ParallelismController.AttemptNew(originWfr.Namespace, originWfr.Spec.WorkflowRef.Name, originWfr.Name)
		switch attemptAction {
		case workflowrun.AttemptActionQueued:
			log.WithField("wfr", originWfr.Name).Infof("Too many WorkflowRun are running, stay pending in queue, will retry in %s", common.ResyncPeriod.String())
			return fmt.Errorf("too many WorkflowRun are running")
		case workflowrun.AttemptActionFailed:
			if err := h.SetStatus(originWfr.Namespace, originWfr.Name, &v1alpha1.Status{
				Phase:              v1alpha1.StatusFailed,
				Reason:             "TooManyWaiting",
				LastTransitionTime: metav1.Time{Time: time.Now()},
			}); err != nil {
				log.WithField("wfr", originWfr.Name).Error("Set status to Failed error, ", err)
				return fmt.Errorf("too many WorkflowRun are running, and set status to Failed error")
			}
		}
		log.WithField("name", originWfr.Name).WithField("attemptAction", attemptAction).Debug("Attempt to run WorkflowRun")
	}

	log.WithField("name", originWfr.Name).Debug("Start to process WorkflowRun")

	// If the WorkflowRun has already been terminated(Completed, Failed, Cancelled), send notifications if necessary,
	// otherwise directly skip it.
	if util.IsWorkflowRunTerminated(originWfr) {
		// If the WorkflowRun already terminated, mark it in the ParallelismController.
		h.ParallelismController.MarkFinished(originWfr.Namespace, originWfr.Spec.WorkflowRef.Name, originWfr.Name)

		// Send notification after workflowrun terminated.
		err := h.sendNotification(originWfr)
		if err != nil {
			log.WithField("wfr", originWfr.Name).Warn("send notification failed", err)
		}
		return nil
	}

	// Add this WorkflowRun to timeout processor, so that it would be cleaned up when time expired.
	err := h.TimeoutProcessor.AddIfNotExist(originWfr)
	if err != nil {
		log.WithField("wfr", originWfr.Name).Warn("add wfr to timeout processor failed: ", err)
		return err
	}

	// If the WorkflowRun is waiting for external events, skip it.
	if originWfr.Status.Overall.Phase == v1alpha1.StatusWaiting {
		return nil
	}

	wfr := originWfr.DeepCopy()
	clusterClient := common.GetExecutionClusterClient(wfr)
	if clusterClient == nil {
		log.WithField("wfr", wfr.Name).Error("Execution cluster client not found")
		return fmt.Errorf("execution cluster client not found")
	}

	operator, err := workflowrun.NewOperator(clusterClient, h.Client, wfr, wfr.Namespace)
	if err != nil {
		log.WithField("wfr", wfr.Name).Error("Failed to create workflowrun operator: ", err)
		return err
	}

	operator.ResolveGlobalVariables()

	if err := operator.Reconcile(); err != nil {
		log.WithField("wfr", wfr.Name).Error("Reconcile error: ", err)
		return err
	}
	return nil
}

// finalize handles the case when a WorkflowRun get deleted.
// It will perform GC immediately for this WorkflowRun.
func (h *Handler) finalize(wfr *v1alpha1.WorkflowRun) error {
	clusterClient := common.GetExecutionClusterClient(wfr)
	if clusterClient == nil {
		log.WithField("wfr", wfr.Name).Error("Execution cluster client not found")
		return fmt.Errorf("Execution cluster client not found")
	}

	operator, err := workflowrun.NewOperator(clusterClient, h.Client, wfr.Name, wfr.Namespace)
	if err != nil {
		log.WithField("wfr", wfr.Name).Error("Failed to create workflowrun operator: ", err)
		return err
	}

	if err = operator.GC(true, true); err != nil {
		log.WithField("wfr", wfr.Name).Warn("GC failed", err)
		return err
	}
	return nil
}

// AddFinalizer adds a finalizer to the object and update the object to the Kubernetes.
func (h *Handler) AddFinalizer(obj interface{}) error {
	originWfr, ok := obj.(*v1alpha1.WorkflowRun)
	if !ok {
		log.WithField("obj", obj).Warning("Expect WorkflowRun, got unknown type resource")
		return fmt.Errorf("unknown resource type")
	}

	if sets.NewString(originWfr.Finalizers...).Has(finalizerWorkflowRun) {
		return nil
	}

	log.WithField("name", originWfr.Name).Debug("Start to add finalizer for workflowRun")

	wfr := originWfr.DeepCopy()
	wfr.ObjectMeta.Finalizers = append(wfr.ObjectMeta.Finalizers, finalizerWorkflowRun)
	_, err := h.Client.CycloneV1alpha1().WorkflowRuns(wfr.Namespace).Update(wfr)
	return err
}

// HandleFinalizer does the finalizer key representing things.
func (h *Handler) HandleFinalizer(obj interface{}) error {
	originWfr, ok := obj.(*v1alpha1.WorkflowRun)
	if !ok {
		log.WithField("obj", obj).Warning("Expect WorkflowRun, got unknown type resource")
		return fmt.Errorf("unknown resource type")
	}

	if !sets.NewString(originWfr.Finalizers...).Has(finalizerWorkflowRun) {
		return nil
	}

	log.WithField("name", originWfr.Name).Debug("Start to process finalizer for workflowRun")

	// Mark the WorkflowRun terminated in ParallelismController
	defer func() {
		h.ParallelismController.MarkFinished(originWfr.Namespace, originWfr.Name, originWfr.Spec.WorkflowRef.Name)
	}()

	// Handler finalizer
	wfr := originWfr.DeepCopy()
	if err := h.finalize(wfr); err != nil {
		return nil
	}

	wfr.ObjectMeta.Finalizers = sets.NewString(wfr.ObjectMeta.Finalizers...).Delete(finalizerWorkflowRun).UnsortedList()
	_, err := h.Client.CycloneV1alpha1().WorkflowRuns(wfr.Namespace).Update(wfr)
	return err
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
