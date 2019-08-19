package usage

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/meta"
)

// Recorder is an interface used to record some information
type Recorder interface {
	// RecordWatcherResource records resource requirements of pvc watcher
	// Create record if it not exists; Update record if it exists.
	RecordWatcherResource(corev1.ResourceRequirements) error

	// GetWatcherResource returns resource requirements of pvc watcher
	GetWatcherResource() (*corev1.ResourceRequirements, error)
}

// NamespaceRecorder is an recorder recoreds namespace information.
type NamespaceRecorder struct {
	namespace string
	client    kubernetes.Interface
}

// NewNamespaceRecorder creates a namespace recorder
func NewNamespaceRecorder(client kubernetes.Interface, namespace string) *NamespaceRecorder {
	return &NamespaceRecorder{
		namespace: namespace,
		client:    client,
	}
}

// RecordWatcherResource records resource requirements of pvc watcher
func (r *NamespaceRecorder) RecordWatcherResource(resource corev1.ResourceRequirements) error {
	resourceRequirements, err := json.Marshal(resource)
	if err != nil {
		return err
	}

	// update namespace annotation with retry
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := r.client.CoreV1().Namespaces().Get(r.namespace, metav1.GetOptions{})
		if err != nil {
			return err
		}

		namespace := origin.DeepCopy()

		if namespace.Annotations == nil {
			namespace.Annotations = make(map[string]string)
		}
		namespace.Annotations[meta.AnnotationPVCWatcherResourceRequirements] = string(resourceRequirements)

		_, err = r.client.CoreV1().Namespaces().Update(namespace)
		if err != nil {
			return err
		}
		return nil
	})
}

// GetWatcherResource gets resource requirements of pvc watcher
func (r *NamespaceRecorder) GetWatcherResource() (*corev1.ResourceRequirements, error) {
	namespace, err := r.client.CoreV1().Namespaces().Get(r.namespace, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if namespace.Annotations == nil {
		return nil, fmt.Errorf("Annotations of namespace %s is nil", namespace.Name)
	}
	resourceRequirements, ok := namespace.Annotations[meta.AnnotationPVCWatcherResourceRequirements]
	if !ok {
		return nil, fmt.Errorf("Annotations %s of namespace %s does not exist", meta.AnnotationPVCWatcherResourceRequirements, namespace.Name)
	}

	watcherResource := &corev1.ResourceRequirements{}
	err = json.Unmarshal([]byte(resourceRequirements), watcherResource)
	return watcherResource, err
}
