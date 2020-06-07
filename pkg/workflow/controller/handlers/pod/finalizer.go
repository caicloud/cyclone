package pod

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	finalizer "github.com/caicloud/cyclone/pkg/workflow/controller/finalizers"
)

const (
	// finalizerSetStageStatus is the finalizer key representing updating stage status
	// to failed as the corresponding pod is being deleted.
	finalizerSetStageStatus string = "pod.cyclone.dev/finalizer-set-stage-status"
)

// NewFinalizer ...
func NewFinalizer(clusterClient kubernetes.Interface) finalizer.Interface {
	return finalizer.NewFinalizer(nil, clusterClient, updateFinalizer, appendFinalizer, removeFinalizer, map[string]finalizer.Handler{
		finalizerSetStageStatus: handleFinalizerSetStageStatus,
	})
}

// updateFinalizer updates the obj to the Kubernetes cluster
func updateFinalizer(_ clientset.Interface, clusterClient kubernetes.Interface, obj finalizer.RuntimeAndMetaInterface) error {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return fmt.Errorf("resource type not support")
	}
	_, err := clusterClient.CoreV1().Pods(pod.Namespace).Update(pod)
	return err
}

// appendFinalizer appends a finalizer to the obj
func appendFinalizer(obj finalizer.RuntimeAndMetaInterface, f string) (finalizer.RuntimeAndMetaInterface, error) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return obj, fmt.Errorf("resource type not support")
	}

	pod.ObjectMeta.Finalizers = append(pod.ObjectMeta.Finalizers, f)
	return pod, nil
}

// removeFinalizer removes a finalizer in the obj
func removeFinalizer(obj finalizer.RuntimeAndMetaInterface, f string) (finalizer.RuntimeAndMetaInterface, error) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return obj, fmt.Errorf("resource type not support")
	}

	pod.ObjectMeta.Finalizers = finalizer.RemoveString(pod.ObjectMeta.Finalizers, f)
	return pod, nil
}

// handleFinalizerSetStageStatus handles the set stage status finalizer
func handleFinalizerSetStageStatus(client clientset.Interface, clusterClient kubernetes.Interface, obj finalizer.RuntimeAndMetaInterface) error {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		log.WithField("obj", obj).Warning("Expect Pod, got unknown type resource")
		return fmt.Errorf("unknown resource type")
	}
	log.WithField("name", pod.Name).Debug("Observed pod is being deleted")

	// Check whether it's GC pod.
	if IsGCPod(pod) {
		return nil
	}

	operator, err := NewOperator(clusterClient, client, pod)
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
