package trigger

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

func getKubernetesClient() kubernetes.Interface {

	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		log.Fatalf("getClusterConfig: %v", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("getClusterConfig: %v", err)
	}
	log.Println("Successfully constructed k8s client")

	return client
}

func main() {

	client := getKubernetesClient()

	informer := cache.NewSharedIndexInformer(

		&cache.ListWatch{

			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Pods(meta_v1.NamespaceDefault).List(options)
			},

			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Pods(meta_v1.NamespaceDefault).Watch(options)
			},
		},

		&api_v1.Pod{},
		0,
		cache.Indexers{},
	)

	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), WorkflowTrigger)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{

		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			log.Printf("Add WorkflowTrigger: %s", key)
			if err == nil {
				queue.Add(key)
			}
		},

		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			log.Printf("Update WorkflowTrigger: %s", key)
			if err == nil {
				queue.Add(key)
			}
		},

		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			log.Printf("Delete WorkflowTrigger: %s", key)
			if err == nil {
				queue.Add(key)
			}
		},
	})

	controller := Controller{
		clientset: client,
		informer:  informer,
		queue:     queue,
		// handler:   &TestHandler{},
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	go controller.Run(stopCh)

	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, syscall.SIGTERM)
	signal.Notify(sigTerm, syscall.SIGINT)
	<-sigTerm
}
