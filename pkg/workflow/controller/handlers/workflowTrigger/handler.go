package workflowTrigger

import (
	log "github.com/sirupsen/logrus"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"
)

type Handler struct {
	CronManager *CronTriggerManager
}

var (
	// Check whether *Handler has implemented handlers.Interface interface.
	_ handlers.Interface = (*Handler)(nil)
)

func (h *Handler) ObjectCreated(obj interface{}) {
	// h.process(obj)
	if wft, err := ToWorkflowTrigger(obj); err != nil {
		log.Warnln(err)
	} else {
		h.CronManager.CreateCron(wft)
	}
}

func (h *Handler) ObjectUpdated(obj interface{}) {
	// h.process(obj)
	if wft, err := ToWorkflowTrigger(obj); err != nil {
		log.Warnln(err)
	} else {
		h.CronManager.UpdateCron(wft)
	}
}

func (h *Handler) ObjectDeleted(obj interface{}) {
	// h.process(obj)
	if wft, err := ToWorkflowTrigger(obj); err != nil {
		log.Warnln(err)
	} else {
		h.CronManager.DeleteCron(wft)
	}
}

func (h *Handler) process(obj interface{}) {
	// Get ConfigMap instance and pass it through all selectors.
	wft, ok := obj.(*v1alpha1.WorkflowTrigger)
	if !ok {
		log.Warning("unknown resource type")
		return
	}
	log.WithField("name", wft.Name).Debug("Start to process WorkflowTrigger.")
	if !h.pass(wft) {
		return
	}
}

func (h *Handler) pass(wft *v1alpha1.WorkflowTrigger) bool {
	return true
}
