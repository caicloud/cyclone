package executioncluster

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
	"github.com/caicloud/cyclone/pkg/workflow/controller/handlers"
	"github.com/caicloud/cyclone/pkg/workflow/controller/store"
)

// Handler ...
type Handler struct {
	Client clientset.Interface
}

// Ensure *Handler has implemented handlers.Interface interface.
var _ handlers.Interface = (*Handler)(nil)

const (
	// finalizerExecutioncluster is the cyclone related finalizer key for execution cluster.
	finalizerExecutioncluster string = "executioncluster.cyclone.dev/finalizer"
)

// NewHandler ...
func NewHandler(client clientset.Interface) *Handler {
	return &Handler{
		Client: client,
	}
}

// Reconcile compares the actual state with the desired, and attempts to
// converge the two.
func (h *Handler) Reconcile(obj interface{}) (res controller.Result, err error) {
	cluster, ok := obj.(*v1alpha1.ExecutionCluster)
	if !ok {
		log.WithField("obj", obj).Warning("Expect ExecutionCluster, got unknown type resource")
		return res, fmt.Errorf("unknown resource type")
	}
	log.WithField("name", cluster.Name).Debug("Observed execution cluster")

	if err := store.RegisterClusterController(cluster); err != nil {
		log.WithField("name", cluster.Name).Error("Register execution cluster controller error: ", err)
		return res, err
	}
	return res, nil
}

// finalize ...
func (h *Handler) finalize(ec *v1alpha1.ExecutionCluster) error {
	if err := store.RemoveClusterController(ec); err != nil {
		log.WithField("name", ec.Name).Error("Remove execution cluster controller error: ", err)
		return err
	}

	return nil
}

// AddFinalizer adds a finalizer to the object and update the object to the Kubernetes.
func (h *Handler) AddFinalizer(obj interface{}) error {
	originCluster, ok := obj.(*v1alpha1.ExecutionCluster)
	if !ok {
		log.WithField("obj", obj).Warning("Expect ExecutionCluster, got unknown type resource")
		return fmt.Errorf("unknown resource type")
	}

	if sets.NewString(originCluster.Finalizers...).Has(finalizerExecutioncluster) {
		return nil
	}

	log.WithField("name", originCluster.Name).Debug("Start to add finalizer for executionCluster")

	ec := originCluster.DeepCopy()
	ec.ObjectMeta.Finalizers = append(ec.ObjectMeta.Finalizers, finalizerExecutioncluster)
	_, err := h.Client.CycloneV1alpha1().ExecutionClusters().Update(ec)
	return err
}

// HandleFinalizer does the finalizer key representing things.
func (h *Handler) HandleFinalizer(obj interface{}) error {
	originCluster, ok := obj.(*v1alpha1.ExecutionCluster)
	if !ok {
		log.WithField("obj", obj).Warning("Expect ExecutionCluster, got unknown type resource")
		return fmt.Errorf("unknown resource type")
	}

	if !sets.NewString(originCluster.Finalizers...).Has(finalizerExecutioncluster) {
		return nil
	}

	log.WithField("name", originCluster.Name).Debug("Start to process finalizer for executionCluster")

	// Handler finalizer
	ec := originCluster.DeepCopy()
	if err := h.finalize(ec); err != nil {
		return nil
	}

	ec.ObjectMeta.Finalizers = sets.NewString(ec.ObjectMeta.Finalizers...).Delete(finalizerExecutioncluster).UnsortedList()
	_, err := h.Client.CycloneV1alpha1().ExecutionClusters().Update(ec)
	return err
}
