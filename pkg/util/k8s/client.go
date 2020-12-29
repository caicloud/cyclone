package k8s

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	cyscheme "github.com/caicloud/cyclone/pkg/k8s/clientset/scheme"
	cyclonev1alpha1 "github.com/caicloud/cyclone/pkg/k8s/clientset/typed/cyclone/v1alpha1"
)

// Scheme consists of kubernetes and cyclone scheme.
var Scheme = runtime.NewScheme()

func init() {
	_ = cyscheme.AddToScheme(Scheme)
	_ = scheme.AddToScheme(Scheme)
}

// Interface consists of kubernetes and cyclone interfaces
type Interface interface {
	kubernetes.Interface
	CycloneV1alpha1() cyclonev1alpha1.CycloneV1alpha1Interface
}

// Clientset contains the cyclone Clientset and kubernetes Clientset
type Clientset struct {
	*kubernetes.Clientset
	cycloneClient *clientset.Clientset
}

// CycloneV1alpha1 retrieves the CycloneV1alpha1Client
func (c *Clientset) CycloneV1alpha1() cyclonev1alpha1.CycloneV1alpha1Interface {
	return c.cycloneClient.CycloneV1alpha1()
}

var _ Interface = (*Clientset)(nil)

// NewForConfig creates a new Clientset for the given config.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	cycloneClient, err := clientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	return &Clientset{
		Clientset:     kubeClient,
		cycloneClient: cycloneClient,
	}, nil
}

// GetClient creates a client for k8s cluster
func GetClient(kubeConfigPath string) (Interface, error) {
	config, err := getKubeConfig(kubeConfigPath)
	if err != nil {
		return nil, err
	}

	return NewForConfig(config)
}

// GetRateLimitClient creates a client for k8s cluster with custom defined qps and burst.
func GetRateLimitClient(kubeConfigPath string, qps float32, burst int) (Interface, error) {
	config, err := getKubeConfig(kubeConfigPath)
	if err != nil {
		return nil, err
	}

	if qps > 0.0 {
		config.QPS = qps
	}

	if burst > 0 {
		config.Burst = burst
	}

	return NewForConfig(config)
}

func getKubeConfig(kubeConfigPath string) (*rest.Config, error) {
	var config *rest.Config
	var err error
	if kubeConfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return nil, err
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}
