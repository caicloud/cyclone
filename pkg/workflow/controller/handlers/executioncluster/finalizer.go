package executioncluster

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	finalizer "github.com/caicloud/cyclone/pkg/workflow/controller/finalizers"
	"github.com/caicloud/cyclone/pkg/workflow/controller/store"
)

const (
	// finalizerDeleteClusterController is the finalizer key representing deleting cluster
	// cluster controller as the corresponding workflow trigger is being deleted.
	finalizerDeleteClusterController string = "executioncluster.cyclone.dev/finalizer-delete-cluster-controller"
)

// NewFinalizer ...
func NewFinalizer(client clientset.Interface) finalizer.Interface {
	return finalizer.NewFinalizer(client, nil, updateFinalizer, appendFinalizer, removeFinalizer, map[string]finalizer.Handler{
		finalizerDeleteClusterController: handleFinalizerDeleteClusterController,
	})
}

// updateFinalizer updates the obj to the Kubernetes cluster
func updateFinalizer(client clientset.Interface, _ kubernetes.Interface, obj finalizer.RuntimeAndMetaInterface) error {
	cluster, ok := obj.(*v1alpha1.ExecutionCluster)
	if !ok {
		return fmt.Errorf("resource type not support")
	}
	_, err := client.CycloneV1alpha1().ExecutionClusters().Update(cluster)
	return err
}

// appendFinalizer appends a finalizer to the obj
func appendFinalizer(obj finalizer.RuntimeAndMetaInterface, f string) (finalizer.RuntimeAndMetaInterface, error) {
	cluster, ok := obj.(*v1alpha1.ExecutionCluster)
	if !ok {
		return obj, fmt.Errorf("resource type not support")
	}

	cluster.ObjectMeta.Finalizers = append(cluster.ObjectMeta.Finalizers, f)
	return cluster, nil
}

// removeFinalizer removes a finalizer in the obj
func removeFinalizer(obj finalizer.RuntimeAndMetaInterface, f string) (finalizer.RuntimeAndMetaInterface, error) {
	cluster, ok := obj.(*v1alpha1.ExecutionCluster)
	if !ok {
		return obj, fmt.Errorf("resource type not support")
	}

	cluster.ObjectMeta.Finalizers = finalizer.RemoveString(cluster.ObjectMeta.Finalizers, f)
	return cluster, nil
}

// handleFinalizerDeleteClusterController handles the delete cluster controller finalizer
func handleFinalizerDeleteClusterController(client clientset.Interface, clusterClient kubernetes.Interface, obj finalizer.RuntimeAndMetaInterface) error {
	cluster, ok := obj.(*v1alpha1.ExecutionCluster)
	if !ok {
		log.WithField("obj", obj).Warning("Expect ExecutionCluster, got unknown type resource")
		return fmt.Errorf("unknown resource type")
	}
	log.WithField("name", cluster.Name).Debug("Observed execution cluster deletion")

	if err := store.RemoveClusterController(cluster); err != nil {
		log.WithField("name", cluster.Name).Error("Remove execution cluster controller error: ", err)
		return err
	}

	return nil
}
