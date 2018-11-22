package controllers

import (
	log "github.com/sirupsen/logrus"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers/workflowTrigger"
)

func NewWorkflowTriggerController(client clientset.Interface) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				return client.CycloneV1alpha1().WorkflowTriggers("").List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				return client.CycloneV1alpha1().WorkflowTriggers("").Watch(options)
			},
		},
		&v1alpha1.WorkflowTrigger{},
		0,
		cache.Indexers{},
	)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return
			}
			log.WithField("name", key).Debug("new WorkflowTrigger observed")
			queue.Add(Event{
				Key:          key,
				EventType:    CREATE,
				ResourceType: "wft",
				Object:       obj,
			})
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err != nil {
				return
			}
			log.WithField("name", key).Debug("WorkflowTrigger update observed")
			queue.Add(Event{
				Key:          key,
				EventType:    UPDATE,
				ResourceType: "wft",
				Object:       new,
			})
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return
			}
			log.WithField("name", key).Debug("deleting WorkflowTrigger")
			queue.Add(Event{
				Key:          key,
				EventType:    DELETE,
				ResourceType: "wft",
				Object:       obj,
			})
		},
	})

	return &Controller{
		name:      "WorkflowTrigger Controller",
		clientSet: client,
		informer:  informer,
		queue:     queue,
		eventHandler: &workflowTrigger.Handler{
			CronManager: workflowTrigger.NewTriggerManager(client),
		},
	}
}
