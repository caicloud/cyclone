package harbor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	. "github.com/caicloud/cargo-admin/pkg/errors"

	"github.com/caicloud/nirvana/log"
)

type ProjectClienter interface {
	CreateProject(name string, isPublic bool) (int64, error)
	ListProjects(page, pageSize int, name, public string) (int, []*HarborProject, error)
	AllProjects(name, public string) ([]*HarborProject, error)
	GetProject(pid int64) (*HarborProject, error)
	DeleteProject(pid int64) error
	GetProjectDeleteable(pid int64) (*HarborProjectDeletableResp, error)
	GetRepoCount(pid int64) (int, error)
}

func (c *Client) CreateProject(name string, isPublic bool) (int64, error) {
	path := APIProjects
	hcpReq := &harborCreateProjectReq{
		Name:     name,
		Metadata: map[string]string{ProMetaPublic: strconv.FormatBool(isPublic)},
	}
	b, err := json.Marshal(hcpReq)
	if err != nil {
		return 0, ErrorUnknownInternal.Error(err)
	}

	log.Infof("%s %s", http.MethodPost, path)
	log.Infof("create project request body: %s", b)
	resp, err := c.do(http.MethodPost, path, bytes.NewReader(b))
	if err != nil {
		return 0, ErrorUnknownInternal.Error(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 {
		redirect := resp.Header.Get("Location")
		log.Infof("create project success, redirect url in resp header: %s", redirect)
		pidStr := strings.TrimPrefix(redirect, path+"/")
		log.Infof("redirect is :%s", redirect)
		log.Infof("pidStr is :%s", pidStr)
		pid, err := strconv.ParseInt(pidStr, 10, 64)
		log.Infof("pid is :%d", pid)
		if err != nil {
			log.Errorf("the returned projectId can not be ParseInt: %v", err)
			return 0, ErrorUnknownInternal.Error(err)
		}
		log.Infof("create project success, projectId: %d", pid)
		return pid, nil
	}

	if resp.StatusCode == 409 {
		log.Errorf("harbor return 409 error: %s", body)
		return 0, ErrorAlreadyExist.Error(fmt.Sprintf("project %v", name))
	}
	log.Errorf("harbor return unexcepted statusCode: %d, resp body: %s", resp.StatusCode, body)

	return 0, ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}

// Get a page of projects that match the given parameters:
// - name: Project with name that contains substring given by the parameter 'name'. If set to empty string, all project
// names match.
// - public: Whether project is a public project, if set to 'true', only public project will match, if set to 'false',
// only private projects match, and if set to empty string, both private and public projects match.
func (c *Client) ListProjects(page, pageSize int, name, public string) (int, []*HarborProject, error) {
	path := ProjectsPath(page, pageSize, name, public)

	log.Infof("%s %s", http.MethodGet, path)
	resp, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		return 0, nil, ErrorUnknownInternal.Error(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 == 2 {
		ret := make([]*HarborProject, 0)
		err := json.Unmarshal(body, &ret)
		if err != nil {
			return 0, nil, ErrorUnknownInternal.Error(err)
		}
		total, err := getTotalFromResp(resp)
		if err != nil {
			log.Errorf("get total from resp error: %v", err)
			return 0, nil, ErrorUnknownInternal.Error(err)
		}
		return total, ret, nil
	}
	log.Errorf("list harbor projects error: %s", body)

	return 0, nil, ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}

// Get all projects matching the given parameters.
// - name: Project with name that contains substring given by the parameter 'name'. If set to empty string, all project
// names match.
// - public: Whether project is a public project, if set to 'true', only public project will match, if set to 'false',
// only private projects match, and if set to empty string, both private and public projects match.
// Here are some examples:
// * Get all projects: AllProjects("", "")
// * Get all public projects: AllProjects("", "true")
// * Get all private projects whose names include "devops": AllProjects("devops", "false")
func (c *Client) AllProjects(name, public string) ([]*HarborProject, error) {
	page := 1
	ret := make([]*HarborProject, 0)
	for {
		total, projects, err := c.ListProjects(page, MaxPageSize, name, public)
		if err != nil {
			return nil, err
		}
		ret = append(ret, projects...)
		if total <= page*MaxPageSize {
			break
		}
		page++
	}
	return ret, nil
}

func (c *Client) ProjectExist(pid int64) (bool, error) {
	path := ProjectPath(pid)

	log.Infof("%s %s", http.MethodGet, path)
	resp, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		return false, ErrorUnknownInternal.Error(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 5 {
		return false, ErrorUnknownInternal.Error(body)
	}

	return resp.StatusCode/100 == 2, nil
}

func (c *Client) ProjectExistByName(name string) (bool, error) {
	path := ProjectExist(name)

	log.Infof("%s %s", http.MethodHead, path)
	resp, err := c.do(http.MethodHead, path, nil)
	if err != nil {
		return false, ErrorUnknownInternal.Error(err)
	}

	return resp.StatusCode/100 == 2, nil
}

func (c *Client) GetProject(pid int64) (*HarborProject, error) {
	path := ProjectPath(pid)

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
		ret := &HarborProject{}
		err := json.Unmarshal(body, &ret)
		if err != nil {
			return nil, ErrorUnknownInternal.Error(err)
		}

		if count, err := c.GetRepoCount(pid); err == nil {
			ret.RepoCount = count
		} else {
			log.Warningf("get repo count for project %s error: %v", pid, err)
		}

		return ret, nil
	}
	if resp.StatusCode == 404 {
		return nil, ErrorContentNotFound.Error(fmt.Sprintf("project: %d", pid))
	}
	log.Errorf("get harbor project '%d' error: %s", pid, body)

	return nil, ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}

func (c *Client) GetProjectByName(name string) (*HarborProject, error) {
	projects, err := c.AllProjects(name, "")
	if err != nil {
		return nil, err
	}
	for _, p := range projects {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, nil
}

func (c *Client) DeleteProject(pid int64) error {
	path := ProjectPath(pid)

	log.Infof("%s %s", http.MethodDelete, path)
	resp, err := c.do(http.MethodDelete, path, nil)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		log.Infof("delete project: %d sucess", pid)
		return nil
	}
	if resp.StatusCode == 404 {
		return ErrorContentNotFound.Error(fmt.Sprintf("project: %d", pid))
	}
	log.Errorf("delete harbor project: %d error: %s", pid, body)

	return ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}

func (c *Client) GetProjectDeleteable(pid int64) (*HarborProjectDeletableResp, error) {
	path := ProjectDeletablePath(pid)

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
		ret := &HarborProjectDeletableResp{}
		err := json.Unmarshal(body, &ret)
		if err != nil {
			return nil, ErrorUnknownInternal.Error(err)
		}
		return ret, nil
	}

	log.Errorf("check project: %d whether is deleteable error: %s", pid, body)
	return nil, ErrorUnknownInternal.Error(fmt.Sprintf("%s", body))
}

func (c *Client) GetRepoCount(pid int64) (int, error) {
	path := ReposPath(pid, "", 1, 1)
	resp, err := c.do(http.MethodGet, path, nil)
	defer resp.Body.Close()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(resp.Header.Get(RespHeaderTotal))
}
