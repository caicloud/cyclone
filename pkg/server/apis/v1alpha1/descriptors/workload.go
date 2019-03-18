package descriptors

import (
	"github.com/caicloud/nirvana/definition"

	handler "github.com/caicloud/cyclone/pkg/server/handler/v1alpha1"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

func init() {
	register(workload...)
}

var workload = []definition.Descriptor{
	{
		Path:        "/workingpods",
		Description: "Workload APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.ListWorkingPods,
				Description: "List all pods of workflowruns",
				Parameters: []definition.Parameter{
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source:      definition.Auto,
						Name:        httputil.PaginationAutoParameter,
						Description: "pagination",
					},
				},
				Results: definition.DataErrorResults("tenant"),
			},
		},
	},
}
