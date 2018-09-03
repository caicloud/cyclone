package descriptor

import (
	"github.com/caicloud/cargo-admin/pkg/api/middlewares/logger"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
)

const headerCargo = "X-Cargo"

var descriptors = []definition.Descriptor{}

func register(ds ...definition.Descriptor) {
	descriptors = append(descriptors, ds...)
}

// Configure APIs for token service and health check
func ConfigService(s *nirvana.Config) {
	middlewares := []definition.Middleware{
		logger.New(log.DefaultLogger()),
	}

	v2 := definition.Descriptor{
		Description: "token service",
		Path:        "/",
		Children:    descriptors,
		Middlewares: middlewares,
		Consumes:    []string{definition.MIMEJSON},
		Produces:    []string{definition.MIMEJSON},
	}

	s.Configure(nirvana.Descriptor(v2))
}
