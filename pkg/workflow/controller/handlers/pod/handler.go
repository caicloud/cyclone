package pod

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"
)

// Handler ...
type Handler struct {
	ClusterClient kubernetes.Interface
	Client        clientset.Interface
}

// Ensure *Handler has implemented handlers.Interface interface.
var _ handlers.Interface = (*Handler)(nil)

// Reconcile compares the actual state with the desired, and attempts to
// converge the two.
func (h *Handler) Reconcile(obj interface{}) error {
	// If Workflow Controller got restarted, previous started pods would be
	// observed by controller with create event. We need to handle update in
	// this case as well. Otherwise WorkflowRun may stuck in running state.
	return h.onUpdate(obj)
}

// ObjectDeleted ...
func (h *Handler) ObjectDeleted(obj interface{}) error {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		log.WithField("obj", obj).Warning("Expect Pod, got unknown type resource")
		return fmt.Errorf("unknown resource type")
	}
	log.WithField("name", pod.Name).Debug("Observed pod deleted")

	// Check whether it's GC pod.
	if IsGCPod(pod) {
		return nil
	}

	operator, err := NewOperator(h.ClusterClient, h.Client, pod)
	if err != nil {
		log.Error("Create operator error: ", err)
		return err
	}

	if err := operator.OnDelete(); err != nil {
		log.WithField("pod", pod.Name).Error("process deleted pod error: ", err)
		return err
	}
	return nil
}

func (h *Handler) onUpdate(obj interface{}) error {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		log.WithField("obj", obj).Warning("Expect Pod, got unknown type resource")
		return fmt.Errorf("unknown resource type")
	}
	log.WithField("name", pod.Name).Debug("Observed pod updated")

	// Check whether it's GC pod.
	if IsGCPod(pod) {
		GCPodUpdated(h.ClusterClient, pod)
		return nil
	}

	// For stage pod, create operator to handle it.
	operator, err := NewOperator(h.ClusterClient, h.Client, pod)
	if err != nil {
		log.Error("Create operator error: ", err)
		return err
	}

	if err := operator.OnUpdated(); err != nil {
		log.WithField("pod", pod.Name).Error("process updated pod error: ", err)
		return err
	}
	return nil
}
