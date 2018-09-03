/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package manager

import (
	"time"

	"gopkg.in/mgo.v2"

	"github.com/caicloud/devops-admin/pkg/api/v1"
	. "github.com/caicloud/devops-admin/pkg/errors"
	"github.com/caicloud/devops-admin/pkg/store"
)

// CreateWorkspace creates a workspace.
func CreateWorkspace(workspace *v1.Workspace) error {
	var err error
	if len(workspace.Tenant) == 0 {
		return ErrorValidationFailed.Format("workspace tenant", "can not be empty")
	}

	if len(workspace.Name) == 0 {
		return ErrorValidationFailed.Format("workspace name", "can not be empty")
	}

	if _, err := GetWorkspace(workspace.Tenant, workspace.Name); err == nil {
		return ErrorAlreadyExist.Format(workspace.Name)
	}

	if err = store.Workspace.Save(workspace); err != nil {
		return err
	}

	return nil
}

// GetWorkspace gets the workspace by name in on tenant.
func GetWorkspace(tenant, name string) (*v1.Workspace, error) {
	workspace, err := store.Workspace.FindByName(tenant, name)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, ErrorContentNotFound.Format(name)
		}

		return nil, err
	}

	return &workspace, nil
}

// ListWorkspaces list workspaces in one tenant.
func ListWorkspaces(tenant string, start, limit int) ([]v1.Workspace, int, error) {
	return store.Workspace.FindOnePage(tenant, start, limit)
}

// UpdateWorkspace updates the workspace by name in on tenant.
func UpdateWorkspace(tenant, name string, newWorkspace *v1.Workspace) (*v1.Workspace, error) {
	workspace, err := GetWorkspace(tenant, name)
	if err != nil {
		return nil, err
	}

	// Only support to update the workspace description.
	workspace.Description = newWorkspace.Description
	workspace.LastUpdateTime = time.Now().Format(time.RFC3339)

	err = store.Workspace.UpdateId(workspace.ID, workspace)
	if err != nil {
		return nil, err
	}

	return workspace, nil
}

// DeleteWorkspace deletes the workspace by name in on tenant.
func DeleteWorkspace(tenant, name string) error {
	return store.Workspace.Delete(tenant, name)
}
