package common

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/caicloud/cyclone/pkg/k8s/clientset"
)

// GetClient creates a client for k8s cluster
func GetClient(masterUrl, kubeConfigPath string) (clientset.Interface, error) {
	var config *rest.Config
	var err error
	if kubeConfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags(masterUrl, kubeConfigPath)
		if err != nil {
			return nil, err
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	return clientset.NewForConfig(config)
}
