package descriptors

import (
	"fmt"

	def "github.com/caicloud/nirvana/definition"

	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1/middlewares"
)

// descriptors describe APIs of current version.
var descriptors []def.Descriptor

// register registers descriptors.
func register(ds ...def.Descriptor) {
	descriptors = append(descriptors, ds...)
}

// Descriptor returns a combined descriptor for current version.
func Descriptor() def.Descriptor {
	return def.Descriptor{
		Description: fmt.Sprintf("%s APIs", v1alpha1.APIVersion),
		Path:        fmt.Sprintf("/%s", v1alpha1.APIVersion),
		Middlewares: middlewares.Middlewares(),
		Children:    descriptors,
	}
}
