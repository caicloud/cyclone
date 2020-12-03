package pod

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"

	"github.com/caicloud/cyclone/pkg/util/k8s"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"
)

// Handler ...
type Handler struct {
	ClusterClient kubernetes.Interface
	Client        k8s.Interface
}

// Ensure *Handler has implemented handlers.Interface interface.
var _ handlers.Interface = (*Handler)(nil)

const (
	// finalizerPod is the cyclone related finalizer key for kubernetes pod.
	finalizerPod string = "pod.cyclone.dev/finalizer"
)

// NewHandler ...
func NewHandler(client k8s.Interface, clusterClient kubernetes.Interface) *Handler {
	return &Handler{
		Client:        client,
		ClusterClient: clusterClient,
	}
}

// Reconcile compares the actual state with the desired, and attempts to
// converge the two.
func (h *Handler) Reconcile(obj interface{}) (res controller.Result, err error) {
	// If Workflow Controller got restarted, previous started pods would be
	// observed by controller with create event. We need to handle update in
	// this case as well. Otherwise WorkflowRun may stuck in running state.

	originPod, ok := obj.(*corev1.Pod)
	if !ok {
		log.WithField("obj", obj).Warning("Expect Pod, got unknown type resource")
		return res, fmt.Errorf("unknown resource type")
	}

	pod := originPod.DeepCopy()
	return res, h.onUpdate(pod)
}

func (h *Handler) onUpdate(pod *corev1.Pod) error {
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

// finalize ...
func (h *Handler) finalize(pod *corev1.Pod) error {
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

// AddFinalizer adds a finalizer to the object and update the object to the Kubernetes.
func (h *Handler) AddFinalizer(obj interface{}) error {
	originPod, ok := obj.(*corev1.Pod)
	if !ok {
		log.WithField("obj", obj).Warning("Expect Pod, got unknown type resource")
		return fmt.Errorf("unknown resource type")
	}

	// Check whether it's workload pod.
	if !IsWorkloadPod(originPod) {
		return nil
	}

	if sets.NewString(originPod.Finalizers...).Has(finalizerPod) {
		return nil
	}

	log.WithField("name", originPod.Name).Debug("Start to add finalizer for pod")

	pod := originPod.DeepCopy()
	pod.ObjectMeta.Finalizers = append(pod.ObjectMeta.Finalizers, finalizerPod)
	_, err := h.ClusterClient.CoreV1().Pods(pod.Namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	return err
}

// HandleFinalizer does the finalizer key representing things.
func (h *Handler) HandleFinalizer(obj interface{}) error {
	originPod, ok := obj.(*corev1.Pod)
	if !ok {
		log.WithField("obj", obj).Warning("Expect Pod, got unknown type resource")
		return fmt.Errorf("unknown resource type")
	}

	// Check whether it's workload pod.
	if !IsWorkloadPod(originPod) {
		return nil
	}

	if !sets.NewString(originPod.Finalizers...).Has(finalizerPod) {
		return nil
	}

	log.WithField("name", originPod.Name).Debug("Start to process finalizer for pod")

	// Handler finalizer
	pod := originPod.DeepCopy()
	if err := h.finalize(pod); err != nil {
		return nil
	}

	pod.ObjectMeta.Finalizers = sets.NewString(pod.ObjectMeta.Finalizers...).Delete(finalizerPod).UnsortedList()
	_, err := h.ClusterClient.CoreV1().Pods(pod.Namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	return err
}
