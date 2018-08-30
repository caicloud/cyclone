package harbor

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	. "github.com/caicloud/cargo-admin/pkg/errors"

	"github.com/caicloud/nirvana/log"
)

const (
	// PriPC : count of private projects
	PriPC = "private_project_count"
	// PriRC : count of private repositories
	PriRC = "private_repo_count"
	// PubPC : count of public projects
	PubPC = "public_project_count"
	// PubRC : count of public repositories
	PubRC = "public_repo_count"
	// TPC : total count of projects
	TPC = "total_project_count"
	// TRC : total count of repositories
	TRC = "total_repo_count"
)

type StatisticsClienter interface {
	GetStatistics() (*HarborStatistics, error)
}

func (c *Client) GetStatistics() (*HarborStatistics, error) {
	path := APIStatistic

	log.Infof("%s %s", http.MethodGet, path)
	resp, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 == 2 {
		ret := &HarborStatistics{}
		err := json.Unmarshal(body, &ret)
		if err != nil {
			return nil, ErrorUnknownInternal.Error(err)
		}
		return ret, nil
	}

	log.Errorf("list harbor statistics error: %s", body)

	return nil, ErrorUnknownInternal.Error(body)
}
