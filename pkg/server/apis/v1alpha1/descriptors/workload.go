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
		Path:        "/runningpods",
		Description: "Workload APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.ListRunningPods,
				Description: "List running pods",
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
