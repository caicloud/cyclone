package controllers

import (
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/k8s/informers"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers/workflowtrigger"
)

func NewWorkflowTriggerController(client clientset.Interface) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	factory := informers.NewSharedInformerFactoryWithOptions(
		client,
		time.Second*30,
	)

	informer := factory.Cyclone().V1alpha1().WorkflowTriggers().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return
			}
			log.WithField("name", key).Debug("new WorkflowTrigger observed")
			queue.Add(Event{
				Key:       key,
				EventType: CREATE,
				Object:    obj,
			})
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err != nil {
				return
			}
			log.WithField("name", key).Debug("WorkflowTrigger update observed")
			queue.Add(Event{
				Key:       key,
				EventType: UPDATE,
				Object:    new,
			})
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return
			}
			log.WithField("name", key).Debug("deleting WorkflowTrigger")
			queue.Add(Event{
				Key:       key,
				EventType: DELETE,
				Object:    obj,
			})
		},
	})

	return &Controller{
		name:      "WorkflowTrigger Controller",
		clientSet: client,
		informer:  informer,
		queue:     queue,
		eventHandler: &workflowtrigger.Handler{
			CronManager: workflowtrigger.NewTriggerManager(client),
		},
	}
}
