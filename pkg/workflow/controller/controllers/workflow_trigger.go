package controllers

import (
	"reflect"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/k8s/informers"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/workflow/common"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers/workflowtrigger"
)

// NewWorkflowTriggerController ...
func NewWorkflowTriggerController(client clientset.Interface) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	factory := informers.NewSharedInformerFactoryWithOptions(
		client,
		common.ResyncPeriod,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = meta.WorkflowTriggerSelector()
		}),
	)

	informer := factory.Cyclone().V1alpha1().WorkflowTriggers().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return
			}
			log.WithField("name", key).Debug("new WorkflowTrigger observed")
			queue.Add(key)
		},
		UpdateFunc: func(old, new interface{}) {
			if reflect.DeepEqual(old, new) {
				return
			}
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err != nil {
				return
			}
			log.WithField("name", key).Debug("WorkflowTrigger update observed")
			queue.Add(key)
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return
			}
			log.WithField("name", key).Debug("deleting WorkflowTrigger")
			queue.Add(key)
		},
	})

	return &Controller{
		name:         "WorkflowTrigger Controller",
		clientSet:    client,
		informer:     informer,
		queue:        queue,
		eventHandler: workflowtrigger.NewHandler(client),
	}
}
