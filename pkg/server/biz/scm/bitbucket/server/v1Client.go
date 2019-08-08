package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-querystring/query"
)

// BitBucket Server API docs: https://developer.atlassian.com/server/bitbucket/reference/rest-api/ .

// AuthType represents an authentication type within BitBucket Server.
type AuthType int

// List of available authentication types.
const (
	// BasicAuth represents basic authentication type.
	BasicAuth AuthType = iota
	// PersonalAccessToken represents personal access token type.
	PersonalAccessToken
)

// V1Client manages communication with the BitBucket Server API.
type V1Client struct {
	// HTTP client used to communicate with the API.
	client *http.Client

	// Base URL for API requests. Defaults to the public BitBucket Server API, but can be
	// set to a domain endpoint to use with a self hosted BitBucket Server server. baseURL
	// should always be specified with a trailing slash.
	baseURL *url.URL

	// Auth type used to make authenticated API calls.
	authType AuthType

	// Username and password used for basic authentication.
	username string
	password string

	// Token used to make authenticated API calls.
	token string

	// User agent used when communicating with the BitBucket Server API.
	UserAgent string

	// Services used for talking to different parts of the BitBucket Server API.
	PullRequests   *PullRequestsService
	Repositories   *RepositoriesService
	Authorizations *AuthorizationsService
}

// Config contains V1Client config information.
type Config struct {
	// Base URL for API requests. Defaults to the public BitBucket Server API, but can be
	// set to a domain endpoint to use with a self hosted BitBucket Server server. baseURL
	// should always be specified with a trailing slash.
	BaseURL string

	// Auth type used to make authenticated API calls.
	AuthType AuthType

	// Username used for basic authentication.
	Username string
	// Password used for basic authentication.
	Password string

	// Token used to make authenticated API calls.
	Token string
}

// ListOpts specifies the optional parameters to various List methods that support pagination.
type ListOpts struct {
	Start *int `url:"start,omitempty" json:"start,omitempty"`
	Limit *int `url:"limit,omitempty" json:"limit,omitempty"`
}

// NewClient returns a BitBucket Server client.
func NewClient(client *http.Client, config Config) (*V1Client, error) {
	if config.AuthType == BasicAuth && config.Username == "" {
		return nil, fmt.Errorf("The username is required in the bitbucket server ")
	}
	base, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(base.Path, "/") {
		base.Path += "/"
	}
	if client == nil {
		client = http.DefaultClient
	}
	v1Client := &V1Client{
		client:    client,
		baseURL:   base,
		authType:  config.AuthType,
		token:     config.Token,
		username:  config.Username,
		password:  config.Password,
		UserAgent: "continuous-integration/cyclone",
	}

	// initialize services
	v1Client.PullRequests = &PullRequestsService{v1Client: v1Client}
	v1Client.Repositories = &RepositoriesService{v1Client: v1Client}
	v1Client.Authorizations = &AuthorizationsService{v1Client: v1Client}
	return v1Client, nil
}

// NewRequest creates an API request.
func (c *V1Client) NewRequest(method, urlStr string, body interface{}, opt interface{}) (*http.Request, error) {
	u, err := c.baseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	if opt != nil {
		q, err := query.Values(opt)
		if err != nil {
			return nil, err
		}
		u.RawQuery = q.Encode()
	}

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	switch c.authType {
	case BasicAuth:
		req.SetBasicAuth(c.username, c.password)
	case PersonalAccessToken:
		req.Header.Add("Authorization", "Bearer "+c.token)
	}
	return req, nil
}

// Do sends an API request and returns the API response.
func (c *V1Client) Do(request *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(request)
	if err != nil {
		return resp, err
	}

	// check response
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return resp, fmt.Errorf("status: %v, Body: %s", resp.Status, string(bodyBytes))
	}
	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}
	return resp, err
}

// Pagination represents BitBucket Server pagination properties
// embedded in list responses.
type Pagination struct {
	Start    *int  `json:"start"`
	Size     *int  `json:"size"`
	Limit    *int  `json:"limit"`
	LastPage *bool `json:"isLastPage"`
	NextPage *int  `json:"nextPageStart"`
}

// SelfLink represents the link of the resource.
type SelfLink struct {
	Href string `json:"href"`
}

// CloneLink represents the link of the repo used to clone.
type CloneLink struct {
	Href string `json:"href"`
	Name string `json:"name"`
}
