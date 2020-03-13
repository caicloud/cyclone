package workflowrun

import (
	"fmt"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8swatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
)

// PodEventWatcher watches Kubernetes events for a pod and records Warning type events to WorkflowRun status.
type PodEventWatcher interface {
	// Invoke Work in background goroutine, this is a blocking method.
	Work(stage, podNamespace, podName string)
}

type podEventWatcher struct {
	clusterClient   kubernetes.Interface
	client          clientset.Interface
	wfrEventUpdater EventUpdater
	// cache all events and update every updateInterval to prevent sending too much request to api server
	latestUpdateTime metav1.Time
	updateInterval   time.Duration
	events           map[string]v1alpha1.StageEvent
}

// NewPodEventWatcher creates a new pod event watcher.
func newPodEventWatcher(clusterClient kubernetes.Interface, client clientset.Interface, wfrNamespace, wfrName string) PodEventWatcher {
	return &podEventWatcher{
		clusterClient:   clusterClient,
		client:          client,
		wfrEventUpdater: NewEventUpdater(client, wfrNamespace, wfrName),
		updateInterval:  time.Duration(5 * time.Second),
		events:          make(map[string]v1alpha1.StageEvent),
	}
}

func (p *podEventWatcher) Work(stage, namespace, podName string) {
	c := make(chan struct{})
	go p.watchPodEvent(stage, namespace, podName, c)
	p.watchPod(namespace, podName, c)
	// Wait 1 second for workflowRun updating
	time.Sleep(1 * time.Second)
}

func (p *podEventWatcher) watchPodEvent(stage, namespace, podName string, c <-chan struct{}) {
	w, err := p.clusterClient.CoreV1().Events(namespace).Watch(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.kind=Pod,involvedObject.name=%s", podName),
	})
	if err != nil {
		log.WithField("namespace", namespace).
			WithField("pod name", podName).Info("Start to watch pod event failed")
		return
	}

	defer w.Stop()

	for {
		select {
		case e := <-w.ResultChan():
			event, ok := e.Object.(*corev1.Event)
			if !ok {
				log.WithField("event type", e.Type).WithField("event object", e.Object).Warn("Not an event")
			}

			// Skip Normal event
			if event.Type != corev1.EventTypeWarning {
				continue
			}

			p.events[event.Name] = v1alpha1.StageEvent{
				Name:          event.Name,
				Reason:        event.Reason,
				Message:       event.Message,
				LastTimestamp: event.LastTimestamp,
				Count:         event.Count,
			}

			p.update(stage, false)
		case <-c:
			p.update(stage, true)
			return
		}
	}

}

func (p *podEventWatcher) watchPod(namespace, podName string, c chan<- struct{}) {
	defer func() {
		c <- struct{}{}
	}()

	w, err := p.clusterClient.CoreV1().Pods(namespace).Watch(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", podName),
	})
	if err != nil {
		log.WithField("namespace", namespace).
			WithField("pod name", podName).Info("Start to watch pod failed")
		return
	}

	defer w.Stop()

	for {
		for e := range w.ResultChan() {
			// Pod deleted, stop watching.
			if e.Type == k8swatch.Deleted {
				return
			}

			pod, ok := e.Object.(*corev1.Pod)
			if !ok {
				log.WithField("event type", e.Type).WithField("event object", e.Object).Warn("Not a pod")
			}
			if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
				return
			}
		}
	}
}

// update updates events to workflowRun every interval time(e.g. 5 seconds)
func (p *podEventWatcher) update(stage string, updateRightNow bool) {
	if len(p.events) == 0 {
		return
	}

	now := metav1.Now()
	if !updateRightNow && now.Sub(p.latestUpdateTime.Time) < p.updateInterval {
		return
	}

	err := p.wfrEventUpdater.UpdateEvents(stage, eventsMapToSlice(p.events))
	if err != nil {
		log.WithField("stage", stage).WithField("err", err).Warn("Update events failed")
		return
	}
	p.events = make(map[string]v1alpha1.StageEvent)
	p.latestUpdateTime = now
}

// EventUpdater update events to the stage status of corresponding workflowRun's status.
// Event with same name will be override.
type EventUpdater interface {
	UpdateEvents(stage string, events []v1alpha1.StageEvent) error
}

type eventUpdater struct {
	client       clientset.Interface
	wfrNamespace string
	wfrName      string
	processor    *eventUpdateProcessor
}

// NewEventUpdater creates an event updater
func NewEventUpdater(client clientset.Interface, wfrNamespace, wfrName string) EventUpdater {
	return &eventUpdater{
		client:       client,
		wfrNamespace: wfrNamespace,
		wfrName:      wfrName,
		processor:    newEventUpdateProcessor(),
	}
}

func (e *eventUpdater) UpdateEvents(stage string, events []v1alpha1.StageEvent) error {
	// Update WorkflowRun status event with retry.
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// get latest wfr
		latest, err := e.client.CycloneV1alpha1().WorkflowRuns(e.wfrNamespace).Get(e.wfrName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		wfr := latest.DeepCopy()
		var stageStatus *v1alpha1.StageStatus
		for k, v := range latest.Status.Stages {
			if k == stage {
				stageStatus = v
			}
		}
		if stageStatus == nil {
			return fmt.Errorf("stage %s status not found", stage)
		}

		// deduplicate and sort
		combined := e.processor.process(stageStatus.Events, events)

		// update wfr
		wfr.Status.Stages[stage].Events = combined
		_, err = e.client.CycloneV1alpha1().WorkflowRuns(e.wfrNamespace).Update(wfr)
		return err
	})
}

type eventUpdateProcessor struct {
	deduplicator *eventUpdateDeduplicator
	sorter       *eventSorter
}

func newEventUpdateProcessor() *eventUpdateProcessor {
	return &eventUpdateProcessor{
		deduplicator: newEventUpdateDeduplicator(),
		sorter:       newEventSorter(),
	}
}

func (p *eventUpdateProcessor) process(older, newer []v1alpha1.StageEvent) []v1alpha1.StageEvent {
	events := p.deduplicator.deduplicat(older, newer)
	p.sorter.sort(events)
	return events
}

type eventUpdateDeduplicator struct {
	events map[string]v1alpha1.StageEvent
}

func newEventUpdateDeduplicator() *eventUpdateDeduplicator {
	return &eventUpdateDeduplicator{events: make(map[string]v1alpha1.StageEvent)}
}

func (p *eventUpdateDeduplicator) deduplicat(older, newer []v1alpha1.StageEvent) []v1alpha1.StageEvent {
	for _, o := range older {
		p.events[o.Name] = o
	}

	for _, n := range newer {
		p.events[n.Name] = n
	}

	return eventsMapToSlice(p.events)
}

// eventSort implements sort.Interface based on the Age field.
type eventSort []v1alpha1.StageEvent

func (e eventSort) Len() int           { return len(e) }
func (e eventSort) Less(i, j int) bool { return e[i].LastTimestamp.After(e[j].LastTimestamp.Time) }
func (e eventSort) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }

type eventSorter struct {
}

func newEventSorter() *eventSorter {
	return &eventSorter{}
}
func (w *eventSorter) sort(events []v1alpha1.StageEvent) {
	sort.Sort(eventSort(events))
}

func eventsMapToSlice(in map[string]v1alpha1.StageEvent) []v1alpha1.StageEvent {
	if in == nil {
		return nil
	}

	events := make([]v1alpha1.StageEvent, 0, len(in))
	for _, v := range in {
		events = append(events, v)
	}
	return events
}
