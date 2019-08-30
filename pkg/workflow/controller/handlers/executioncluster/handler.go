package executioncluster

import (
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

// ObjectCreated ...
func (h *Handler) ObjectCreated(obj interface{}) {
	cluster, ok := obj.(*v1alpha1.ExecutionCluster)
	if !ok {
		log.Warning("unknown resource type")
		return
	}
	log.WithField("name", cluster.Name).Debug("Observed new execution cluster")

	err := store.RegisterClusterController(cluster)
	if err != nil {
		log.WithField("name", cluster.Name).Error("Register execution cluster controller error: ", err)
	}
}

// ObjectUpdated ...
func (h *Handler) ObjectUpdated(old, new interface{}) {
	cluster, ok := new.(*v1alpha1.ExecutionCluster)
	if !ok {
		log.Warning("unknown resource type")
		return
	}
	log.WithField("name", cluster.Name).Debug("Observed new update to execution cluster")

	err := store.RegisterClusterController(cluster)
	if err != nil {
		log.WithField("name", cluster.Name).Error("Register execution cluster controller error: ", err)
	}
}

// ObjectDeleted ...
func (h *Handler) ObjectDeleted(obj interface{}) {
	cluster, ok := obj.(*v1alpha1.ExecutionCluster)
	if !ok {
		log.Warning("unknown resource type")
		return
	}
	log.WithField("name", cluster.Name).Debug("Observed execution cluster deletion")

	err := store.RemoveClusterController(cluster)
	if err != nil {
		log.WithField("name", cluster.Name).Error("Remove execution cluster controller error: ", err)
	}
}
