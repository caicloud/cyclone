package controllers

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"
)

type Controller struct {
	name         string
	clientSet    clientset.Interface
	queue        workqueue.RateLimitingInterface
	informer     cache.SharedIndexInformer
	eventHandler handlers.Interface
}

type EventType int

const (
	CREATE EventType = iota
	UPDATE
	DELETE
)

type Event struct {
	Key          string
	EventType    EventType
	Object       interface{}
}

func (c *Controller) Run(stopCh <-chan struct{}) {
	defer c.queue.ShutDown()

	log.WithField("name", c.name).Info("Start controller.")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("timeout to sync caches"))
	}

	wait.Until(c.work, time.Second, stopCh)
}

func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

func (c *Controller) work() {
	for c.nextWork() {
	}
}

func (c *Controller) nextWork() bool {
	event, shutdown := c.queue.Get()
	if shutdown {
		return false
	}

	defer c.queue.Done(event)
	err := c.doWork(event.(Event))
	if err == nil {
		c.queue.Forget(event)
	} else if c.queue.NumRequeues(event) < 3 {
		log.Errorf("process %s failed (will retry): %v", event.(Event).Key, err)
		c.queue.AddRateLimited(event)
	} else {
		log.Errorf("process %s failed (gave up): %v", event.(Event).Key, err)
		c.queue.Forget(event)
		utilruntime.HandleError(err)
	}

	return true
}

func (c *Controller) doWork(e Event) error {
	switch e.EventType {
	case CREATE:
		c.eventHandler.ObjectCreated(e.Object)
	case UPDATE:
		c.eventHandler.ObjectUpdated(e.Object)
	case DELETE:
		c.eventHandler.ObjectDeleted(e.Object)
	}

	return nil
}
