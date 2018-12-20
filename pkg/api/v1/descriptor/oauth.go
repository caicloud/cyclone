package descriptor

import (
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/operators/validator"

	"github.com/caicloud/cyclone/pkg/api/v1/handler"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

func init() {
	register(oauth...)
}

var oauth = []definition.Descriptor{
	{
		Path:        "/scms/{scm}/code",
		Description: " Oauth API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetAuthCodeURL,
				Description: "Get request URL and call it to get oauth code",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.SCM,
					},
				},
				Results: definition.DataErrorResults("get oauth code URL"),
			},
		},
	},
	{
		Path:        "/scms/{scm}/callback",
		Description: "Oauth API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetToken,
				Description: "Handle the callback of oauth",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.SCM,
					},
					{
						Source:    definition.Query,
						Name:      "code",
						Operators: []definition.Operator{validator.String("required")},
					},
					{
						Source:    definition.Query,
						Name:      "state",
						Operators: []definition.Operator{validator.String("required")},
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
}
