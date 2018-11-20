package storageservice

import (
	"fmt"

	resv1b1 "github.com/caicloud/clientset/pkg/apis/resource/v1beta1"
	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	pkgruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/caicloud/storage-admin/pkg/constants"
	"github.com/caicloud/storage-admin/pkg/controller/common"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/util"
)

const (
	ControllerNameService = "StorageServiceController"
)

type Controller struct {
	*common.Framework
	kc  kubernetes.Interface              // control cluster client
	ccg kubernetes.ClusterClientsetGetter // user cluster clients

	tpCache *util.StorageTypeListWatchCache

	eventTrigger       chan string
	resyncPeriodSecond int

	maxRetryTimes int
}

func NewController(kc kubernetes.Interface, ccg kubernetes.ClusterClientsetGetter) (*Controller, error) {
	if controllerArgsChecker(kc, ccg) != nil {
		return nil, fmt.Errorf("can't get control clientset")
	}
	// create the pod watcher
	fieldSelector := fields.Everything()
	listWatcher := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (pkgruntime.Object, error) {
			options.FieldSelector = fieldSelector.String()
			return kc.ResourceV1beta1().StorageServices().List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.Watch = true
			options.FieldSelector = fieldSelector.String()
			return kc.ResourceV1beta1().StorageServices().Watch(options)
		},
	}

	// create the workqueue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the StorageService than the version which was responsible for triggering the update.
	indexer, informer := cache.NewIndexerInformer(listWatcher, &resv1b1.StorageService{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	}, cache.Indexers{})

	return NewControllerOpt(queue, indexer, informer, kc, ccg)
}
func NewControllerOpt(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller,
	kc kubernetes.Interface, ccg kubernetes.ClusterClientsetGetter) (*Controller, error) {
	if controllerArgsChecker(kc, ccg) != nil {
		return nil, fmt.Errorf("can't get control clientset")
	}
	c := new(Controller)
	c.Framework = common.NewControllerFramework(queue, indexer, informer, c)
	c.ccg = ccg
	c.kc = kc
	tpCache, e := util.NewStorageTypeListWatchCache(c.kubeClient)
	if e != nil {
		return nil, e
	}
	c.tpCache = tpCache
	c.maxRetryTimes = constants.DefaultControllerMaxRetryTimes
	c.resyncPeriodSecond = constants.DefaultControllerResyncPeriodSecond
	c.eventTrigger = make(chan string)
	return c, nil
}

func controllerArgsChecker(kc kubernetes.Interface, ccg kubernetes.ClusterClientsetGetter) error {
	if kc == nil || ccg == nil {
		return fmt.Errorf("can't get clientset")
	}
	return nil
}

func (c *Controller) kubeClient() kubernetes.Interface {
	return c.kc
}

func (c *Controller) Name() string {
	return ControllerNameService
}

func (c *Controller) Run(threadiness int, stopCh chan struct{}) {
	go c.tpCache.Run(stopCh)
	go c.loop(stopCh)
	c.Framework.Run(threadiness, stopCh)
}

func (c *Controller) InitCrd() error {
	if e := kubernetes.InitStorageAdminMetaMainCluster(c.kubeClient()); e != nil {
		return e
	}
	if e := kubernetes.InitDefaultStorageTypes(c.kubeClient()); e != nil {
		return e
	}
	return nil
}

// syncProcess is the business logic of the controller.
// In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (c *Controller) SyncProcess(key string) error {
	_, exists, err := c.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a StorageService, so that we will see a delete for one pod
		glog.Infof("StorageService %v has been real deleted", key)
		// don't rely on real delete, maybe miss
	}

	// Note that you also have to check the uid if you have a local controlled resource, which
	// is dependent on the actual instance, to detect that a Pod was recreated with the same name
	glog.Infof("Sync for StorageService %s", key)
	c.eventTrigger <- key
	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Controller) HandleError(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.GetQueue().Forget(key)
		return
	}

	// This controller retries c.maxRetryTimes times if something goes wrong. After that, it stops trying.
	if c.maxRetryTimes < 1 || c.GetQueue().NumRequeues(key) < c.maxRetryTimes {
		glog.Infof("Error syncing StorageService %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.GetQueue().AddRateLimited(key)
		return
	}

	c.GetQueue().Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	glog.Infof("Dropping StorageService %q out of the queue: %v", key, err)
}

func (c *Controller) SetMaxRetryTimes(maxRetryTimes int) {
	c.maxRetryTimes = maxRetryTimes
}

func (c *Controller) SetResyncPeriodSecond(resyncPeriodSecond int) {
	c.resyncPeriodSecond = resyncPeriodSecond
}

func RunStorageServiceExample(ec *Controller) { // for test
	// We can now warm up the cache for initial synchronization.
	// Let's suppose that we knew about a pod "mypod" on our last run, therefore add it to the cache.
	// If this pod is not there anymore, the controller will be notified about the removal after the
	// cache has synchronized.

	//ns := &resv1b1.StorageService{
	//	StorageService: storagev1.StorageService{
	//		ObjectMeta: metav1.ObjectMeta{
	//			Name: "storageclass-test",
	//		},
	//	},
	//	Status: resv1b1.StorageServiceStatus{
	//		Phase: resv1b1.StorageServicePending,
	//	},
	//}
	//ec.GetIndexer().Add(ns)

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go ec.Run(1, stop)

	// Wait forever
	select {}
}
