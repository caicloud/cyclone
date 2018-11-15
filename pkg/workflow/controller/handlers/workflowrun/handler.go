package workflowrun

import (
	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"

	log "github.com/sirupsen/logrus"
)

// Selector is a selector of WorkflowRun, it defines the logic
// to judge whether a WorkflowRun meet some conditions.
type Selector func(wfr *v1alpha1.WorkflowRun) bool

// Name defines a ConfigMapSelector who selects ConfigMap namespace.
func Namespace(namespace string) Selector {
	return func(wfr *v1alpha1.WorkflowRun) bool {
		return wfr.Namespace == namespace
	}
}

type Handler struct {
	// Selectors of WorkflowRun, only those passed all the selectors
	// would be processed by this handler.
	Selectors []Selector
}

// Check whether *Handler has implemented handlers.Interface interface.
var _ handlers.Interface = (*Handler)(nil)

func (h *Handler) ObjectCreated(obj interface{}) {
	h.process(obj)
}

func (h *Handler) ObjectUpdated(obj interface{}) {
	h.process(obj)
}

func (h *Handler) ObjectDeleted(obj interface{}) {
	return
}

func (h *Handler) process(obj interface{}) {
	// Get ConfigMap instance and pass it through all selectors.
	wfr, ok := obj.(*v1alpha1.WorkflowRun)
	if !ok {
		log.Warning("unknown resource type")
		return
	}
	log.WithField("name", wfr.Name).Debug("Start to process WorkflowRun.")
	if !h.pass(wfr) {
		return
	}
}

func (h *Handler) pass(cm *v1alpha1.WorkflowRun) bool {
	for _, selector := range h.Selectors {
		if !selector(cm) {
			return false
		}
	}
	return true
}
