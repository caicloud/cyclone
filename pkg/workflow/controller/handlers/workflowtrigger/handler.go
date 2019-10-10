package workflowtrigger

import (
	log "github.com/sirupsen/logrus"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"
)

// Handler ...
type Handler struct {
	CronManager *CronTriggerManager
}

var (
	// Check whether *Handler has implemented handlers.Interface interface.
	_ handlers.Interface = (*Handler)(nil)
)

// ObjectCreated ...
func (h *Handler) ObjectCreated(obj interface{}) {
	if wft, err := ToWorkflowTrigger(obj); err != nil {
		log.Warn("Convert to WorkflowTrigger error: ", err)
	} else if wft.Spec.Type == v1alpha1.TriggerTypeCron {
		h.CronManager.CreateCron(wft)
	}
}

// ObjectUpdated ...
func (h *Handler) ObjectUpdated(obj interface{}) {
	if wft, err := ToWorkflowTrigger(obj); err != nil {
		log.Warn("Convert to WorkflowTrigger error: ", err)
	} else if wft.Spec.Type == v1alpha1.TriggerTypeCron {
		h.CronManager.UpdateCron(wft)
	}
}

// ObjectDeleted ...
func (h *Handler) ObjectDeleted(obj interface{}) {
	if wft, err := ToWorkflowTrigger(obj); err != nil {
		log.Warn("Convert to WorkflowTrigger error: ", err)
	} else if wft.Spec.Type == v1alpha1.TriggerTypeCron {
		h.CronManager.DeleteCron(wft)
	}
}
