package descriptor

import (
	"github.com/caicloud/cargo-admin/pkg/api/admin/handlers"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"

	"github.com/caicloud/nirvana/definition"
)

func init() {
	register(imageUpload)
}

var imageUpload = definition.Descriptor{
	Description: "Upload image via tarball API",
	Children: []definition.Descriptor{
		{
			Path:        "/registries/{registry}/projects/{project}/uploads",
			Consumes:    []string{definition.MIMEOctetStream},
			Definitions: []definition.Definition{uploadImage},
		},
	},
}

var uploadImage = definition.Definition{
	Method:      definition.Create,
	Description: "Upload Image Tarball",
	Function:    handlers.UploadImage,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Path,
			Name:        "registry",
			Description: "registry name",
		},
		{
			Source:      definition.Path,
			Name:        "project",
			Description: "project name",
		},
		{
			Source:      definition.Body,
			Description: "Image tarball content",
		},
	},
	Results: []definition.Result{
		{
			Destination: definition.Data,
			Description: "Image upload result",
		},
		{
			Destination: definition.Error,
		},
	},
}
