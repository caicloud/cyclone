package kubernetes

import (
	"net/http"

	"github.com/caicloud/clientset/kubernetes"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	IsNotFound      = errors.IsNotFound
	IsAlreadyExists = errors.IsAlreadyExists
	IsConflict      = errors.IsConflict
	IsInvalid       = errors.IsInvalid
)

type Config = rest.Config

type Interface interface {
	kubernetes.Interface

	RestConfig() *Config
}

// 直接调用 k8s 方式实现
type Clientset struct {
	kubernetes.Interface
	restConfig Config
}

func NewClientFromFlags(masterUrl, kubeconfigPath string) (Interface, error) {
	cfg, e := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
	if e != nil {
		return nil, e
	}

	return NewClientFromRestConfig(cfg)
}

func NewClientFromRestConfig(restConf *rest.Config) (Interface, error) {
	cs, e := kubernetes.NewForConfig(restConf)
	if e != nil {
		return nil, e
	}

	return &Clientset{
		restConfig: *restConf,
		Interface:  cs,
	}, nil
}

func NewClientFromUser(host, user, pwd string) (Interface, error) {
	return NewClientFromRestConfig(&rest.Config{
		Host:     host,
		Username: user,
		Password: pwd,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	})
}

func NewClientFromHttpHeader(req *http.Request) (Interface, error) {
	cfg, e := ParseConfigFromRequest(req)
	if e != nil {
		return nil, e
	}
	return NewClientFromRestConfig(cfg)
}

func (c *Clientset) RestConfig() *rest.Config            { return &c.restConfig }
func (c *Clientset) KubeClientSet() kubernetes.Interface { return c.Interface }

func (c *Clientset) Close() error {
	return nil
}
