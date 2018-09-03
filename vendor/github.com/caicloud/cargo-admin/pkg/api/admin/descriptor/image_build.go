package descriptor

import (
	"github.com/caicloud/cargo-admin/pkg/api/admin/handlers"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"

	"github.com/caicloud/nirvana/definition"
)

func init() {
	register(imageBuild)
}

var imageBuild = definition.Descriptor{
	Description: "Build image from Dockerfile and push to project",
	Children: []definition.Descriptor{
		{
			Path:        "/dockerfiles",
			Definitions: []definition.Definition{listDockerfiles},
		},
		{
			Path:        "/registries/{registry}/projects/{project}/builds",
			Consumes:    []string{definition.MIMEOctetStream},
			Definitions: []definition.Definition{buildImage},
		},
	},
}

var listDockerfiles = definition.Definition{
	Method:      definition.List,
	Description: "List Dockerfile templates",
	Function:    handlers.ListDockerfiles,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
	},
	Results: definition.DataErrorResults("Dockerfile templates list"),
}

var buildImage = definition.Definition{
	Method:      definition.Create,
	Description: "Build image from Dockerfile and push to project",
	Function:    handlers.BuildImage,
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
			Source:      definition.Header,
			Name:        types.HeaderTag,
			Description: "image tag to be built",
		},
		{
			Source:      definition.Body,
			Description: "Image build context",
		},
	},
	Results: []definition.Result{{Destination: definition.Error}},
}
