package controllers

import (
	"reflect"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers/pod"
)

// NewPodController ...
func NewPodController(clusterClient kubernetes.Interface, client clientset.Interface) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	factory := informers.NewSharedInformerFactoryWithOptions(
		clusterClient,
		controller.Config.ResyncPeriodSeconds*time.Second,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = meta.CyclonePodSelector()
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
			if reflect.DeepEqual(old, new) {
				return
			}
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err != nil {
				return
			}
			queue.Add(Event{
				Key:       key,
				EventType: UPDATE,
				Object:    new,
				OldObject: old,
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
		name:          "Workflow Pod Controller",
		clientSet:     client,
		clusterClient: clusterClient,
		informer:      informer,
		queue:         queue,
		eventHandler: &pod.Handler{
			ClusterClient: clusterClient,
			Client:        client,
		},
	}
}
