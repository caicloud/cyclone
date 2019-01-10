// +nirvana:api=descriptors:"Descriptor"

package apis

import (
	"github.com/caicloud/cyclone/pkg/server/apis/middlewares"
	v1alpha1 "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1/descriptors"
	v1 "github.com/caicloud/cyclone/pkg/server/apis/v1/descriptors"

	def "github.com/caicloud/nirvana/definition"
)

// Descriptor returns a combined descriptor for APIs of all versions.
func Descriptor() def.Descriptor {
	return def.Descriptor{
		Description: "APIs",
		Path:        "/apis",
		Middlewares: middlewares.Middlewares(),
		Consumes:    []string{def.MIMEJSON},
		Produces:    []string{def.MIMEJSON},
		Children: []def.Descriptor{
			v1alpha1.Descriptor(),
			v1.Descriptor()
		},
	}
}
