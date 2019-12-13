package configmap

import (
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

// ObjectCreated ...
func (h *Handler) ObjectCreated(obj interface{}) error {
	return h.process(obj)
}

// ObjectUpdated ...
func (h *Handler) ObjectUpdated(old, new interface{}) error {
	return h.process(new)
}

// ObjectDeleted ...
func (h *Handler) ObjectDeleted(obj interface{}) error {
	cm, ok := obj.(*api_v1.ConfigMap)
	if !ok {
		log.Warning("unknown resource type")
		return nil
	}

	log.WithField("name", cm.Name).Warn("Observed ConfigMap deletion.")
	return nil
}

func (h *Handler) process(obj interface{}) error {
	cm, ok := obj.(*api_v1.ConfigMap)
	if !ok {
		log.Warning("unknown resource type")
		return nil
	}
	log.WithField("name", cm.Name).Debug("Start to process ConfigMap.")

	// Reload config from this ConfigMap instance.
	log.WithField("name", cm.Name).Info("Start to reload config from ConfigMap")
	if err := controller.LoadConfig(cm); err != nil {
		log.WithField("configMap", cm.Name).Errorf("reload config from ConfigMap error: %v", err)
	}

	return nil
}
