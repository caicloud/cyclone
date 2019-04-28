package common

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
)

// NewClusterClient creates a client for k8s cluster
func NewClusterClient(c *v1alpha1.ClusterCredential, inCluster bool) (*kubernetes.Clientset, error) {
	if inCluster {
		client, err := newInclusterK8sClient()
		if err == nil {
			return client, nil
		}
	}
	return newK8sClient(c)
}

func newK8sClient(c *v1alpha1.ClusterCredential) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// if KubeConfig is not empty, use it firstly, otherwise, use username/password.
	if c.KubeConfig != nil {
		config, err = clientcmd.NewDefaultClientConfig(*c.KubeConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			log.Infof("NewDefaultClientConfig error: %v", err)
			return nil, err
		}
	} else {
		if c.TLSClientConfig == nil {
			c.TLSClientConfig = &v1alpha1.TLSClientConfig{Insecure: true}
		}

		config = &rest.Config{
			Host:        c.Server,
			BearerToken: c.BearerToken,
			Username:    c.User,
			Password:    c.Password,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: c.TLSClientConfig.Insecure,
				CAFile:   c.TLSClientConfig.CAFile,
				CAData:   c.TLSClientConfig.CAData,
			},
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return client, nil
}

func newInclusterK8sClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return client, nil
}
