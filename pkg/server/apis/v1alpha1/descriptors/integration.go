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
						Source:      definition.Query,
						Name:        httputil.IncludePublicQueryParameter,
						Default:     true,
						Description: "Whether include system level resources",
					},
					{
						Source:      definition.Auto,
						Name:        httputil.PaginationAutoParameter,
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
						Source: definition.Header,
						Name:   httputil.PublicHeaderName,
					},
					{
						Source:      definition.Body,
						Description: "JSON body to describe the new integration",
					},
					{
						Source:      definition.Header,
						Name:        httputil.HeaderDryRun,
						Default:     false,
						Description: "Whether to do a rehearsal of creating integration",
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
					{
						Source: definition.Header,
						Name:   httputil.PublicHeaderName,
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
						Source: definition.Header,
						Name:   httputil.PublicHeaderName,
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
					{
						Source: definition.Header,
						Name:   httputil.PublicHeaderName,
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
	{
		Path:        "/integrations/{integration}/opencluster",
		Description: "Integrations APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Update,
				Function:    handler.OpenCluster,
				Description: "Open cluster type integration to execute workflow",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Name of the tenant whose integration to operate",
					},
					{
						Source:      definition.Path,
						Name:        "integration",
						Description: "Name of the integration to operate",
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
	{
		Path:        "/integrations/{integration}/closecluster",
		Description: "Integrations APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Update,
				Function:    handler.CloseCluster,
				Description: "Close cluster type integration that used to execute workflow",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Name of the tenant whose integration to operate",
					},
					{
						Source:      definition.Path,
						Name:        "integration",
						Description: "Name of the integration to operate",
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
	{
		Path:        "/integrations/{integration}/scmrepos",
		Description: "Integrations APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.ListSCMRepos,
				Description: "List repos for integrated SCM",
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
				Results: definition.DataErrorResults("repos gotten for integrated SCM"),
			},
		},
	},
	{
		Path:        "/integrations/{integration}/scmrepos/{repo}/branches",
		Description: "Integrations APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.ListSCMBranches,
				Description: "List branches for integrated SCM",
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
					{
						Source:      definition.Path,
						Name:        "repo",
						Description: "Name of SCM repo",
					},
				},
				Results: definition.DataErrorResults("branches gotten for integrated SCM"),
			},
		},
	},
	{
		Path:        "/integrations/{integration}/scmrepos/{repo}/tags",
		Description: "Integrations APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.ListSCMTags,
				Description: "List tags for integrated SCM",
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
					{
						Source:      definition.Path,
						Name:        "repo",
						Description: "Name of SCM repo",
					},
				},
				Results: definition.DataErrorResults("tags gotten for integrated SCM"),
			},
		},
	},
	{
		Path:        "/integrations/{integration}/scmrepos/{repo}/dockerfiles",
		Description: "Integrations APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.ListSCMDockerfiles,
				Description: "List Dockerfiles for integrated SCM",
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
					{
						Source:      definition.Path,
						Name:        "repo",
						Description: "Name of SCM repo",
					},
				},
				Results: definition.DataErrorResults("Dockerfiles gotten for integrated SCM"),
			},
		},
	},
}
