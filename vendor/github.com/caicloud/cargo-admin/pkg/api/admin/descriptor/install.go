package descriptor

import (
	"github.com/caicloud/cargo-admin/pkg/api/middlewares/logger"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
)

var descriptors = []definition.Descriptor{}

func register(ds ...definition.Descriptor) {
	descriptors = append(descriptors, ds...)
}

func ConfigService(s *nirvana.Config) {
	middlewares := []definition.Middleware{
		logger.New(log.DefaultLogger()),
	}

	v1 := definition.Descriptor{
		Description: "cargo-admin API v2",
		Path:        "/api/v2",
		Children:    descriptors,
		Middlewares: middlewares,
		Consumes:    []string{definition.MIMEJSON},
		Produces:    []string{definition.MIMEJSON},
	}

	s.Configure(nirvana.Descriptor(v1))
}
