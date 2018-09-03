/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package store

// Mongo Safe struct in mgo.v2
type Safe struct {
	W        int    `json:"w"`        // Min # of servers to ack before success
	WMode    string `json:"wmode"`    // Write mode for MongoDB 2.0+ (e.g. "majority")
	WTimeout int    `json:"wtimeout"` // Milliseconds to wait for W before timing out
	FSync    bool   `json:"fsync"`    // Sync via the journal if present, or via data files sync otherwise
	J        bool   `json:"j"`        // Sync via the journal if present
}

// Mongo config
type MgoConfig struct {
	Addrs []string `json:"addrs"`
	DB    string   `json:"db"`
	Mode  string   `json:"mode"`
	Safe  *Safe    `json:"safe"`
}
