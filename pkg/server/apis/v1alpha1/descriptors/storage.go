package descriptors

import (
	"github.com/caicloud/nirvana/definition"

	handler "github.com/caicloud/cyclone/pkg/server/handler/v1alpha1"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

func init() {
	register(usages...)
}

var usages = []definition.Descriptor{
	{
		Path:        "/storage/usages",
		Description: "Storage usages APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Create,
				Function:    handler.ReportStorageUsage,
				Description: "Report storage usages",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.NamespaceHeaderName,
						Description: "Namespace",
					},
					{
						Source:      definition.Body,
						Description: "JSON body to describe the usages",
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
			{
				Method:      definition.Get,
				Function:    handler.GetStorageUsage,
				Description: "Get storage usages",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Tenant name",
					},
				},
				Results: definition.DataErrorResults("storage usage"),
			},
		},
	},
	{
		Path:        "/storage/cleanup",
		Description: "Clean storage paths",
		Definitions: []definition.Definition{
			{
				Method:      definition.Create,
				Function:    handler.Cleanup,
				Description: "Clean storage paths",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Name of the tenant whose storage to clean",
					},
					{
						Source:      definition.Body,
						Description: "JSON body to describe paths to clean",
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
}
