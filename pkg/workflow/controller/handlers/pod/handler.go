package pod

import (
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

type Handler struct {
	Client clientset.Interface
}

// Ensure *Handler has implemented handlers.Interface interface.
var _ handlers.Interface = (*Handler)(nil)

func (h *Handler) ObjectCreated(obj interface{}) {
}

func (h *Handler) ObjectUpdated(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		log.Warning("unknown resource type")
		return
	}
	log.WithField("name", pod.Name).Debug("Observed pod updated")

	operator, err := NewOperator(h.Client, pod)
	if err != nil {
		log.Error("Create operator error: ", err)
		return
	}

	err = operator.OnUpdated()
	if err != nil {
		log.WithField("pod", pod.Name).Error("process updated pod error: ", err)
	}
}

func (h *Handler) ObjectDeleted(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		log.Warning("unknown resource type")
		return
	}
	log.WithField("name", pod.Name).Debug("Observed pod deleted")

	operator, err := NewOperator(h.Client, pod)
	if err != nil {
		log.Error("Create operator error: ", err)
		return
	}

	err = operator.OnDelete()
	if err != nil {
		log.WithField("pod", pod.Name).Error("process deleted pod error: ", err)
	}
}
