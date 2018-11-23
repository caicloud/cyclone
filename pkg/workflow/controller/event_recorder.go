package controller

import (
	"sync"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
)

var once sync.Once
var eventRecorder record.EventRecorder

func GetEventRecorder(client clientset.Interface) record.EventRecorder {
	once.Do(func() {
		log.Info("Creating event recorder")
		broadcaster := record.NewBroadcaster()
		broadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: client.CoreV1().Events("")})
		eventRecorder = broadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "WorkflowRun Controller"})
	})

	return eventRecorder
}