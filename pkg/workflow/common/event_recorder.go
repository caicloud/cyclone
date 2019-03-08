package common

import (
	"sync"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
)

const (
	// EventSourceWfrController represents events send from workflowrun controller.
	EventSourceWfrController string = "WorkflowRunController"
)

// EventRecorder is used to record events to k8s, controllers here would use it to record
// events reflecting the WorkflowRun executing process.
var eventRecorder record.EventRecorder

// Once is used to ensure that the eventRecorder is initailized only once.
var once sync.Once

// GetEventRecorder get the event recorder object. Create it of not exists yet.
func GetEventRecorder(client clientset.Interface, component string) record.EventRecorder {
	once.Do(func() {
		log.Info("Creating event recorder")
		broadcaster := record.NewBroadcaster()
		broadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: client.CoreV1().Events("")})
		eventRecorder = broadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: component})
	})

	return eventRecorder
}
