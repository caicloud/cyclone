package store

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/common"
)

// ControllerRegistry manages all controllers created for each cluster.
// When a new Cluster found, create controller for it and add its stop
// channel to this registry if not exist. When a Cluster removed, close
// the channel in this map, and delete the key.
var ControllerRegistry map[string]*ClusterController
var lock *sync.Mutex

// ClusterController ...
type ClusterController struct {
	// Cluster ...
	Cluster *v1alpha1.ExecutionCluster
	// Cluster client
	Client kubernetes.Interface
	// Channel that can be used to stop the controller
	StopCh chan struct{}
}

// NewClusterChan is channels for clusters that got created.
var NewClusterChan chan *ClusterController

func init() {
	ControllerRegistry = make(map[string]*ClusterController)
	lock = &sync.Mutex{}
	NewClusterChan = make(chan *ClusterController)
}

// RegisterClusterController register a cluster to this registry.
func RegisterClusterController(cluster *v1alpha1.ExecutionCluster) error {
	lock.Lock()
	defer lock.Unlock()

	if _, ok := ControllerRegistry[cluster.Name]; ok {
		log.Infof("cluster %s already registered, skip it", cluster.Name)
		return nil
	}

	log.Infof("register cluster controller for cluster %s", cluster.Name)

	client, err := getClusterClient(cluster)
	if err != nil {
		log.Errorf("create client for cluster %s error: %v", cluster.Name, err)
		return err
	}

	clusterController := &ClusterController{
		Cluster: cluster,
		Client:  client,
		StopCh:  make(chan struct{}),
	}
	NewClusterChan <- clusterController
	ControllerRegistry[cluster.Name] = clusterController

	return nil
}

// RemoveClusterController stop and remove cluster from this registry.
func RemoveClusterController(cluster *v1alpha1.ExecutionCluster) error {
	lock.Lock()
	defer lock.Unlock()

	c, ok := ControllerRegistry[cluster.Name]
	if !ok {
		return nil
	}

	close(c.StopCh)
	delete(ControllerRegistry, cluster.Name)
	return nil
}

// GetClusterClient gets cluster client with the given cluster name
func GetClusterClient(name string) kubernetes.Interface {
	lock.Lock()
	defer lock.Unlock()

	c, ok := ControllerRegistry[name]
	if ok {
		return c.Client
	}

	return nil
}

// getClusterClient get kube client from cluster crd
func getClusterClient(cluster *v1alpha1.ExecutionCluster) (kubernetes.Interface, error) {
	if cluster == nil {
		return nil, fmt.Errorf("nil cluster")
	}

	return common.NewClusterClient(&cluster.Spec.Credential, cluster.Name == common.ControlClusterName)
}
