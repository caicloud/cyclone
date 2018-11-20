package common

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Handler interface {
	Name() string
	SyncProcess(key string) error
	HandleError(err error, key interface{})
}

type Framework struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller

	handler Handler
}

func (c *Framework) GetIndexer() cache.Indexer                 { return c.indexer }
func (c *Framework) GetQueue() workqueue.RateLimitingInterface { return c.queue }
func (c *Framework) GetInformer() cache.Controller             { return c.informer }

func NewControllerFramework(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller,
	handler Handler) *Framework {
	return &Framework{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
		handler:  handler,
	}
}

func (c *Framework) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	err := c.handler.SyncProcess(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	c.handler.HandleError(err, key)
	return true
}

func (c *Framework) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	glog.Infof("Starting %s controller", c.handler.Name())

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	glog.Infof("Stopping %s controller", c.handler.Name())
}

func (c *Framework) runWorker() {
	for c.processNextItem() {
	}
}
