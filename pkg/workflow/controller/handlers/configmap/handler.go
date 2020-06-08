package configmap

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	api_v1 "k8s.io/api/core/v1"

	"github.com/caicloud/cyclone/pkg/workflow/controller"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"
)

// Handler ...
type Handler struct {
}

// Ensure *Handler has implemented handlers.Interface interface.
var _ handlers.Interface = (*Handler)(nil)

// Reconcile compares the actual state with the desired, and attempts to
// converge the two.
func (h *Handler) Reconcile(obj interface{}) error {
	return h.process(obj)
}

// ObjectDeleted ...
func (h *Handler) ObjectDeleted(obj interface{}) error {
	return nil
}

func (h *Handler) process(obj interface{}) error {
	cm, ok := obj.(*api_v1.ConfigMap)
	if !ok {
		log.WithField("obj", obj).Warning("Expect ConfigMap, got unknown type resource")
		return fmt.Errorf("unknown resource type")
	}
	log.WithField("name", cm.Name).Debug("Start to process ConfigMap.")

	// Reload config from this ConfigMap instance.
	log.WithField("name", cm.Name).Info("Start to reload config from ConfigMap")

	if err := controller.LoadConfig(cm); err != nil {
		log.WithField("configMap", cm.Name).Errorf("reload config from ConfigMap error: %v", err)
		return err
	}
	return nil
}

// AddFinalizer adds finalizers to the object and update the object to the Kubernetes.
func (h *Handler) AddFinalizer(obj interface{}) error {
	return nil
}

// HandleFinalizer does the finalizer key representing things.
func (h *Handler) HandleFinalizer(obj interface{}) error {
	return nil
}
