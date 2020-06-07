package workflowrun

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/common"
	finalizer "github.com/caicloud/cyclone/pkg/workflow/controller/finalizers"
	"github.com/caicloud/cyclone/pkg/workflow/workflowrun"
)

const (
	// finalizerGC is the finalizer key representing workflowRun gc
	finalizerGC string = "workflowrun.cyclone.dev/finalizer-gc"
	// finalizerParallelism is the finalizer key representing removing workflowRun from parallelism queue
	finalizerParallelism string = "workflowrun.cyclone.dev/finalizer-parallelism"
)

// NewFinalizer ...
func NewFinalizer(client clientset.Interface) finalizer.Interface {
	return finalizer.NewFinalizer(client, nil, updateFinalizer, appendFinalizer, removeFinalizer, map[string]finalizer.Handler{
		finalizerGC:          handleFinalizerGC,
		finalizerParallelism: handleFinalizerParallelism,
	})
}

// updateFinalizer updates the obj to the Kubernetes cluster
func updateFinalizer(client clientset.Interface, _ kubernetes.Interface, obj finalizer.RuntimeAndMetaInterface) error {
	wfr, ok := obj.(*v1alpha1.WorkflowRun)
	if !ok {
		return fmt.Errorf("resource type not support")
	}
	_, err := client.CycloneV1alpha1().WorkflowRuns(wfr.Namespace).Update(wfr)
	return err
}

// appendFinalizer appends a finalizer to the obj
func appendFinalizer(obj finalizer.RuntimeAndMetaInterface, f string) (finalizer.RuntimeAndMetaInterface, error) {
	wfr, ok := obj.(*v1alpha1.WorkflowRun)
	if !ok {
		return obj, fmt.Errorf("resource type not support")
	}

	wfr.ObjectMeta.Finalizers = append(wfr.ObjectMeta.Finalizers, f)
	return wfr, nil
}

// removeFinalizer removes a finalizer in the obj
func removeFinalizer(obj finalizer.RuntimeAndMetaInterface, f string) (finalizer.RuntimeAndMetaInterface, error) {
	wfr, ok := obj.(*v1alpha1.WorkflowRun)
	if !ok {
		return obj, fmt.Errorf("resource type not support")
	}

	wfr.ObjectMeta.Finalizers = finalizer.RemoveString(wfr.ObjectMeta.Finalizers, f)
	return wfr, nil
}

// handleFinalizerGC handles the gc finalizer
func handleFinalizerGC(client clientset.Interface, _ kubernetes.Interface, obj finalizer.RuntimeAndMetaInterface) error {
	wfr, ok := obj.(*v1alpha1.WorkflowRun)
	if !ok {
		return fmt.Errorf("resource type not support")
	}

	log.WithField("name", wfr.Name).Debug("Start to GC for WorkflowRun delete")
	clusterClient := common.GetExecutionClusterClient(wfr)
	if clusterClient == nil {
		log.WithField("wfr", wfr.Name).Error("Execution cluster client not found")
		return fmt.Errorf("Execution cluster client not found")
	}

	operator, err := workflowrun.NewOperator(clusterClient, client, wfr, wfr.Namespace)
	if err != nil {
		log.WithField("wfr", wfr.Name).Error("Failed to create workflowrun operator: ", err)
		return err
	}

	if err = operator.GC(true, true); err != nil {
		log.WithField("wfr", wfr.Name).Warn("GC failed", err)
		return err
	}
	return nil
}

// handleFinalizerParallelism handles the parallelism finalizer
func handleFinalizerParallelism(client clientset.Interface, _ kubernetes.Interface, obj finalizer.RuntimeAndMetaInterface) error {
	wfr, ok := obj.(*v1alpha1.WorkflowRun)
	if !ok {
		return fmt.Errorf("resource type not support")
	}

	// h.ParallelismController.MarkFinished(originWfr.Namespace, originWfr.Name, originWfr.Spec.WorkflowRef.Name)
	log.WithField("name", wfr.Name).Debug("Start to perform ParallelismController")
	return nil
}
