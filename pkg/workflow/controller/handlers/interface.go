package handlers

import (
	"github.com/caicloud/cyclone/pkg/workflow/controller"
)

// Interface ...
type Interface interface {
	// Reconcile compares the actual state with the desired, and attempts to
	// converge the two.
	Reconcile(obj interface{}) (controller.Result, error)

	// AddFinalizer adds a finalizer to the object if it not exists,
	// If a finalizer added, this func needs to update the object to the Kubernetes.
	AddFinalizer(obj interface{}) (added bool, err error)

	// HandleFinalizer needs to do things:
	// - execute the finalizer, like deleting any external resources associated with the obj
	// - remove the coorspending finalizer key from the obj
	// - update the object to the Kubernetes
	//
	// Ensure that this func must be idempotent and safe to invoke
	// multiple types for same object.
	HandleFinalizer(obj interface{}) error
}
