package harbor

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	. "github.com/caicloud/cargo-admin/pkg/errors"

	"github.com/caicloud/nirvana/log"
)

type VolumeClienter interface {
	GetVolumes() (*HarborVolumes, error)
}

func (c *Client) GetVolumes() (*HarborVolumes, error) {
	path := VolumesPath()

	log.Infof("%s %s", http.MethodGet, path)
	resp, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}

	if resp.StatusCode/100 == 2 {
		ret := &HarborVolumes{}
		err := json.Unmarshal(body, ret)
		if err != nil {
			return nil, ErrorUnknownInternal.Error(err)
		}
		return ret, nil
	}

	log.Errorf("get volumes info error (%d): %s", resp.StatusCode, body)
	return nil, ErrorUnknownInternal.Error(body)
}
