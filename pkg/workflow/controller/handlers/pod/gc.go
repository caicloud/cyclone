package pod

import (
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/caicloud/cyclone/pkg/meta"
)

// IsGCPod judges whether a pod is a GC pod by check whether it has
// label "gc.cyclone.dev".
func IsGCPod(pod *corev1.Pod) bool {
	if pod == nil || pod.Labels == nil {
		return false
	}

	if kind := pod.Labels[meta.LabelPodKind]; kind != meta.PodKindGC.String() {
		return false
	}

	return true
}

// IsWorkloadPod judges whether a pod is a workload pod by check whether it has
// label "workload.cyclone.dev".
func IsWorkloadPod(pod *corev1.Pod) bool {
	if pod == nil || pod.Labels == nil {
		return false
	}

	if kind := pod.Labels[meta.LabelPodKind]; kind != meta.PodKindWorkload.String() {
		return false
	}

	return true
}

// GCPodUpdated handles GC pod update. If GC pod is terminated, it will be deleted.
func GCPodUpdated(client kubernetes.Interface, pod *corev1.Pod) {
	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		if err := client.CoreV1().Pods(pod.Namespace).Delete(pod.Name, &metav1.DeleteOptions{}); err != nil {
			if errors.IsNotFound(err) {
				return
			}
			log.WithField("pod", pod.Name).Warn("Delete GC pod error: ", err)
		}
	}
}
