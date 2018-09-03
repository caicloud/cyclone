package harbor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"

	. "github.com/caicloud/cargo-admin/pkg/errors"

	"github.com/caicloud/nirvana/log"
)

type TagClienter interface {
	ListTags(projectName string, repoName string) ([]*HarborTag, error)
	ListTagNames(projectName, repoName string) ([]string, error)
	GetTag(projectName, repoName, tag string) (*HarborTag, error)
	GetTagVulnerabilities(projectName, repoName, tag string) ([]*HarborVulnerability, error)
	DeleteTag(projectName, repoName, tag string) error
}

func (c *Client) ListTags(projectName string, repoName string) ([]*HarborTag, error) {
	path := TagsPath(projectName, repoName)

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
		tags := make([]*HarborTag, 0)
		err := json.Unmarshal(body, &tags)
		if err != nil {
			log.Errorf("unmarshal tags error: %v", err)
			log.Infof("resp body: %s", body)
			return nil, ErrorUnknownInternal.Error(err)
		}
		sort.Sort(TagsSortByDateDes(tags))
		return tags, nil
	}

	return nil, ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}

func (c *Client) ListTagNames(projectName, repoName string) ([]string, error) {
	tags, err := c.ListTags(projectName, repoName)
	if err != nil {
		return nil, err
	}

	tagStr := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagStr = append(tagStr, tag.Name)
	}

	return tagStr, nil
}

func (c *Client) GetTag(projectName, repoName, tag string) (*HarborTag, error) {
	path := TagPath(projectName, repoName, tag)

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

	log.Infof("%s", body)

	if resp.StatusCode/100 == 2 {
		ret := &HarborTag{}
		err := json.Unmarshal(body, ret)
		if err != nil {
			log.Errorf("unmarshal tag error: %v", err)
			log.Infof("resp body: %s", body)
			return &HarborTag{}, nil
		}
		return ret, nil
	}

	return nil, ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}

func (c *Client) GetTagVulnerabilities(projectName, repoName, tag string) ([]*HarborVulnerability, error) {
	path := TagVulnerabilityPath(projectName, repoName, tag)

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
		ret := make([]*HarborVulnerability, 0)
		err := json.Unmarshal(body, &ret)
		if err != nil {
			return nil, ErrorUnknownInternal.Error(err)
		}
		return ret, nil
	}

	return nil, ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}

func (c *Client) DeleteTag(projectName, repoName, tag string) error {
	path := TagPath(projectName, repoName, tag)

	log.Infof("%s %s", http.MethodGet, path)
	resp, err := c.do(http.MethodDelete, path, nil)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}
	if resp.StatusCode/100 == 2 {
		return nil
	}

	return ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}
