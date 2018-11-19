package controllers

import (
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers/pod"

	corev1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

func NewPodController(client clientset.Interface) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				options.LabelSelector = workflow.PodLabelSelector
				return client.CoreV1().Pods("").List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				options.LabelSelector = workflow.PodLabelSelector
				return client.CoreV1().Pods("").Watch(options)
			},
		},
		&corev1.Pod{},
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
				ResourceType: "pod",
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
				ResourceType: "pod",
				Object:       new,
			})
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return
			}
			queue.Add(Event{
				Key:          key,
				EventType:    DELETE,
				ResourceType: "pod",
				Object:       obj,
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
