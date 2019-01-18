package descriptors

import (
	"github.com/caicloud/nirvana/definition"

	handler "github.com/caicloud/cyclone/pkg/server/handler/v1alpha1"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

func init() {
	register(integration...)
}

var integration = []definition.Descriptor{
	{
		Path:        "/integrations",
		Description: "Integration APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.List,
				Function:    handler.ListIntegrations,
				Description: "List integrations",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Name of the tenant whose integrations to list",
					},
					{
						Source:      definition.Auto,
						Name:        "pagination",
						Description: "pagination",
					},
				},
				Results: definition.DataErrorResults("integration"),
			},
			{
				Method:      definition.Create,
				Function:    handler.CreateIntegration,
				Description: "Create integration",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Name of the tenant to create integration for",
					},
					{
						Source:      definition.Body,
						Description: "JSON body to describe the new integration",
					},
				},
				Results: definition.DataErrorResults("created integration"),
			},
		},
	},
	{
		Path:        "/integrations/{integration}",
		Description: "Integrations APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetIntegration,
				Description: "Get integration",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Name of the tenant whose integration to get",
					},
					{
						Source:      definition.Path,
						Name:        "integration",
						Description: "Name of the integration to get",
					},
				},
				Results: definition.DataErrorResults("integration gotten"),
			},
			{
				Method:      definition.Update,
				Function:    handler.UpdateIntegration,
				Description: "Update integration",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Name of the tenant whose integration to update",
					},
					{
						Source:      definition.Path,
						Name:        "integration",
						Description: "Name of the integration to update",
					},
					{
						Source:      definition.Body,
						Description: "JSON body to describe the updated integration",
					},
				},
				Results: definition.DataErrorResults("integration updated"),
			},
			{
				Method:      definition.Delete,
				Function:    handler.DeleteIntegration,
				Description: "Delete integration",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Name of the tenant whose integration to delete",
					},
					{
						Source:      definition.Path,
						Name:        "integration",
						Description: "Name of the integration to delete",
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
}
