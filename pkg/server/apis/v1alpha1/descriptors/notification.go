package descriptors

import (
	"github.com/caicloud/nirvana/definition"

	handler "github.com/caicloud/cyclone/pkg/server/handler/v1alpha1"
)

func init() {
	register(notification...)
}

var notification = []definition.Descriptor{
	{
		Path:        "/notifications/workflowruns",
		Description: "Notification APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Create,
				Function:    handler.ReceiveWorkflowRunNotification,
				Description: "Handle notifications about workflowruns from workflow controller",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Body,
						Name:        "workflowrun",
						Description: "workflowrun",
					},
				},
				Results: definition.DataErrorResults("notification"),
			},
		},
	},
}
