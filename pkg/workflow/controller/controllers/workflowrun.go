package controllers

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
	handlers "github.com/caicloud/cyclone/pkg/workflow/controller/handlers/workflowrun"
	"github.com/caicloud/cyclone/pkg/workflow/workflowrun"
)

func NewWorkflowRunController(client clientset.Interface) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				return client.CycloneV1alpha1().WorkflowRuns("").List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				return client.CycloneV1alpha1().WorkflowRuns("").Watch(options)
			},
		},
		&v1alpha1.WorkflowRun{},
		0,
		cache.Indexers{},
	)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return
			}
			queue.Add(Event{
				Key:          key,
				EventType:    CREATE,
				ResourceType: "wfr",
				Object:       obj,
			})
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err != nil {
				return
			}
			queue.Add(Event{
				Key:          key,
				EventType:    UPDATE,
				ResourceType: "wfr",
				Object:       new,
			})
		},
	})

	return &Controller{
		name:      "WorkflowRun Controller",
		clientSet: client,
		informer:  informer,
		queue:     queue,
		eventHandler: &handlers.Handler{
			Client:         client,
			TimeoutProcessor: workflowrun.NewTimeoutProcessor(client),
			GCProcessor: workflowrun.NewGCProcessor(client, controller.Config.GC.Enabled),
		},
	}
}
