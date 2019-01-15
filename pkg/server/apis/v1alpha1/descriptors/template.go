package descriptors

import (
	"github.com/caicloud/nirvana/definition"

	"github.com/caicloud/cyclone/pkg/server/handler"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

func init() {
	register(template...)
}

var template = []definition.Descriptor{
	{
		Path:        "/templates",
		Description: "Stage template APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.ListTemplates,
				Description: "List stage templates",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Name of the tenant whose templates to list",
					},
					{
						Source:      definition.Query,
						Name:        httputil.IncludePublicQueryParameter,
						Default:     true,
						Description: "Whether include system level stage templates",
					},
					{
						Source:      definition.Auto,
						Name:        "pagination",
						Description: "pagination",
					},
				},
				Results: definition.DataErrorResults("stage templates"),
			},
			{
				Method:      definition.Create,
				Function:    handler.CreateTemplate,
				Description: "Create stage template",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Name of the tenant whose templates to list",
					},
					{
						Source:      definition.Body,
						Description: "JSON body to describe the new template",
					},
				},
				Results: definition.DataErrorResults("created stage template"),
			},
		},
	},
	{
		Path:        "/templates/{template}",
		Description: "Stage template APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetTemplate,
				Description: "Get stage template",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Name of the tenant whose templates to list",
					},
					{
						Source:      definition.Path,
						Name:        "template",
						Description: "Name of the stage template to get",
					},
				},
				Results: definition.DataErrorResults("stage template gotten"),
			},
			{
				Method:      definition.Update,
				Function:    handler.UpdateTemplate,
				Description: "Update stage template",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Name of the tenant whose templates to list",
					},
					{
						Source:      definition.Path,
						Name:        "template",
						Description: "Name of the stage template to get",
					},
					{
						Source:      definition.Body,
						Description: "JSON body to describe the updated template",
					},
				},
				Results: definition.DataErrorResults("stage template updated"),
			},
			{
				Method:      definition.Delete,
				Function:    handler.DeleteTemplate,
				Description: "Delete stage template",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Name of the tenant whose templates to list",
					},
					{
						Source:      definition.Path,
						Name:        "template",
						Description: "Name of the stage template to get",
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
}
