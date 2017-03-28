package storage

import (
	"github.com/docker/notary/storage/rethinkdb"
)

// These consts are the index names we've defined for RethinkDB
const (
	rdbSha256Idx        = "sha256"
	rdbGunRoleIdx       = "gun_role"
	rdbGunRoleSha256Idx = "gun_role_sha256"
)

var (
	// TUFFilesRethinkTable is the table definition of notary server's TUF metadata files
	TUFFilesRethinkTable = rethinkdb.Table{
		Name:       RDBTUFFile{}.TableName(),
		PrimaryKey: "gun_role_version",
		SecondaryIndexes: map[string][]string{
			rdbSha256Idx:         nil,
			"gun":                nil,
			"timestamp_checksum": nil,
			rdbGunRoleIdx:        {"gun", "role"},
			rdbGunRoleSha256Idx:  {"gun", "role", "sha256"},
		},
		// this configuration guarantees linearizability of individual atomic operations on individual documents
		Config: map[string]string{
			"write_acks": "majority",
		},
		JSONUnmarshaller: rdbTUFFileFromJSON,
	}
)
