package descriptor

import (
	"github.com/caicloud/nirvana/definition"

	"github.com/caicloud/cyclone/pkg/api/v1/handler"
)

func init() {
	register(templates...)
}

var templates = []definition.Descriptor{
	{
		Path:        "/configtemplates",
		Description: "Pipeline Config Template API",
		Definitions: []definition.Definition{
			{
				Method:      definition.List,
				Function:    handler.ListConfigTemplates,
				Description: "Get all cyclone built-in pipeline config templates",
				Results: []definition.Result{
					{
						Destination: definition.Data,
						Description: "config templates",
					},
				},
			},
		},
	},
}
