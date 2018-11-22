package workflowrun

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
)

// Check whether garbage collection is needed. We will return true to indicate
// garbage collection needed when:
// - The garbage collection hasn't been performed on this WorkflowRun yet.
// - The WorkflowRun has already been terminated,
// - The configured delay time is expired.
func checkGC(wfr *v1alpha1.WorkflowRun) bool {
	if wfr == nil {
		return false
	}

	// If it's already cleaned up, skip it.
	if wfr.Status.Cleaned {
		return false
	}

	// If it's not in terminated state, skip it.
	if wfr.Status.Overall.Status != v1alpha1.StatusCompleted &&
		wfr.Status.Overall.Status != v1alpha1.StatusError {
		return false
	}

	// If the configured delay time is not expired, skip it.
	if wfr.Status.Overall.LastTransitionTime.Add(time.Second * controller.Config.GC.DelaySeconds).After(time.Now()) {
		return false
	}

	return true
}

// GCProcessor processes garbage collection for WorkflowRun objects.
type GCProcessor struct {
	client   clientset.Interface
	items    map[string]*workflowRunItem
	enabled  bool
}

func NewGCProcessor(client clientset.Interface, enabled bool) *GCProcessor {
	processor := &GCProcessor{
		client: client,
		items: make(map[string]*workflowRunItem),
		enabled: enabled,
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
		name: wfr.Name,
		namespace: wfr.Namespace,
		expireTime: wfr.Status.Overall.LastTransitionTime.Time.Add(time.Second * controller.Config.GC.DelaySeconds),
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
		operator, err := NewOperator(p.client, i.name, i.namespace)
		if err != nil {
			log.WithField("wfr", i.name).Warn("Create operator for gc error: ", err)
			continue
		}
		if err = operator.GC(); err != nil {
			log.WithField("wfr", i.name).Warn("GC error: ", err)
			continue
		}

		delete(p.items, i.String())
	}
}