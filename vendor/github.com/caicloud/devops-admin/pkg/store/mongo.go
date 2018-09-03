/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package store

import (
	"fmt"
	"strings"
	"time"

	log "github.com/golang/glog"
	"gopkg.in/mgo.v2"
)

const (
	socketTimeout  = time.Second * 5
	syncTimeout    = time.Second * 5
	dialTimeout    = time.Second * 10
	tickerDuration = time.Second * 5

	// Read preference modes are specific to mgo:
	Eventual  int = 0 // Same as Nearest, but may change servers between reads.
	Monotonic int = 1 // Same as SecondaryPreferred before first write. Same as Primary after first write.
	Strong    int = 2 // Same as Primary.

	// defaultDatabase represents the default name of database.
	defaultDatabase = "devops-admin"

	// workspaceCollectionName represents the name of the collection for workspace.
	workspaceCollectionName = "workspaces"
)

var g_modes = map[string]int{
	"eventual":  0,
	"monotonic": 1,
	"mono":      1,
	"strong":    2,
}

var session *mgo.Session
var mclosed chan struct{}

// Init mongo session
func InitMongo(conf *MgoConfig, closing chan struct{}) (chan struct{}, error) {
	mclosed = make(chan struct{})
	var err error
	info := &mgo.DialInfo{
		Addrs:    conf.Addrs,
		Database: conf.DB,
		Timeout:  dialTimeout,
	}
	if len(info.Database) == 0 {
		info.Database = defaultDatabase
	}

	session, err = mgo.DialWithInfo(info)
	if err != nil {
		log.Errorf("Connect mongodb failed: %v", err)
		return nil, err
	}
	EnsureSafe(session, conf.Safe)
	if err = SetMode(session, conf.Mode, true); err != nil {
		return nil, err
	}
	session.SetSocketTimeout(socketTimeout)
	session.SetSyncTimeout(syncTimeout)

	go backgroundMongo(closing)

	DB := session.DB(conf.DB)
	Workspace = &_Workspace{Collection: DB.C(workspaceCollectionName)}
	if err = ensureIndexes(); err != nil {
		return nil, err
	}

	return mclosed, nil
}

func ensureIndexes() error {
	return Workspace.EnsureIndexes()
}

// Ensure session safe mode
func EnsureSafe(s *mgo.Session, safe *Safe) {
	if safe == nil {
		return
	}
	s.EnsureSafe(&mgo.Safe{
		W:        safe.W,
		WMode:    safe.WMode,
		WTimeout: safe.WTimeout,
		FSync:    safe.FSync,
		J:        safe.J,
	})
}

// Set mongo safe mode
func SetMode(s *mgo.Session, modeFriendly string, refresh bool) error {
	mode, ok := g_modes[strings.ToLower(modeFriendly)]
	if !ok {
		return fmt.Errorf("invalid mgo mode")
	}
	switch mode {
	case Eventual:
		s.SetMode(mgo.Eventual, refresh)
	case Monotonic:
		s.SetMode(mgo.Monotonic, refresh)
	case Strong:
		s.SetMode(mgo.Strong, refresh)
	}

	return nil
}
