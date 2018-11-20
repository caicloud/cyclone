package admin

import (
	"fmt"
	"net/http"

	restful "github.com/emicklei/go-restful"
	"github.com/golang/glog"

	"github.com/caicloud/storage-admin/pkg/admin/datavolume"
	"github.com/caicloud/storage-admin/pkg/admin/storageclass"
	"github.com/caicloud/storage-admin/pkg/admin/storageservice"
	"github.com/caicloud/storage-admin/pkg/admin/storagetype"
	"github.com/caicloud/storage-admin/pkg/constants"
)

const (
	APIVersion = constants.APIVersion
)

func (s *Server) initialize() {
	restful.EnableTracing(true)
	s.addServiceEndpoints()
}

func (s *Server) addServiceEndpoints() {
	ws := new(restful.WebService)

	ws.Path(constants.RootPath).
		Consumes(constants.MimeJson, constants.MimeText).
		Produces(constants.MimeJson, constants.MimeText)

	storagetype.AddEndpoints(ws, s.c)
	storageservice.AddEndpoints(ws, s.c)
	storageclass.AddEndpoints(ws, s.c)
	datavolume.AddEndpoints(ws, s.c)

	restful.Add(ws)
}

func (s *Server) Run(stopCh chan struct{}, port int) {
	s.initialize()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: restful.DefaultContainer,
	}

	go func() {
		<-stopCh
		glog.Infof("receive close signal")
		server.Close()
	}()

	e := server.ListenAndServe()
	if e != nil && e != http.ErrServerClosed {
		glog.Fatalf("ListenAndServe failed, %v", e)
	}
	glog.Info("ListenAndServe closed")
}
