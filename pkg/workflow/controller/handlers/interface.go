package handlers

type Interface interface {
	ObjectCreated(obj interface{})
	ObjectUpdated(new interface{})
	ObjectDeleted(obj interface{})
}
