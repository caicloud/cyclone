package controllers

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/caicloud/cyclone/pkg/util/k8s"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"
)

// Controller ...
type Controller struct {
	name          string
	clusterClient kubernetes.Interface
	clientSet     k8s.Interface
	queue         workqueue.RateLimitingInterface
	informer      cache.SharedIndexInformer
	eventHandler  handlers.Interface
}

// Run ...
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) {
	defer c.queue.ShutDown()

	log.WithField("name", c.name).WithField("threadiness", threadiness).Info("Start controller.")

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
	res, err := c.doWork(key.(string))
	if res.Requeue != nil {
		requeue := *res.Requeue
		if !requeue {
			log.Warningf("process %s failed (requeue=false, gave up)", key)
			c.queue.Forget(key)
		} else if res.RequeueAfter > 0 {
			log.Warningf("process %s failed (requeue=true, will retry after %s)", key, res.RequeueAfter)
			c.queue.Forget(key)
			c.queue.AddAfter(key, res.RequeueAfter)
		} else {
			log.Warningf("process %s failed (requeue=true, will retry)", key)
			c.queue.AddRateLimited(key)
		}
	} else if err != nil {
		if c.queue.NumRequeues(key) < 3 {
			log.Errorf("process %s failed (will retry): %v", key, err)
			c.queue.AddRateLimited(key)
		} else {
			log.Errorf("process %s failed (gave up): %v", key, err)
			c.queue.Forget(key)
			utilruntime.HandleError(err)
		}
	} else {
		c.queue.Forget(key)
	}

	return true
}

func (c *Controller) doWork(key string) (res controller.Result, err error) {
	obj, exists, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return res, fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	if obj == nil || !exists {
		log.WithField("obj", obj).WithField("exist", exists).Debug("Object is nil or not exist")
		return res, nil
	}

	object, ok := obj.(metav1.Object)
	if !ok {
		log.WithField("obj", obj).Warning("Expect it is a Kubernetes resource object, got unknown type resource")
		return res, fmt.Errorf("unknown resource type")
	}

	// The object deletion timestamp is not zero value that indicates the resource is being deleted
	if !object.GetDeletionTimestamp().IsZero() {
		return res, c.eventHandler.HandleFinalizer(object)
	}

	// Add finalizer if needed
	if err := c.eventHandler.AddFinalizer(object); err != nil {
		return res, err
	}

	return c.eventHandler.Reconcile(obj)
}
