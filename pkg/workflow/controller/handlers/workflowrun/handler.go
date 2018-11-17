package workflowrun

import (
	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"

	log "github.com/sirupsen/logrus"
)

// Handler handles changes of WorkflowRun CR.
type Handler struct {
	Client clientset.Interface
}

// Ensure *Handler has implemented handlers.Interface interface.
var _ handlers.Interface = (*Handler)(nil)

func (h *Handler) ObjectCreated(obj interface{}) {
	// TODO(ChenDe): Handle timeout and GC.

	h.process(obj)
}

func (h *Handler) ObjectUpdated(obj interface{}) {
	h.process(obj)
}

func (h *Handler) ObjectDeleted(obj interface{}) {
	// TODO(ChenDe)
	return
}

func (h *Handler) process(obj interface{}) {
	originWfr, ok := obj.(*v1alpha1.WorkflowRun)
	if !ok {
		log.Warning("unknown resource type")
		return
	}
	log.WithField("name", originWfr.Name).Debug("Start to process WorkflowRun.")

	// If the WorkflowRun has already been finished, skip it.
	if originWfr.Status.Overall.Status == v1alpha1.StatusCompleted ||
		originWfr.Status.Overall.Status == v1alpha1.StatusError {
		return
	}

	wfr := originWfr.DeepCopy()
	operator := NewOperator(h.Client)
	operator.Reconcile(wfr)
}
