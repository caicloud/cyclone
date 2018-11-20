package admin

import (
	"github.com/caicloud/storage-admin/pkg/admin/content"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
)

type Server struct {
	c content.Interface
}

func NewServer(kc kubernetes.Interface, ctrlCluster, clusterAdminAddr string) (*Server, error) {
	c, e := content.NewSimpleContent(kc, ctrlCluster, clusterAdminAddr)
	if e != nil {
		return nil, e
	}
	s := &Server{
		c: c,
	}
	return s, nil
}
