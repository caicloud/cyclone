package descriptors

import (
	"github.com/caicloud/nirvana/definition"

	handler "github.com/caicloud/cyclone/pkg/server/handler/v1alpha1"
)

func init() {
	register(webhook...)
}

var webhook = []definition.Descriptor{
	{
		Path:        "/tenants/{tenant}/webhook",
		Description: "Webhook APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Create,
				Function:    handler.HandleWebhook,
				Description: "Handle webhooks from integrated systems",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Path,
						Name:        "tenant",
						Description: "tenant",
					},
					{
						Source:      definition.Query,
						Name:        "eventType",
						Description: "the webhook eventType, support SCM for now",
					},
					{
						Source: definition.Query,
						Name:   "integration",
					},
				},
				Results: definition.DataErrorResults("webhook"),
			},
		},
	},
}
