package workflowTrigger

import (
	log "github.com/sirupsen/logrus"

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
	if wft, err := ToWorkflowTrigger(obj); err != nil {
		log.Warn("Convert to WorkflowTrigger error: ", err)
	} else {
		h.CronManager.CreateCron(wft)
	}
}

func (h *Handler) ObjectUpdated(obj interface{}) {
	if wft, err := ToWorkflowTrigger(obj); err != nil {
		log.Warn("Convert to WorkflowTrigger error: ", err)
	} else {
		h.CronManager.UpdateCron(wft)
	}
}

func (h *Handler) ObjectDeleted(obj interface{}) {
	if wft, err := ToWorkflowTrigger(obj); err != nil {
		log.Warn("Convert to WorkflowTrigger error: ", err)
	} else {
		h.CronManager.DeleteCron(wft)
	}
}
