package controller

import (
	"time"
)

// Result contains the result of a Reconciler invocation.
type Result struct {
	// Requeue tells the Controller to requeue the reconcile key.
	// If it is set to false, the controller will not requeue even if error occurred.
	// If it is nil, the controller will retry at most 3 times on errors.
	Requeue *bool

	// RequeueAfter if greater than 0, tells the Controller to requeue the reconcile key after the Duration.
	// Implies that Requeue is true, there is no need to set Requeue to true at the same time as RequeueAfter.
	RequeueAfter time.Duration
}
