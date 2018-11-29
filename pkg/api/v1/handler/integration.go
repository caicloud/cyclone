package handler

import (
	"context"

	"github.com/caicloud/cyclone/pkg/api"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

// CreateIngegration handles the request to create ingegration.
func CreateIngegration(ctx context.Context) (*api.Integration, error) {
	integration := &api.Integration{}
	err := contextutil.GetJsonPayload(ctx, integration)
	if err != nil {
		return nil, err
	}

	createdIntegration, err := integrationManager.CreateIntegration(integration)
	if err != nil {
		return nil, err
	}

	return createdIntegration, nil
}

// ListIntegrations handles the request to list all ingegrations.
func ListIntegrations(ctx context.Context) (api.ListResponse, error) {
	integrations, err := integrationManager.ListIntegrations()
	if err != nil {
		return api.ListResponse{}, nil
	}

	return httputil.ResponseWithList(integrations, len(integrations)), nil
}

// DeleteIntegration handles the request to delete the ingegration.
func DeleteIntegration(ctx context.Context, name string) error {
	if err := integrationManager.DeleteIntegration(name); err != nil {
		return err
	}

	return nil
}

// UpdateIntegration handles the request to update the ingegration.
func UpdateIntegration(ctx context.Context, name string) (*api.Integration, error) {
	integration := &api.Integration{}
	err := contextutil.GetJsonPayload(ctx, integration)
	if err != nil {
		return nil, err
	}

	updatedIntegration, err := integrationManager.UpdateIntegration(name, integration)
	if err != nil {
		return nil, err
	}

	return updatedIntegration, nil
}

// GetIntegration handles the request to get the ingegration.
func GetIntegration(ctx context.Context, name string) (*api.Integration, error) {
	return integrationManager.GetIntegration(name)
}
