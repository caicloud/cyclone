package handlers

// Interface ...
type Interface interface {
	// // ObjectCreated handles object creation
	// ObjectCreated(obj interface{})
	// // ObjectUpdated handles object update
	// ObjectUpdated(old, new interface{})
	// ObjectDeleted handles object deletion
	ObjectDeleted(obj interface{}) error
	//
	Reconcile(obj interface{}) error
}
