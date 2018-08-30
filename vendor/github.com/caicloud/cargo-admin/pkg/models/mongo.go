package models

import (
	"strings"
	"time"

	"github.com/caicloud/cargo-admin/pkg/env"

	"gopkg.in/mgo.v2"

	"github.com/caicloud/nirvana/log"
)

const (
	socketTimeout = time.Second * 5
	syncTimeout   = time.Second * 5

	// Read preference modes are specific to mgo:
	Eventual  int = 0 // Same as Nearest, but may change servers between reads.
	Monotonic int = 1 // Same as SecondaryPreferred before first write. Same as Primary after first write.
	Strong    int = 2 // Same as Primary.
)

var g_modes = map[string]int{
	"eventual":  0,
	"monotonic": 1,
	"mono":      1,
	"strong":    2,
}

var session *mgo.Session

func initMongo(conf *env.MgoConfig, closing chan struct{}) (*mgo.Database, chan struct{}, error) {
	closed := make(chan struct{})
	var err error
	session, err = mgo.Dial(conf.Addrs)
	if err != nil {
		log.Fatal("Connect mongodb failed: ", err)
		return nil, nil, err
	}

	EnsureSafe(session, conf.Safe)
	SetMode(session, conf.Mode, true)
	session.SetSocketTimeout(socketTimeout)
	session.SetSyncTimeout(syncTimeout)

	go backgroundMongo(closing, closed)

	return session.DB(conf.DB), closed, nil
}

// Init mongo session for token service
func InitTokenMongo(conf *env.MgoConfig, closing chan struct{}) (chan struct{}, error) {
	db, closed, err := initMongo(conf, closing)
	if err != nil {
		log.Fatalf("init mongo connection error: %v", err)
		return nil, err
	}

	Project = &_Project{Collection: db.C("project")}
	DockerAccount = &_DockerAccount{Collection: db.C("docker_account")}
	Registry = &_Registry{Collection: db.C("registry")}
	return closed, nil
}

// Init mongo session for admin service
func InitAdminMongo(conf *env.MgoConfig, closing chan struct{}) (chan struct{}, error) {
	db, closed, err := initMongo(conf, closing)
	if err != nil {
		log.Fatalf("init mongo connection error: %v", err)
		return nil, err
	}

	Project = &_Project{Collection: db.C("project")}
	DockerAccount = &_DockerAccount{Collection: db.C("docker_account")}
	Registry = &_Registry{Collection: db.C("registry")}
	Repository = &_Repository{Collection: db.C("repository")}
	Replication = &_Replication{Collection: db.C("replication")}
	Record = &_Record{Collection: db.C("record")}
	RecordImage = &_RecordImage{Collection: db.C("record_image")}
	ensureIndexes()
	return closed, nil
}

func ensureIndexes() {
	Project.EnsureIndexes()
	DockerAccount.EnsureIndexes()
	Registry.EnsureIndexes()
	Repository.EnsureIndexes()
	Replication.EnsureIndexes()
	Record.EnsureIndexes()
	RecordImage.EnsureIndexes()
}

// Ensure session safe mode
func EnsureSafe(s *mgo.Session, safe *env.Safe) {
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
func SetMode(s *mgo.Session, modeFriendly string, refresh bool) {
	mode, ok := g_modes[strings.ToLower(modeFriendly)]
	if !ok {
		log.Fatal("invalid mgo mode")
		return
	}
	switch mode {
	case Eventual:
		s.SetMode(mgo.Eventual, refresh)
	case Monotonic:
		s.SetMode(mgo.Monotonic, refresh)
	case Strong:
		s.SetMode(mgo.Strong, refresh)
	}
}
