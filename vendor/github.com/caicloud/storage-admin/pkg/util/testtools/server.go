package testtools

import (
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	restful "github.com/emicklei/go-restful"

	"github.com/caicloud/storage-admin/pkg/admin/content"
	"github.com/caicloud/storage-admin/pkg/constants"
	"github.com/caicloud/storage-admin/pkg/errors"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
)

// port

const (
	PostBase = 2333
)

var (
	port = PostBase + time.Now().UnixNano()%666
)

func GetNewPort() int {
	np := atomic.AddInt64(&port, 1)
	return int(np)
}

// content

type FakeContent struct {
	cm map[string]kubernetes.Interface

	ctrlCluster string
}

func NewFakeContent(cm map[string]kubernetes.Interface, ctrlCluster string) *FakeContent {
	s := &FakeContent{
		cm:          cm,
		ctrlCluster: ctrlCluster,
	}
	return s
}

func (c *FakeContent) GetClient() kubernetes.Interface { return c.cm[c.ctrlCluster] }

func (c *FakeContent) GetSubClient(clusterName string) (kubernetes.Interface, *errors.FormatError) {
	kc := c.cm[clusterName]
	if kc == nil {
		return nil, errors.NewError().SetErrorBadClusterConfig(clusterName, errors.ErrVarKubeClientNil)
	}
	return kc, nil
}

// for test cases

func AddEndpointsFunc(c content.Interface,
	addEndpoints func(*restful.WebService, content.Interface)) func(*restful.WebService) {
	return func(ws *restful.WebService) {
		addEndpoints(ws, c)
	}
}

func RunHttpServer(port int, addEndpoints func(ws *restful.WebService), t *testing.T) (server *http.Server) {
	restful.EnableTracing(true)
	ws := new(restful.WebService)
	ws.Path(constants.RootPath).
		Consumes(constants.MimeJson, constants.MimeText).
		Produces(constants.MimeJson, constants.MimeText)

	addEndpoints(ws)

	container := restful.NewContainer()
	container.ServeMux = http.NewServeMux()
	container.Add(ws)

	server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: container,
	}

	go func() {
		e := server.ListenAndServe()
		if e != nil && e != http.ErrServerClosed {
			t.Fatalf("ListenAndServe failed, %v", e)
		} else if e == http.ErrServerClosed {
			t.Log("ListenAndServe closed")
		}
	}()

	// in case sometimes server may start a little later
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	for i := 0; i < 50; i++ {
		if c, e := net.Dial("tcp", addr); e == nil {
			c.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}
