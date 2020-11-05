package controllers

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/k8s/informers"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
	handlers "github.com/caicloud/cyclone/pkg/workflow/controller/handlers/workflowrun"
)

// NewWorkflowRunController ...
func NewWorkflowRunController(client clientset.Interface) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	factory := informers.NewSharedInformerFactoryWithOptions(
		client,
		controller.Config.ResyncPeriodSeconds*time.Second,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = meta.WorkflowRunSelector()
		}),
	)

	informer := factory.Cyclone().V1alpha1().WorkflowRuns().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return
			}
			queue.Add(key)
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err != nil {
				return
			}
			queue.Add(key)
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return
			}
			queue.Add(key)
		},
	})

	return &Controller{
		name:         "WorkflowRun Controller",
		clientSet:    client,
		informer:     informer,
		queue:        queue,
		eventHandler: handlers.NewHandler(client, controller.Config.GC.Enabled, controller.Config.Limits.MaxWorkflowRuns, controller.Config.Parallelism),
	}
}
