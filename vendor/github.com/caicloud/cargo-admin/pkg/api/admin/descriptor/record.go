package descriptor

import (
	"github.com/caicloud/cargo-admin/pkg/api/admin/handlers"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"

	"github.com/caicloud/nirvana/definition"
)

func init() {
	register(records)
}

var records = definition.Descriptor{
	Description: "Record API",
	Children: []definition.Descriptor{
		{
			Path:        "/replications/{replication}/records",
			Definitions: []definition.Definition{listRecords},
		},
		{
			Path:        "/replications/{replication}/records/{record}",
			Definitions: []definition.Definition{getRecord},
		},
		{
			Path:        "/replications/{replication}/records/{record}/images",
			Definitions: []definition.Definition{listRecordImages},
		},
	},
}

var listRecords = definition.Definition{
	Method:      definition.List,
	Description: "List records",
	Function:    handlers.ListRecords,
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
			Source:      definition.Query,
			Name:        "status",
			Description: "replication status",
		},
		{
			Source:      definition.Query,
			Name:        "triggerType",
			Description: "replicaiton trigger type",
		},
		{
			Source:      definition.Auto,
			Name:        "pagination",
			Description: "pagination",
		},
	},
	Results: definition.DataErrorResults("record list"),
}

var getRecord = definition.Definition{
	Method:      definition.Get,
	Description: "Get a record",
	Function:    handlers.GetRecord,
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
			Source:      definition.Path,
			Name:        "record",
			Description: "record id",
		},
	},
	Results: definition.DataErrorResults("record information"),
}

var listRecordImages = definition.Definition{
	Method:      definition.List,
	Description: "List record images",
	Function:    handlers.ListRecordImages,
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
			Source:      definition.Path,
			Name:        "record",
			Description: "record id",
		},
		{
			Source:      definition.Query,
			Name:        "status",
			Description: "replication status",
		},
		{
			Source:      definition.Auto,
			Name:        "pagination",
			Description: "pagination",
		},
	},
	Results: definition.DataErrorResults("record image list"),
}
