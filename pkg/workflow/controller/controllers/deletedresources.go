package controllers

import (
	"fmt"
	"sync"
)

type deletedResourceCollectionInterface interface {
	Add(key string, obj interface{})
	Remove(key string)
	Get(key string) (interface{}, error)
}

type deletedResourceCollection struct {
	collection map[string]interface{}
	sync.Mutex
}

func newDeletedResourceCollection() deletedResourceCollectionInterface {
	return &deletedResourceCollection{
		collection: make(map[string]interface{}),
	}
}

// Add ...
func (c *deletedResourceCollection) Add(key string, obj interface{}) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.collection[key] = obj
}

// Remove ...
func (c *deletedResourceCollection) Remove(key string) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	delete(c.collection, key)
}

// Get ...
func (c *deletedResourceCollection) Get(key string) (interface{}, error) {
	if obj, ok := c.collection[key]; ok {
		return obj, nil
	}
	return nil, fmt.Errorf("object %s not exist", key)
}
