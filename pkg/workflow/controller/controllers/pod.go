package controllers

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/k8s/informers"
	"github.com/caicloud/cyclone/pkg/workflow/common"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers/pod"
)

func NewPodController(client clientset.Interface) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	factory := informers.NewSharedInformerFactoryWithOptions(
		client,
		time.Second*30,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = common.PodLabelSelector
		}),
	)

	informer := factory.Core().V1().Pods().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return
			}
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
			queue.Add(Event{
				Key:       key,
				EventType: DELETE,
				Object:    obj,
			})
		},
	})

	return &Controller{
		name:      "Workflow Pod Controller",
		clientSet: client,
		informer:  informer,
		queue:     queue,
		eventHandler: &pod.Handler{
			Client: client,
		},
	}
}
