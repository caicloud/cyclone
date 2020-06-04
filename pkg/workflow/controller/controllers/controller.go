package controllers

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	log "github.com/sirupsen/logrus"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"
)

// Controller ...
type Controller struct {
	name          string
	clusterClient kubernetes.Interface
	clientSet     clientset.Interface
	queue         workqueue.RateLimitingInterface
	drCollection  deletedResourceCollectionInterface
	informer      cache.SharedIndexInformer
	eventHandler  handlers.Interface
}

// EventType ...
type EventType int

const (
	// CREATE indicates creation event
	CREATE EventType = iota
	// UPDATE indicates updating event
	UPDATE
	// DELETE indicates deletion event
	DELETE
)

// Event ...
type Event struct {
	Key       string
	EventType EventType
	Object    interface{}
	OldObject interface{}
}

// Run ...
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) {
	defer c.queue.ShutDown()

	log.WithField("name", c.name).Info("Start controller.")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("timeout to sync caches"))
	}

	log.Infof("Cyclone controller %s synced and ready", c.name)

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.work, time.Second, stopCh)
	}

	<-stopCh
	glog.Infof("Shutting down %s controller", c.name)
}

// HasSynced ...
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

func (c *Controller) work() {
	for c.nextWork() {
	}
}

func (c *Controller) nextWork() bool {
	key, shutdown := c.queue.Get()
	if shutdown {
		return false
	}

	defer c.queue.Done(key)
	err := c.doWork(key.(string))
	if err == nil {
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < 3 {
		log.Errorf("process %s failed (will retry): %v", key, err)
		c.queue.AddRateLimited(key)
	} else {
		log.Errorf("process %s failed (gave up): %v", key, err)
		c.queue.Forget(key)
		utilruntime.HandleError(err)
	}

	return true
}

func (c *Controller) doWork(key string) error {
	obj, exists, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	if !exists {
		deleteObj, err := c.drCollection.Get(key)
		if err != nil {
			log.WithField("key", key).Error("Deleted object lost")
			return nil
		}
		if err := c.eventHandler.ObjectDeleted(deleteObj); err != nil {
			return err
		}
		c.drCollection.Remove(key)
		return nil
	}
	return c.eventHandler.Reconcile(obj)
}
