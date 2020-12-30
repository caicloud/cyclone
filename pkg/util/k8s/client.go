package k8s

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	cyscheme "github.com/caicloud/cyclone/pkg/k8s/clientset/scheme"
)

// Scheme consists of kubernetes and cyclone scheme.
var Scheme = runtime.NewScheme()

func init() {
	cyscheme.AddToScheme(Scheme)
	_ = scheme.AddToScheme(Scheme)
}

// GetClient creates a client for k8s cluster
func GetClient(kubeConfigPath string) (clientset.Interface, error) {
	config, err := getKubeConfig(kubeConfigPath)
	if err != nil {
		return nil, err
	}

	return clientset.NewForConfig(config)
}

// GetRateLimitClient creates a client for k8s cluster with custom defined qps and burst.
func GetRateLimitClient(kubeConfigPath string, qps float32, burst int) (clientset.Interface, error) {
	config, err := getKubeConfig(kubeConfigPath)
	if err != nil {
		return nil, err
	}

	if qps > 0.0 {
		config.QPS = float32(qps)
	}

	if burst > 0 {
		config.Burst = burst
	}

	return clientset.NewForConfig(config)
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
