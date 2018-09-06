package descriptor

import (
	"github.com/caicloud/nirvana/definition"

	"github.com/caicloud/cyclone/pkg/api/v1/handler"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

func init() {
	register(events...)
}

var events = []definition.Descriptor{
	{
		Path:        "/events/{eventid}",
		Description: "Event API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetEvent,
				Description: "Get event by id",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.EventPathParameterName,
					},
				},
				Results: definition.DataErrorResults("event"),
			},
			{
				Method:      definition.Update,
				Function:    handler.SetEvent,
				Description: "Set the event by id",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.EventPathParameterName,
					},
				},
				Results: definition.DataErrorResults("event"),
			},
		},
	},
}
