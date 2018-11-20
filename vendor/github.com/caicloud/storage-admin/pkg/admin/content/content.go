package content

import (
	"github.com/caicloud/storage-admin/pkg/cluster"
	"github.com/caicloud/storage-admin/pkg/errors"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
)

type Interface interface {
	GetClient() kubernetes.Interface
	GetSubClient(clusterName string) (kubernetes.Interface, *errors.FormatError)
}

type SimpleContent struct {
	kc kubernetes.Interface

	ctrlCluster string
	cdsHost     string
}

func NewSimpleContent(kc kubernetes.Interface, ctrlCluster, cdsHost string) (Interface, error) {
	if kc == nil {
		return nil, errors.ErrVarKubeClientNil
	}
	if len(cdsHost) == 0 {
		return nil, errors.ErrVarCdsHostEmpty
	}
	s := &SimpleContent{
		kc:          kc,
		ctrlCluster: ctrlCluster,
		cdsHost:     cdsHost,
	}
	return s, nil
}

func (c *SimpleContent) GetClient() kubernetes.Interface { return c.kc }

func (c *SimpleContent) GetSubClient(clusterName string) (kubernetes.Interface, *errors.FormatError) {
	if clusterName == c.ctrlCluster {
		return c.kc, nil
	}
	// TODO cache
	kc, e := cluster.GetKubeClientFromClusterAdmin(c.cdsHost, clusterName)
	if e != nil {
		return nil, errors.NewError().SetErrorBadClusterConfig(clusterName, e)
	}
	return kc, nil
}
