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
		Path:        "/tenants/{tenant}/integrations/{integration}/webhook",
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
						Source:      definition.Path,
						Name:        "integration",
						Description: "Integration name",
					},
				},
				Results: definition.DataErrorResults("webhook"),
			},
		},
	},
}
