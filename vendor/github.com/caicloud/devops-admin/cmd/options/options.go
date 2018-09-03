/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package options

import (
	"github.com/spf13/pflag"
)

// Flags represents the flags needed to start the application.
type Flags struct {
	Address       string
	Port          int32
	CycloneServer string
	MongoServers  []string
	MongoDatabase string
}

// AddFlags adds the application flags to the flagset of pflag.
func (f *Flags) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&f.Address, "host", f.Address, "The IP address for the devops admin to serve on.")
	fs.Int32Var(&f.Port, "port", f.Port, "The port for the devops admin to serve on.")
	fs.StringVar(&f.CycloneServer, "cyclone-server", f.CycloneServer, "The address of cyclone server.")
	fs.StringSliceVar(&f.MongoServers, "mongo-servers", f.MongoServers, "Comma-separated list of Mongo servers.")
	fs.StringVar(&f.MongoDatabase, "mongo-database", f.MongoDatabase, "The database of Mongo.")
}
