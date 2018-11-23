package controllers

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/k8s/informers"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers/configmap"
)

func NewConfigMapController(client clientset.Interface, namespace string, cm string) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	factory := informers.NewSharedInformerFactoryWithOptions(
		client,
		time.Second * 30,
		informers.WithNamespace(namespace),
		informers.WithTweakListOptions(func(options *metav1.ListOptions){
			options.FieldSelector = fmt.Sprintf("metadata.name==%s", cm)
		}),
	)

	informer := factory.Core().V1().ConfigMaps().Informer()
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
				Object:       new,
			})
		},
	})

	return &Controller{
		name:         "ConfigMap Controller",
		clientSet:    client,
		informer:     informer,
		queue:        queue,
		eventHandler: &configmap.Handler{},
	}
}
