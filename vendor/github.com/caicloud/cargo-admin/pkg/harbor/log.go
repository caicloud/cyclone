package harbor

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	. "github.com/caicloud/cargo-admin/pkg/errors"

	"github.com/caicloud/nirvana/log"
)

type LogClienter interface {
	ListLogs(startTime int64, endTime int64, op string) ([]*HarborAccessLog, error)
	ListProjectLogs(pid int64, startTime int64, endTime int64, op string) ([]*HarborAccessLog, error)
}

func (c *Client) ListLogs(startTime int64, endTime int64, op string) ([]*HarborAccessLog, error) {
	path := LogsPath(startTime, endTime, op)

	log.Infof("%s %s", http.MethodGet, path)
	resp, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}

	if resp.StatusCode/100 == 2 {
		hlogs := make([]*HarborAccessLog, 0)
		err := json.Unmarshal(body, &hlogs)
		if err != nil {
			log.Errorf("unmarshal logs error: %v", err)
			return nil, ErrorUnknownInternal.Error(err)
		}
		return logFilter(hlogs), nil
	}

	return nil, ErrorUnknownInternal.Error(body)
}

func (c *Client) ListProjectLogs(pid int64, startTime int64, endTime int64, op string) ([]*HarborAccessLog, error) {
	path := ProjectLogsPath(pid, startTime, endTime, op)

	log.Infof("%s %s", http.MethodGet, path)
	resp, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}

	if resp.StatusCode/100 == 2 {
		hlogs := make([]*HarborAccessLog, 0)
		err := json.Unmarshal(body, &hlogs)
		if err != nil {
			log.Errorf("unmarshal logs error: %v", err)
			return nil, ErrorUnknownInternal.Error(err)
		}
		return logFilter(hlogs), nil
	}

	return nil, ErrorUnknownInternal.Error(body)
}

func logFilter(hlogs []*HarborAccessLog) []*HarborAccessLog {
	ret := make([]*HarborAccessLog, 0)
	for _, hlog := range hlogs {
		if hlog.RepoName == "library/hello-world" && hlog.RepoTag == "latest" && hlog.Operation == "pull" {
			continue
		}
		ret = append(ret, hlog)
	}
	return ret
}
