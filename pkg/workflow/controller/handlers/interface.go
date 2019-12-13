package handlers

// Interface ...
type Interface interface {
	// ObjectCreated handles object creation
	ObjectCreated(obj interface{}) error
	// ObjectUpdated handles object update
	ObjectUpdated(old, new interface{}) error
	// ObjectDeleted handles object deletion
	ObjectDeleted(obj interface{}) error
}
