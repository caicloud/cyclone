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
		Path:        "/pipelinetemplates",
		Description: "Template API",
		Definitions: []definition.Definition{
			{
				Method:      definition.List,
				Function:    handler.ListTemplates,
				Description: "Get all cyclone built-in pipeline templates",
				Results: []definition.Result{
					{
						Destination: definition.Data,
						Description: "templates",
					},
				},
			},
		},
	},
}
