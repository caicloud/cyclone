package util

import (
	"fmt"

	resv1b1 "github.com/caicloud/clientset/pkg/apis/resource/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	pkgruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/caicloud/storage-admin/pkg/kubernetes"
)

type ListWatchCache struct {
	indexer  cache.Indexer
	informer cache.Controller
}

func NewListWatchCache(listWatcher cache.ListerWatcher, objType runtime.Object) (*ListWatchCache, error) {
	if listWatcher == nil {
		return nil, fmt.Errorf("nil ListerWatcher for ListWatchCache")
	}
	if objType == nil {
		return nil, fmt.Errorf("nil runtime.Object for type")
	}
	indexer, informer := cache.NewIndexerInformer(listWatcher, objType, 0,
		cache.ResourceEventHandlerFuncs{}, cache.Indexers{})
	return &ListWatchCache{
		indexer:  indexer,
		informer: informer,
	}, nil
}

func (c *ListWatchCache) Run(stopCh chan struct{}) {
	defer utilruntime.HandleCrash()

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	<-stopCh
}

func (c *ListWatchCache) Get(key string) (item interface{}, exists bool, err error) {
	return c.indexer.GetByKey(key)
}

func (c *ListWatchCache) GetInNamespace(namespace, key string) (item interface{}, exists bool, err error) {
	return c.indexer.GetByKey(namespace + "/" + key)
}

type StorageTypeListWatchCache struct {
	ListWatchCache
	kcGetter func() kubernetes.Interface
}

func NewStorageTypeListWatchCache(kcGetter func() kubernetes.Interface) (*StorageTypeListWatchCache, error) {
	if kcGetter == nil {
		return nil, fmt.Errorf("nil kube clientset getter")
	}
	fieldSelector := fields.Everything()
	listWatcher := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (pkgruntime.Object, error) {
			options.FieldSelector = fieldSelector.String()
			kc := kcGetter()
			if kc == nil {
				return nil, fmt.Errorf("can't get kube client")
			}
			return kc.ResourceV1beta1().StorageTypes().List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.Watch = true
			options.FieldSelector = fieldSelector.String()
			kc := kcGetter()
			if kc == nil {
				return nil, fmt.Errorf("can't get kube client")
			}
			return kc.ResourceV1beta1().StorageTypes().Watch(options)
		},
	}
	c, e := NewListWatchCache(listWatcher, &resv1b1.StorageType{})
	if e != nil {
		return nil, e
	}
	return &StorageTypeListWatchCache{ListWatchCache: *c, kcGetter: kcGetter}, nil
}

func (c *StorageTypeListWatchCache) Get(name string) (*resv1b1.StorageType, error) {
	if obj, exist, e := c.ListWatchCache.Get(name); exist && obj != nil && e == nil {
		if tp, _ := obj.(*resv1b1.StorageType); tp != nil && tp.Name == name {
			return tp, nil
		}
	}
	kc := c.kcGetter()
	if kc == nil {
		return nil, fmt.Errorf("can't get kube client")
	}
	tp, e := kc.ResourceV1beta1().StorageTypes().Get(name, metav1.GetOptions{})
	if e != nil {
		return nil, e
	}
	return tp, nil
}
