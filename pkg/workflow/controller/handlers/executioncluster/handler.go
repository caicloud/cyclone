package executioncluster

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"
	"github.com/caicloud/cyclone/pkg/workflow/controller/store"
)

// Handler ...
type Handler struct {
	Client clientset.Interface
}

// Ensure *Handler has implemented handlers.Interface interface.
var _ handlers.Interface = (*Handler)(nil)

// Reconcile compares the actual state with the desired, and attempts to
// converge the two.
func (h *Handler) Reconcile(obj interface{}) error {
	cluster, ok := obj.(*v1alpha1.ExecutionCluster)
	if !ok {
		log.WithField("obj", obj).Warning("Expect ExecutionCluster, got unknown type resource")
		return fmt.Errorf("unknown resource type")
	}
	log.WithField("name", cluster.Name).Debug("Observed execution cluster")

	err := store.RegisterClusterController(cluster)
	if err != nil {
		log.WithField("name", cluster.Name).Error("Register execution cluster controller error: ", err)
	}
	return err
}

// ObjectDeleted ...
func (h *Handler) ObjectDeleted(obj interface{}) error {
	cluster, ok := obj.(*v1alpha1.ExecutionCluster)
	if !ok {
		log.WithField("obj", obj).Warning("Expect ExecutionCluster, got unknown type resource")
		return fmt.Errorf("unknown resource type")
	}
	log.WithField("name", cluster.Name).Debug("Observed execution cluster deletion")

	err := store.RemoveClusterController(cluster)
	if err != nil {
		log.WithField("name", cluster.Name).Error("Remove execution cluster controller error: ", err)
	}

	return err
}
