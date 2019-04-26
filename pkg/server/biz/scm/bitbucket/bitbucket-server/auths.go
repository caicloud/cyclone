package server

import (
	"context"
	"fmt"
	"net/http"
)

// AuthorizationsService handles communication with the authorization related
// methods of the BitBucket Server API.
// docs: https://docs.atlassian.com/bitbucket-server/rest/6.2.0/bitbucket-access-tokens-rest.html .
type AuthorizationsService struct {
	v1Client *V1Client
}

// PermissionReq represents the options of creating access token.
type PermissionReq struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

// PermissionResp represents the response body of creating access token.
type PermissionResp struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

const (
	// RepoAdmin represents admin permission in repo.
	RepoAdmin = "REPO_ADMIN"
	// ProjectRead represents reading permission in project.
	ProjectRead = "PROJECT_READ"
)

// CreateAccessToken create a new access token in BitBucket Server.
func (server *AuthorizationsService) CreateAccessToken(ctx context.Context, user string, permissionReq PermissionReq) (string, *http.Response, error) {
	u := fmt.Sprintf("rest/access-tokens/1.0/users/%s", user)
	req, err := server.v1Client.NewRequest(http.MethodPut, u, permissionReq, nil)
	if err != nil {
		return "", nil, err
	}
	var permissionResp PermissionResp
	resp, err := server.v1Client.Do(req, &permissionResp)
	return permissionResp.Token, resp, err
}
