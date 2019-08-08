package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/caicloud/nirvana/log"

	"github.com/caicloud/cyclone/pkg/server/biz/scm"
)

// RepositoriesService handles communication with the repository related.
type RepositoriesService struct {
	v1Client *V1Client
}

// Repository contains data from a BitBucket Server Repository.
type Repository struct {
	Slug          string  `json:"slug"`
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	ScmID         string  `json:"scmId"`
	State         string  `json:"state"`
	StatusMessage string  `json:"statusMessage"`
	Forkable      bool    `json:"forkable"`
	Project       Project `json:"project"`
	Public        bool    `json:"public"`
	Links         struct {
		Clone []CloneLink `json:"clone"`
		Self  []SelfLink  `json:"self"`
	} `json:"links"`
}

// Project contains data from a BitBucket Server Project.
type Project struct {
	Key         string `json:"key"`
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Public      bool   `json:"public"`
	Type        string `json:"type"`
	Links       struct {
		Self []SelfLink `json:"self"`
	} `json:"links"`
}

// Repositories is a set of repositories.
type Repositories struct {
	Pagination
	Values []Repository `json:"values"`
}

// Branch contains git Branch information.
type Branch struct {
	ID           string `json:"id"`
	DisplayID    string `json:"displayId"`
	Type         string `json:"type"`
	LatestCommit string `json:"latestCommit"`
	IsDefault    bool   `json:"isDefault"`
}

// Branches is a set of branches.
type Branches struct {
	Pagination
	Values []Branch `json:"values"`
}

// Tag contains git Tag information.
type Tag struct {
	ID           string `json:"id"`
	DisplayID    string `json:"displayId"`
	Type         string `json:"type"`
	LatestCommit string `json:"latestCommit"`
	Hash         string `json:"hash"`
}

// Tags is a set of tags.
type Tags struct {
	Pagination
	Values []Tag `json:"values"`
}

// PullRequests is a set of PullRequest.
type PullRequests struct {
	Pagination
	Values []scm.PullRequest `json:"values"`
}

// Files is a set of files' name in a repo.
type Files struct {
	Pagination
	Values []string `json:"values"`
}

// StatusReq represents the options of creating commit status.
type StatusReq struct {
	State       string `json:"state"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

// ListRepositories list repositories in the Bitbucket Server.
func (server *RepositoriesService) ListRepositories(ctx context.Context, project string, opt *ListOpts) (*Repositories, *http.Response, error) {
	var u string
	if len(project) == 0 {
		u = "rest/api/1.0/repos"
	} else {
		u = fmt.Sprintf("rest/api/1.0/projects/%s/repos", project)
	}

	req, err := server.v1Client.NewRequest(http.MethodGet, u, nil, opt)
	if err != nil {
		return nil, nil, err
	}
	var repos Repositories
	resp, err := server.v1Client.Do(req, &repos)
	return &repos, resp, err
}

// ListBranches list branches on the repository.
func (server *RepositoriesService) ListBranches(ctx context.Context, project string, repo string, opt *ListOpts) (*Branches, *http.Response, error) {
	u := fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s/branches", project, repo)
	req, err := server.v1Client.NewRequest(http.MethodGet, u, nil, opt)
	if err != nil {
		return nil, nil, err
	}
	var branches Branches
	resp, err := server.v1Client.Do(req, &branches)
	log.Infof("branch: %+v, error: %+v", branches, err)
	return &branches, resp, err
}

// ListFiles list files on the repository.
func (server *RepositoriesService) ListFiles(ctx context.Context, project string, repo string, opt *ListOpts) (*Files, *http.Response, error) {
	u := fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s/files", project, repo)
	req, err := server.v1Client.NewRequest(http.MethodGet, u, nil, opt)
	if err != nil {
		return nil, nil, err
	}
	var files Files
	resp, err := server.v1Client.Do(req, &files)
	return &files, resp, err
}

// ListTags list tags on the repository.
func (server *RepositoriesService) ListTags(ctx context.Context, project string, repo string, opt *ListOpts) (*Tags, *http.Response, error) {
	u := fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s/tags", project, repo)
	req, err := server.v1Client.NewRequest(http.MethodGet, u, nil, opt)
	if err != nil {
		return nil, nil, err
	}
	var tags Tags
	resp, err := server.v1Client.Do(req, &tags)
	return &tags, resp, err
}

// ListPullRequests list pull requests on the repository.
func (server *RepositoriesService) ListPullRequests(ctx context.Context, project, repo string, opt *PullRequestListOpts) (*PullRequests, *http.Response, error) {
	// Detailed info: https://docs.atlassian.com/bitbucket-server/rest/6.5.1/bitbucket-rest.html#idp254
	u := fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s/pull-requests", project, repo)
	req, err := server.v1Client.NewRequest(http.MethodGet, u, nil, opt)
	if err != nil {
		return nil, nil, err
	}

	var prs PullRequests
	resp, err := server.v1Client.Do(req, &prs)
	return &prs, resp, err
}

// CreateStatus create a commit status.
func (server *RepositoriesService) CreateStatus(ctx context.Context, commitID string, input *StatusReq) (*http.Response, error) {
	u := fmt.Sprintf("rest/build-status/1.0/commits/%s", commitID)
	req, err := server.v1Client.NewRequest(http.MethodPost, u, input, nil)
	if err != nil {
		return nil, err
	}
	resp, err := server.v1Client.Do(req, nil)
	return resp, err
}
