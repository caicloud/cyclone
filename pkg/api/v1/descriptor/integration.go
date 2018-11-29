package descriptor

import (
	"github.com/caicloud/nirvana/definition"

	"github.com/caicloud/cyclone/pkg/api/v1/handler"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

func init() {
	register(integrations...)
}

var integrations = []definition.Descriptor{
	{
		Path:        "/integrations",
		Description: "Integration API",
		Definitions: []definition.Definition{
			{
				Method:      definition.List,
				Function:    handler.ListIntegrations,
				Description: "List all integrations",
				Results:     definition.DataErrorResults("integrations"),
			},
			{
				Method:      definition.Create,
				Function:    handler.CreateIngegration,
				Description: "Add integration",
				Results:     definition.DataErrorResults("integration"),
			},
		},
	},
	{
		Path:        "/integrations/{integration}",
		Description: "Integration API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Delete,
				Function:    handler.DeleteIntegration,
				Description: "Delete the integration",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.IntegrationPathParameterName,
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
			{
				Method:      definition.Update,
				Function:    handler.UpdateIntegration,
				Description: "Update the integration",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.IntegrationPathParameterName,
					},
				},
				Results: definition.DataErrorResults("integration"),
			},
			{
				Method:      definition.Get,
				Function:    handler.GetIntegration,
				Description: "Get the integration",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.IntegrationPathParameterName,
					},
				},
				Results: definition.DataErrorResults("integration"),
			},
		},
	},
}
