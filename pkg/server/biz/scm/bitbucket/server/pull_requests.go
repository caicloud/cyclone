package server

import (
	"context"
	"fmt"
	"net/http"
)

// PullRequest represents a BitBucket Server pull request on a repository.
type PullRequest struct {
	ID          int    `json:"id"`
	Version     int    `json:"version"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	Open        bool   `json:"open"`
	Closed      bool   `json:"closed"`
	CreatedDate int64  `json:"createdDate"`
	UpdatedDate int64  `json:"updatedDate"`
	FromRef     struct {
		ID           string     `json:"id"`
		DisplayID    string     `json:"displayId"`
		LatestCommit string     `json:"latestCommit"`
		Repository   Repository `json:"repository"`
	} `json:"fromRef"`
	ToRef struct {
		ID           string     `json:"id"`
		DisplayID    string     `json:"displayId"`
		LatestCommit string     `json:"latestCommit"`
		Repository   Repository `json:"repository"`
	} `json:"toRef"`
	Locked bool `json:"locked"`
	Author struct {
		User struct {
			Name         string `json:"name"`
			EmailAddress string `json:"emailAddress"`
			ID           int    `json:"id"`
			DisplayName  string `json:"displayName"`
			Active       bool   `json:"active"`
			Slug         string `json:"slug"`
			Type         string `json:"type"`
			Links        struct {
				Self []struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"links"`
		} `json:"user"`
		Role     string `json:"role"`
		Approved bool   `json:"approved"`
		Status   string `json:"status"`
	} `json:"author"`
	Reviewers    []interface{} `json:"reviewers"`
	Participants []interface{} `json:"participants"`
	Links        struct {
		Self []SelfLink `json:"self"`
	} `json:"links"`
}

// PullRequestsService handles communication with the pull request related
type PullRequestsService struct {
	v1Client *V1Client
}

// GetPullRequest get a specific pull request.
func (server *PullRequestsService) GetPullRequest(ctx context.Context, project string, repo string, number int) (*PullRequest, *http.Response, error) {
	u := fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s/pull-requests/%d", project, repo, number)
	req, err := server.v1Client.NewRequest(http.MethodGet, u, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	var pr PullRequest
	resp, err := server.v1Client.Do(req, &pr)
	return &pr, resp, err
}
