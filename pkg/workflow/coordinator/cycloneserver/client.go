package cycloneserver

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"

	websocketutil "github.com/caicloud/cyclone/pkg/util/websocket"
)

const (
	cycloneAPIVersion = "/apis/v1alpha1"

	apiPathForLogStream = "/workflowruns/%s/streamlogs"
)

// Client ...
type Client interface {
	PushLogStream(ns, workflowrun, stage, container string, reader io.Reader, close chan struct{}) error
}

type client struct {
	baseURL string
	client  *http.Client
}

// NewClient ...
func NewClient(cycloneServer string) Client {
	baseURL := strings.TrimRight(cycloneServer, "/")
	if !strings.Contains(baseURL, "://") {
		baseURL = "http://" + baseURL
	}

	return &client{
		baseURL: baseURL,
		client:  http.DefaultClient,
	}
}

// PushLogStream ...
func (c *client) PushLogStream(ns, workflowrun, stage, container string, reader io.Reader, close chan struct{}) error {
	path := fmt.Sprintf(apiPathForLogStream, workflowrun)
	host := strings.TrimPrefix(c.baseURL, "http://")
	host = strings.TrimPrefix(host, "https://")
	requestURL := url.URL{
		Host:     host,
		Path:     cycloneAPIVersion + path,
		RawQuery: fmt.Sprintf("namespace=%s&stage=%s&container=%s", ns, stage, container),
		Scheme:   "ws",
	}

	log.Infof("Path: %s", requestURL.String())
	return websocketutil.SendStream(requestURL.String(), reader, close)

}
