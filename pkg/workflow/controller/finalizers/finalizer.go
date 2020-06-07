package finalizer

import (
	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
)

// RuntimeAndMetaInterface wraps the runtime.Object and meta.Object
// In fact the runtime.Object is not used by now, but we defined it
// here to grammatically ensure users to pass the Kubernetes
// resource object directly instead of object.ObjectMeta
type RuntimeAndMetaInterface interface {
	runtime.Object
	metav1.Object
}

// Interface ...
type Interface interface {
	// IsBeingDeleted indicates the obj is being deleted
	IsBeingDeleted(obj metav1.Object) bool
	// AddFinalizersIfNotExist adds the finalizers to the object if it not exists,
	// This func will also update the object to Kubernetes.
	AddFinalizersIfNotExist(obj RuntimeAndMetaInterface) error
	// DoFinalize executes the corresponding handlers.
	DoFinalize(obj RuntimeAndMetaInterface) error
}

// Ensure *Finalizer has implemented Interface.
var _ Interface = (*Finalizer)(nil)

// Handler handles a corresponding finalizer
type Handler func(client clientset.Interface, clusterClient kubernetes.Interface, obj RuntimeAndMetaInterface) error

// Updater updates the kubernetes object
type Updater func(client clientset.Interface, clusterClient kubernetes.Interface, metaObj RuntimeAndMetaInterface) error

// Appender appendes f to the kubernetes resource object's ObjectMeta.Finalizers
type Appender func(object RuntimeAndMetaInterface, f string) (RuntimeAndMetaInterface, error)

// Remover removes f from the kubernetes resource object's ObjectMeta.Finalizers
type Remover func(RuntimeAndMetaInterface, string) (RuntimeAndMetaInterface, error)

// Finalizer implements the Interface
type Finalizer struct {
	update        Updater
	append        Appender
	remove        Remover
	handlers      map[string]Handler
	client        clientset.Interface
	clusterClient kubernetes.Interface
}

// NewFinalizer ...
func NewFinalizer(client clientset.Interface, clusterClient kubernetes.Interface, update Updater, append Appender, remove Remover, handlers map[string]Handler) Interface {
	return &Finalizer{
		client:        client,
		clusterClient: clusterClient,
		update:        update,
		append:        append,
		remove:        remove,
		handlers:      handlers,
	}
}

// IsBeingDeleted ...
func (f *Finalizer) IsBeingDeleted(obj metav1.Object) bool {
	return !obj.GetDeletionTimestamp().IsZero()
}

// AddFinalizersIfNotExist ...
func (f *Finalizer) AddFinalizersIfNotExist(obj RuntimeAndMetaInterface) error {
	exists := obj.GetFinalizers()
	for finalizer := range f.handlers {
		if !containsString(exists, finalizer) {
			var err error
			obj, err = f.append(obj, finalizer)
			if err != nil {
				return err
			}

		}
	}
	return f.update(f.client, f.clusterClient, obj)
}

// DoFinalize ...
func (f *Finalizer) DoFinalize(obj RuntimeAndMetaInterface) error {
	for _, finalizer := range obj.GetFinalizers() {
		if fn, ok := f.handlers[finalizer]; ok {
			if err := fn(f.client, f.clusterClient, obj); err != nil {
				return err
			} else {
				obj, err = f.remove(obj, finalizer)
				log.WithField("wfr", obj.GetName()).WithField("finalizers", obj.GetFinalizers()).Error("=== Test FinalizerRemove")
				if err != nil {
					return err
				}
			}
		}
	}

	return f.update(f.client, f.clusterClient, obj)
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// RemoveString is a helper func to remove a element in a slice
func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
