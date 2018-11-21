package workflowrun

import (
	log "github.com/sirupsen/logrus"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"
	"github.com/caicloud/cyclone/pkg/workflow/workflowrun"
)

// Handler handles changes of WorkflowRun CR.
type Handler struct {
	Client         clientset.Interface
	TimeoutProcessor *workflowrun.TimeoutProcessor
	GCProcessor    *workflowrun.GCProcessor
}

// Ensure *Handler has implemented handlers.Interface interface.
var _ handlers.Interface = (*Handler)(nil)

func (h *Handler) ObjectCreated(obj interface{}) {
	originWfr, ok := obj.(*v1alpha1.WorkflowRun)
	if !ok {
		log.Warning("unknown resource type")
		return
	}
	log.WithField("name", originWfr.Name).Debug("Start to process WorkflowRun.")

	// Add the WorkflowRun object to GC processor, it will be checked before actually added to
	// the GC queue.
	h.GCProcessor.Add(originWfr)

	// If the WorkflowRun has already been terminated or waiting for external events, skip it.
	if originWfr.Status.Overall.Status == v1alpha1.StatusCompleted ||
		originWfr.Status.Overall.Status == v1alpha1.StatusError ||
		originWfr.Status.Overall.Status == v1alpha1.StatusWaiting {
		return
	}

	// Add this WorkflowRun to timeout processor, so that it would be cleaned up when time exipred.
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

func (h *Handler) ObjectUpdated(obj interface{}) {
	originWfr, ok := obj.(*v1alpha1.WorkflowRun)
	if !ok {
		log.Warning("unknown resource type")
		return
	}
	log.WithField("name", originWfr.Name).Debug("Start to process WorkflowRun.")

	// Add the WorkflowRun object to GC processor, it will be checked before actually added to
	// the GC queue.
	h.GCProcessor.Add(originWfr)

	// If the WorkflowRun has already been terminated or waiting for external events, skip it.
	if originWfr.Status.Overall.Status == v1alpha1.StatusCompleted ||
		originWfr.Status.Overall.Status == v1alpha1.StatusError ||
		originWfr.Status.Overall.Status == v1alpha1.StatusWaiting {
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

func (h *Handler) ObjectDeleted(obj interface{}) {
	return
}