package handlers

// Interface ...
type Interface interface {
	// ObjectDeleted handles object deletion
	ObjectDeleted(obj interface{}) error
	// Reconcile compares the actual state with the desired, and attempts to
	// converge the two.
	Reconcile(obj interface{}) error
}
