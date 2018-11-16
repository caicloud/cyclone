package controllers

import (
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers/configmap"

	log "github.com/sirupsen/logrus"
	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

func NewConfigMapController(client clientset.Interface, namespace string, cm string) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().ConfigMaps(namespace).List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().ConfigMaps(namespace).Watch(options)
			},
		},
		&api_v1.ConfigMap{},
		0,
		cache.Indexers{},
	)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return
			}
			log.WithField("name", key).Debug("new configMap observed")
			queue.Add(Event{
				Key:          key,
				EventType:    CREATE,
				ResourceType: "cm",
				Object:       obj,
			})
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err != nil {
				return
			}
			log.WithField("name", key).Debug("configMap update observed")
			queue.Add(Event{
				Key:          key,
				EventType:    UPDATE,
				ResourceType: "cm",
				Object:       new,
			})
		},
	})

	return &Controller{
		name:      "ConfigMap Controller",
		clientSet: client,
		informer:  informer,
		queue:     queue,
		eventHandler: &configmap.Handler{
			Selectors: []configmap.Selector{
				configmap.Name(cm),
				configmap.Namespace(namespace)},
		},
	}
}
