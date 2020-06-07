package workflowtrigger

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	finalizer "github.com/caicloud/cyclone/pkg/workflow/controller/finalizers"
	"github.com/caicloud/cyclone/pkg/workflow/workload/pod"
)

const (
	// finalizerDeleteCronTask is the finalizer key representing deleting cron task
	// as the corresponding workflow trigger is being deleted.
	finalizerDeleteCronTask string = "workflowtrigger.cyclone.dev/finalizer-delete-cron-task"
)

// NewFinalizer ...
func NewFinalizer(client clientset.Interface) finalizer.Interface {
	return finalizer.NewFinalizer(client, nil, updateFinalizer, appendFinalizer, removeFinalizer, map[string]finalizer.Handler{
		finalizerDeleteCronTask: handleFinalizerDeleteCronTask,
	})
}

// updateFinalizer updates the obj to the Kubernetes cluster
func updateFinalizer(client clientset.Interface, _ kubernetes.Interface, obj finalizer.RuntimeAndMetaInterface) error {
	wft, ok := obj.(*v1alpha1.WorkflowTrigger)
	if !ok {
		return fmt.Errorf("resource type not support")
	}
	_, err := client.CycloneV1alpha1().WorkflowTriggers(wft.Namespace).Update(wft)
	return err
}

// appendFinalizer appends a finalizer to the obj
func appendFinalizer(obj finalizer.RuntimeAndMetaInterface, f string) (finalizer.RuntimeAndMetaInterface, error) {
	wft, ok := obj.(*v1alpha1.WorkflowTrigger)
	if !ok {
		return obj, fmt.Errorf("resource type not support")
	}

	wft.ObjectMeta.Finalizers = append(wft.ObjectMeta.Finalizers, f)
	return wft, nil
}

// removeFinalizer removes a finalizer in the obj
func removeFinalizer(obj finalizer.RuntimeAndMetaInterface, f string) (finalizer.RuntimeAndMetaInterface, error) {
	wft, ok := obj.(*v1alpha1.WorkflowTrigger)
	if !ok {
		return obj, fmt.Errorf("resource type not support")
	}

	wft.ObjectMeta.Finalizers = finalizer.RemoveString(wft.ObjectMeta.Finalizers, f)
	return wft, nil
}

// handleFinalizerDeleteCronTask handles the delete cron task finalizer
func handleFinalizerDeleteCronTask(client clientset.Interface, clusterClient kubernetes.Interface, obj finalizer.RuntimeAndMetaInterface) error {
	wft, err := ToWorkflowTrigger(obj)
	if err != nil {
		log.Warn("Convert to WorkflowTrigger error: ", err)
		return err
	}

	log.WithField("name", pod.Name).Debug("Observed workflowTrigger is being deleted")
	if wft.Spec.Type == v1alpha1.TriggerTypeCron {
		// TODO,Fixme: the handler can not invoke c.CronManager
		// c.CronManager.DeleteCron(wft)
	}
	return nil
}
