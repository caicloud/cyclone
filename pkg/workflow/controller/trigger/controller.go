package trigger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/tskdsb/cyclone/pkg/apis/cyclone/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/robfig/cron"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	ContentTypeJSON string = "application/json"

	ServerHostPort string = ""
	RunWorkflowAPI string = "/workflows/%s/workflowruns"

	WorkflowTrigger  = "WorkflowTrigger"
	WorkflowRun      = "WorkflowRun"
	NamespaceDefault = "default"
)

var ScheduleTaskS map[v1alpha1.WorkflowTrigger]*ScheduleTask

func CreateTrigger(obj interface{}) {
	wft, ok := obj.(v1alpha1.WorkflowTrigger)
	if !ok {
		log.Printf("type shoud be %s\n", WorkflowTrigger)
		return
	}
	st, err := AddScheduleTask(wft)
	if err != nil {
		log.Printf("create %s failed: %s\n", WorkflowTrigger, err)
	} else {
		ScheduleTaskS[wft] = st
	}
}

func UpdateTrigger(oldObj, newObj interface{}) {

}

func DeleteTrigger(obj interface{}) {
	wft, ok := obj.(v1alpha1.WorkflowTrigger)
	if !ok {
		log.Printf("type shoud be %s\n", WorkflowTrigger)
		return
	}
	ScheduleTaskS[wft].Cron.Stop()
}

func object2reader(i interface{}) (io.Reader, error) {
	byteS, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(byteS), nil
}

type ScheduleTask struct {
	Cron            *cron.Cron
	SuccCount       int
	FailCount       int
	WorkflowTrigger v1alpha1.WorkflowTrigger
	WorkflowRun     v1alpha1.WorkflowRun
}

func (st ScheduleTask) Run() {

	workflowRun := v1alpha1.WorkflowRun{
		TypeMeta: v1.TypeMeta{
			Kind:       WorkflowRun,
			APIVersion: "",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      st.WorkflowTrigger.Spec.WorkflowRef.Name,
			Namespace: NamespaceDefault,
		},
		Spec: v1alpha1.WorkflowRunSpec{
			StartStages:    st.WorkflowTrigger.Spec.StartStages,
			EndStages:      st.WorkflowTrigger.Spec.EndStages,
			ServiceAccount: st.WorkflowTrigger.Spec.ServiceAccount,
			Resources:      st.WorkflowTrigger.Spec.Resources,
			Stages:         st.WorkflowTrigger.Spec.Stages,
		},
	}

	body, err := object2reader(workflowRun)
	if err != nil {
		log.Println(err)
		return
	}

	resp, err := http.Post(fmt.Sprintf(ServerHostPort+RunWorkflowAPI, workflowRun), ContentTypeJSON, body)
	if err != nil {
		st.FailCount++
		log.Println(err)
		return
	} else {
		st.SuccCount++
		st.WorkflowRun = workflowRun
	}

	respMsg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("%s", respMsg)
	}
}

func AddScheduleTask(workflowTrigger v1alpha1.WorkflowTrigger) (st *ScheduleTask, err error) {

	c := cron.New()
	err = c.AddJob(workflowTrigger.Spec.Schedule, st)
	if err != nil {
		return
	}
	st = &ScheduleTask{
		WorkflowTrigger: workflowTrigger,
		Cron:            c,
	}
	st.Cron.Start()
	return
}

type Controller struct {
	// logger    *log.Entry
	clientset kubernetes.Interface
	queue     workqueue.RateLimitingInterface
	informer  cache.SharedIndexInformer
	// handler   Handler
}

func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	log.Println("Controller.Run: initiating")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Error syncing cache"))
		return
	}
	log.Println("Controller.Run: cache sync complete")

	wait.Until(c.runWorker, time.Second, stopCh)
}

func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

func (c *Controller) runWorker() {
	log.Println("Controller.runWorker: starting")

	for c.processNextItem() {
		log.Println("Controller.runWorker: processing next item")
	}

	log.Println("Controller.runWorker: completed")
}

func (c *Controller) processNextItem() bool {

	log.Println("Controller.processNextItem: start")

	key, quit := c.queue.Get()

	if quit {
		return false
	}

	defer c.queue.Done(key)

	keyRaw := key.(string)

	item, exists, err := c.informer.GetIndexer().GetByKey(keyRaw)
	if err != nil {
		if c.queue.NumRequeues(key) < 5 {
			log.Printf("Controller.processNextItem: Failed processing item with key %s with error %v, retrying", key, err)
			c.queue.AddRateLimited(key)
		} else {
			log.Printf("Controller.processNextItem: Failed processing item with key %s with error %v, no more retries", key, err)
			c.queue.Forget(key)
			utilruntime.HandleError(err)
		}
	}

	if !exists {
		log.Printf("Controller.processNextItem: object deleted detected: %s", keyRaw)
		DeleteTrigger(item)
		c.queue.Forget(key)
	} else {
		log.Printf("Controller.processNextItem: object created detected: %s", keyRaw)
		CreateTrigger(item)
		c.queue.Forget(key)
	}

	return true
}
