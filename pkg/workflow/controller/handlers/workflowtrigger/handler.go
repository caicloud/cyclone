package workflowtrigger

import (
	log "github.com/sirupsen/logrus"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	finalizer "github.com/caicloud/cyclone/pkg/workflow/controller/finalizers"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"
)

// Handler ...
type Handler struct {
	CronManager *CronTriggerManager
	Finalizers  finalizer.Interface
}

var (
	// Check whether *Handler has implemented handlers.Interface interface.
	_ handlers.Interface = (*Handler)(nil)
)

// NewHandler ...
func NewHandler(client clientset.Interface) *Handler {
	return &Handler{
		CronManager: NewTriggerManager(client),
		Finalizers: finalizer.NewFinalizer(client, nil, updateFinalizer, appendFinalizer, removeFinalizer, map[string]finalizer.Handler{
			finalizerDeleteCronTask: handleFinalizerDeleteCronTask,
		}),
	}
}

// Reconcile compares the actual state with the desired, and attempts to
// converge the two.
func (h *Handler) Reconcile(obj interface{}) error {
	wft, err := ToWorkflowTrigger(obj)
	if err != nil {
		log.Warn("Convert to WorkflowTrigger error: ", err)
		return err
	}
	if wft.Spec.Type == v1alpha1.TriggerTypeCron {
		h.CronManager.UpdateCron(wft)
	}
	return nil
}

// ObjectDeleted ...
func (h *Handler) ObjectDeleted(obj interface{}) error {
	wft, err := ToWorkflowTrigger(obj)
	if err != nil {
		log.Warn("Convert to WorkflowTrigger error: ", err)
		return err
	}
	if wft.Spec.Type == v1alpha1.TriggerTypeCron {
		h.CronManager.DeleteCron(wft)
	}
	return nil
}
