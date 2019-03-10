package workflowrun

import (
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/record"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/common"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
)

// GCProcessor processes garbage collection for WorkflowRun objects.
type GCProcessor struct {
	client   clientset.Interface
	recorder record.EventRecorder
	items    map[string]*workflowRunItem
	enabled  bool
}

// NewGCProcessor create new GC processor.
func NewGCProcessor(client clientset.Interface, enabled bool) *GCProcessor {
	processor := &GCProcessor{
		client:   client,
		recorder: common.GetEventRecorder(client, common.EventSourceWfrController),
		items:    make(map[string]*workflowRunItem),
		enabled:  enabled,
	}
	if enabled {
		go processor.run()
	}

	return processor
}

// Add WorkflowRun object to GC processor, it will firstly judge whether the WorkflowRun
// object needs GC, if it's true, it will perform GC on it in the right time.
func (p *GCProcessor) Add(wfr *v1alpha1.WorkflowRun) {
	if !p.enabled {
		return
	}

	if !checkGC(wfr) {
		return
	}

	item := &workflowRunItem{
		name:       wfr.Name,
		namespace:  wfr.Namespace,
		expireTime: wfr.Status.Overall.LastTransitionTime.Time.Add(time.Second * controller.Config.GC.DelaySeconds),
		retry:      controller.Config.GC.RetryCount,
	}
	p.items[item.String()] = item

	log.WithField("wfr", wfr.Name).
		WithField("gc_time", item.expireTime).
		Debug("Added to GCProcessor")
}

// Enable the processor and start it.
func (p *GCProcessor) Enable() {
	if p.enabled {
		return
	}

	p.enabled = true
	go p.run()
}

func (p *GCProcessor) run() {
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ticker.C:
			p.process()
		}
	}
}

func (p *GCProcessor) process() {
	var expired []*workflowRunItem
	for _, v := range p.items {
		if v.expireTime.Before(time.Now()) {
			expired = append(expired, v)
		}
	}

	for _, i := range expired {
		i.retry--
		log.WithField("wfr", i.name).Info("Start GC")
		operator, err := NewOperator(p.client, i.name, i.namespace)
		if err != nil {
			log.WithField("wfr", i.name).Warn("Create operator for gc error: ", err)
			if i.retry <= 0 {
				delete(p.items, i.String())
			}
			continue
		}
		if err = operator.GC(i.retry <= 0, false); err != nil {
			log.WithField("wfr", i.name).Warn("GC error: ", err)
			if i.retry <= 0 {
				delete(p.items, i.String())
			}
			continue
		}
		log.WithField("wfr", i.name).Info("GC succeeded")

		delete(p.items, i.String())
	}
}

// Check whether this WorkflowRun object is ready for GC, return true if:
// - The garbage collection hasn't been performed on this WorkflowRun yet.
// - The WorkflowRun has already been terminated,
func checkGC(wfr *v1alpha1.WorkflowRun) bool {
	if wfr == nil {
		return false
	}

	// If it's already cleaned up, skip it.
	if wfr.Status.Cleaned {
		return false
	}

	// If it's not in terminated state(Completed, Error, Cancel), skip it.
	if wfr.Status.Overall.Status != v1alpha1.StatusCompleted &&
		wfr.Status.Overall.Status != v1alpha1.StatusError &&
		wfr.Status.Overall.Status != v1alpha1.StatusCancelled {
		return false
	}

	return true
}
