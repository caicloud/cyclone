package descriptor

import (
	"github.com/caicloud/cargo-admin/pkg/api/admin/handlers"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"

	"github.com/caicloud/nirvana/definition"
)

func init() {
	register(imageCopy)
}

var imageCopy = definition.Descriptor{
	Description: "Image Copy API",
	Children: []definition.Descriptor{
		{
			Path:        "/imagecopies",
			Definitions: []definition.Definition{triggerImageCopy},
		},
	},
}

var triggerImageCopy = definition.Definition{
	Method:      definition.Create,
	Description: "Trigger Image Copy",
	Function:    handlers.TriggerImageCopy,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Header,
			Name:        types.HeaderUser,
			Description: "username",
		},
		{
			Source:      definition.Body,
			Description: "trigger image copy request body",
		},
	},
	Results: []definition.Result{{Destination: definition.Error}},
}
