package descriptor

import (
	"github.com/caicloud/cargo-admin/pkg/api/admin/handlers"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"

	"github.com/caicloud/nirvana/definition"
)

func init() {
	register(replications)
}

var replications = definition.Descriptor{
	Description: "Replication API",
	Children: []definition.Descriptor{
		{
			Path:        "/replications",
			Definitions: []definition.Definition{createReplication, listReplications},
		},
		{
			Path:        "/replications/{replication}",
			Definitions: []definition.Definition{getReplication, updateReplication, deleteReplication},
		},
		{
			Path:        "/replications/{replication}/action",
			Definitions: []definition.Definition{triggerReplication},
		},
	},
}

var createReplication = definition.Definition{
	Method:      definition.Create,
	Description: "Create replication",
	Function:    handlers.CreateReplication,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Body,
			Description: "create replication request body",
		},
	},
	Results: definition.DataErrorResults("replication"),
}

var listReplications = definition.Definition{
	Method:      definition.List,
	Description: "List replications",
	Function:    handlers.ListReplications,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderSeqID,
			Description: "sequence id",
		},
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Query,
			Name:        "direction",
			Description: "which kind of direction",
		},
		{
			Source:      definition.Query,
			Name:        "registry",
			Description: "registry name",
		},
		{
			Source:      definition.Query,
			Name:        "project",
			Description: "project name",
		},
		{
			Source:      definition.Query,
			Name:        "triggerType",
			Description: "replicaiton trigger type",
		},
		{
			Source:      definition.Query,
			Name:        "q",
			Description: "query keyword",
		},
		{
			Source:      definition.Auto,
			Name:        "pagination",
			Description: "pagination",
		},
	},
	Results: []definition.Result{
		{
			Destination: definition.Data,
			Description: "replication list",
		},
		{
			Destination: definition.Meta,
		},
		{
			Destination: definition.Error,
		},
	},
}

var getReplication = definition.Definition{
	Method:      definition.Get,
	Description: "Get replication",
	Function:    handlers.GetReplication,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Path,
			Name:        "replication",
			Description: "replication name",
		},
	},
	Results: definition.DataErrorResults("replication"),
}

var updateReplication = definition.Definition{
	Method:      definition.Update,
	Description: "Update replication",
	Function:    handlers.UpdateReplication,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Path,
			Name:        "replication",
			Description: "replication name",
		},
		{
			Source:      definition.Body,
			Description: "update replication request body",
		},
	},
	Results: definition.DataErrorResults("replication"),
}

var deleteReplication = definition.Definition{
	Method:      definition.Delete,
	Description: "Delete replication",
	Function:    handlers.DeleteReplication,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Path,
			Name:        "replication",
			Description: "replication name",
		},
	},
	Results: []definition.Result{{Destination: definition.Error}},
}

var triggerReplication = definition.Definition{
	Method:      definition.Create,
	Description: "Trigger replication",
	Function:    handlers.TriggerReplication,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Path,
			Name:        "replication",
			Description: "replication name",
		},
		{
			Source:      definition.Body,
			Description: "trigger replication request body",
		},
	},
	Results: []definition.Result{{Destination: definition.Error}},
}
