package configmap

import (
	"github.com/caicloud/cyclone/pkg/workflow/controller"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"

	log "github.com/sirupsen/logrus"
	api_v1 "k8s.io/api/core/v1"
)

// Selector is a selector of ConfigMap, it defines the logic
// to judge whether a ConfigMap meet some conditions.
type Selector func(*api_v1.ConfigMap) bool

// Name defines a ConfigMapSelector who selects ConfigMap name.
func Name(name string) Selector {
	return func(cm *api_v1.ConfigMap) bool {
		return cm.Name == name
	}
}

// Name defines a ConfigMapSelector who selects ConfigMap namespace.
func Namespace(namespace string) Selector {
	return func(cm *api_v1.ConfigMap) bool {
		return cm.Namespace == namespace
	}
}

type Handler struct {
	// Selectors of ConfigMap, only those passed all the selectors
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
	cm, ok := obj.(*api_v1.ConfigMap)
	if !ok {
		log.Warning("unknown resource type")
		return
	}
	log.WithField("name", cm.Name).Debug("Start to process ConfigMap.")
	if !h.pass(cm) {
		return
	}

	// Reload config from this ConfigMap instance.
	log.WithField("name", cm.Name).Info("Start to reload config from ConfigMap")
	if err := controller.ReloadConfig(cm); err != nil {
		log.WithField("configMap", cm.Name).Errorf("reload config from ConfigMap error: %v", err)
	}
}

func (h *Handler) pass(cm *api_v1.ConfigMap) bool {
	for _, selector := range h.Selectors {
		if !selector(cm) {
			return false
		}
	}
	return true
}
