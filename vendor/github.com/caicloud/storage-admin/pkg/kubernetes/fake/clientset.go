package fake

import (
	kubefake "github.com/caicloud/clientset/kubernetes/fake"
	"k8s.io/client-go/rest"
)

type Clientset struct {
	*kubefake.Clientset

	restConfig *rest.Config
}

func NewSimpleClientset() *Clientset {
	return &Clientset{
		Clientset:  kubefake.NewSimpleClientset(),
		restConfig: &rest.Config{},
	}
}

func (c *Clientset) RestConfig() *rest.Config { return c.restConfig }

func (c *Clientset) SetRestConfig(cfg *rest.Config) *Clientset {
	c.restConfig = cfg
	return c
}

func NewClientset(kube *kubefake.Clientset) *Clientset {
	return &Clientset{
		Clientset: kube,
	}
}

var (
	NewSimpleFakeKubeClientset = kubefake.NewSimpleClientset
)
