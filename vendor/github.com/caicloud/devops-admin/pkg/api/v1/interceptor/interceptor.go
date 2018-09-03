/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package interceptor

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	log "github.com/golang/glog"

	"github.com/caicloud/devops-admin/pkg/api/v1"
	"github.com/caicloud/devops-admin/pkg/manager"
	httputil "github.com/caicloud/devops-admin/pkg/util/http"
)

// IsWorkspaceOperation judges whether the request is a workspace operation.
func IsWorkspaceOperation(requestPath string) bool {
	pathPattern := "^/api/v1/workspaces(/[^/]*)?/?$"

	re := regexp.MustCompile(pathPattern)
	return re.MatchString(requestPath)
}

// IsSpecifiedWorkspaceOperation judges whether the request is to operate specified workspace.
func IsSpecifiedWorkspaceOperation(requestPath string) bool {
	pathPattern := "^/api/v1/workspaces/[^/]+/?$"

	re := regexp.MustCompile(pathPattern)
	return re.MatchString(requestPath)
}

// HandleWorkspace handles the workspace request. Some requests need to be proxied to Cyclone.
func HandleWorkspace(response http.ResponseWriter, request *http.Request, proxy http.Handler) {
	requestPath := strings.TrimSuffix(request.URL.Path, "/")
	requestMethod := request.Method
	notSupported := false

	log.Infof("Proxy the workspace request: %s %s", requestMethod, requestPath)

	tenant, err := httputil.GetTenant(request)
	if err != nil {
		log.Errorf("Fail to get tenant from request header as %s", err.Error())
		httputil.ResponseWithError(response, err)
		return
	}

	if requestPath == "/api/v1/workspaces" {
		switch requestMethod {
		case http.MethodGet:
			start, limit, err := httputil.GetPagination(request)
			if err != nil {
				httputil.ResponseWithError(response, err)
				return
			}

			workspaces, total, err := manager.ListWorkspaces(tenant, start, limit)
			if err != nil {
				httputil.ResponseWithError(response, err)
				return
			}

			httputil.ResponseWithList(response, workspaces, total)
		case http.MethodPost:
			workspace := &v1.Workspace{}
			if err := httputil.GetJsonPayload(request, workspace); err != nil {
				httputil.ResponseWithError(response, err)
				return
			}

			workspace.Tenant = tenant
			workspace.CycloneProject = fmt.Sprintf("%s_%s", tenant, workspace.Name)

			err = manager.CreateWorkspace(workspace)
			if err != nil {
				httputil.ResponseWithError(response, err)
				return
			}

			if err = adaptProxyURL(request, tenant, workspace.Name); err != nil {
				httputil.ResponseWithError(response, err)
				return
			}

			type Project struct {
				Name        string
				Description string
			}
			project := Project{
				Name:        workspace.CycloneProject,
				Description: workspace.Description,
			}
			if err = httputil.SetJsonPayload(request, project); err != nil {
				httputil.ResponseWithError(response, err)
				return
			}

			fmt.Printf("request url: %s", request.URL.Path)
			proxy.ServeHTTP(response, request)
		default:
			notSupported = true
		}
	}

	if IsSpecifiedWorkspaceOperation(requestPath) {
		name := requestPath[strings.LastIndex(requestPath, "/")+1:]

		switch requestMethod {
		case http.MethodGet:
			workspace, err := manager.GetWorkspace(tenant, name)
			if err != nil {
				httputil.ResponseWithError(response, err)
				return
			}

			httputil.ResponseWithHeaderAndEntity(response, http.StatusOK, workspace)
		case http.MethodPatch:
			workspace := &v1.Workspace{}
			if err := httputil.GetJsonPayloadAndKeepState(request, workspace); err != nil {
				httputil.ResponseWithError(response, err)
				return
			}

			workspace, err := manager.UpdateWorkspace(tenant, name, workspace)
			if err != nil {
				httputil.ResponseWithError(response, err)
				return
			}

			if err := adaptProxyURL(request, tenant, name); err != nil {
				httputil.ResponseWithError(response, err)
				return
			}

			// Cylone only supports PUT method to update project.
			request.Method = http.MethodPut
			proxy.ServeHTTP(response, request)
		case http.MethodDelete:
			if err := adaptProxyURL(request, tenant, name); err != nil {
				httputil.ResponseWithError(response, err)
				return
			}

			if err := manager.DeleteWorkspace(tenant, name); err != nil {
				httputil.ResponseWithError(response, err)
				return
			}

			proxy.ServeHTTP(response, request)
		default:
			notSupported = true
		}
	}

	if notSupported {
		// TODO (robin) Return 405: Method Not Allowed
		err = fmt.Errorf("method %s not allowed for request path %s", requestMethod, requestPath)
		log.Errorf(err.Error())

		httputil.ResponseWithError(response, err)
	}
}

// IsPipelineOperation judges whether the request is to operate pipelines and pipeline records.
func IsPipelineOperation(requestPath string) bool {
	pathPattern := "^/api/v1/workspaces/[^/]+/pipelines((\\?[^/?]*)?|(/[^/]*)?|/[^/]+" +
		"/records((\\?[^/?]*)?|(/[^/]*)?|/[^/]+/status))/?$"

	re := regexp.MustCompile(pathPattern)
	return re.MatchString(requestPath)
}

// HandleWorkspace handles the request for pipelines and pipeline records. Some requests need to be proxied to Cyclone.
func HandlePipeline(response http.ResponseWriter, request *http.Request, proxy http.Handler) {
	requestPath := strings.TrimSuffix(request.URL.Path, "/")
	pathParts := strings.Split(requestPath, "/")
	name := pathParts[4]

	log.Infof("Proxy the pipeline request: %s %s", request.Method, requestPath)

	tenant, err := httputil.GetTenant(request)
	if err != nil {
		log.Errorf("Fail to get tenant from request header as %s", err.Error())
		httputil.ResponseWithError(response, err)
		return
	}

	if err = adaptProxyURL(request, tenant, name); err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	proxy.ServeHTTP(response, request)
}

// adaptProxyURL adapts the request URL before proxied to Cyclone.
func adaptProxyURL(request *http.Request, tenant, name string) error {
	workspace, err := manager.GetWorkspace(tenant, name)
	if err != nil {
		log.Errorf("fail to get workspace %s in tenant %s as %s", name, tenant, err.Error())
		return err
	}

	requestPath := strings.TrimSuffix(request.URL.Path, "/")
	pathParts := strings.Split(requestPath, "/")
	pathPartsLength := len(pathParts)
	if pathPartsLength > 4 {
		pathParts[4] = workspace.CycloneProject
	}

	pathParts[3] = "projects"
	request.URL.Path = strings.Join(pathParts, "/")

	return nil
}
