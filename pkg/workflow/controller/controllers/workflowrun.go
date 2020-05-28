package controllers

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/k8s/informers"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/workflow/common"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
	handlers "github.com/caicloud/cyclone/pkg/workflow/controller/handlers/workflowrun"
	"github.com/caicloud/cyclone/pkg/workflow/workflowrun"
)

// NewWorkflowRunController ...
func NewWorkflowRunController(client clientset.Interface) *Controller {
	drCollection := newDeletedResourceCollection()
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	factory := informers.NewSharedInformerFactoryWithOptions(
		client,
		common.ResyncPeriod,
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
			drCollection.Add(key, obj)
			queue.Add(key)
		},
	})

	return &Controller{
		name:         "WorkflowRun Controller",
		clientSet:    client,
		informer:     informer,
		queue:        queue,
		drCollection: drCollection,
		eventHandler: &handlers.Handler{
			Client:                client,
			TimeoutProcessor:      workflowrun.NewTimeoutProcessor(client),
			GCProcessor:           workflowrun.NewGCProcessor(client, controller.Config.GC.Enabled),
			LimitedQueues:         workflowrun.NewLimitedQueues(client, controller.Config.Limits.MaxWorkflowRuns),
			ParallelismController: workflowrun.NewParallelismController(controller.Config.Parallelism),
		},
	}
}
