package model

import (
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/caicloud/cyclone/pkg/server/config"
	"github.com/caicloud/nirvana/log"
)

const (
	socketTimeout  = time.Second * 5
	syncTimeout    = time.Second * 5
	tickerDuration = time.Second * 5

	// Read preference modes are specific to mgo:
	Eventual  int = 0 // Same as Nearest, but may change servers between reads.
	Monotonic int = 1 // Same as SecondaryPreferred before first write. Same as Primary after first write.
	Strong    int = 2 // Same as Primary.
)

var session *mgo.Session
var gModes = map[string]int{
	"eventual":  0,
	"monotonic": 1,
	"mono":      1,
	"strong":    2,
}

func initMongo(mongoConf config.MongoConfig, closing chan struct{}) (*mgo.Database, chan struct{}, error) {
	closed := make(chan struct{})
	var err error
	session, err = mgo.Dial(mongoConf.Addrs)
	if err != nil {
		log.Fatalf("connect mongodb failed, %s\n", err)
		return nil, nil, err
	}

	ensureSafe(session, mongoConf.Safe)
	setMode(session, mongoConf.Mode, true)
	session.SetSocketTimeout(socketTimeout)
	session.SetSyncTimeout(syncTimeout)

	go backgroundMongo(closing, closed)
	return session.DB(mongoConf.DB), closed, nil
}

func InitMongo(mongoConf config.MongoConfig, closing chan struct{}) (chan struct{}, error) {
	db, closed, err := initMongo(mongoConf, closing)
	if err != nil {
		log.Fatalf("init mongodb failed, %s\n", err)
		return nil, err
	}

	Vulnerability = &VulnerabilityInfo{Collection: db.C("vulnerability")}
	return closed, nil
}

// Ensure session safe mode
func ensureSafe(s *mgo.Session, safe *config.Safe) {
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
func setMode(s *mgo.Session, modeFriendly string, refresh bool) {
	mode, ok := gModes[strings.ToLower(modeFriendly)]
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

// Background goroutine for mongo. It can hold mongo connection & close session when progress exit.
func backgroundMongo(closing chan struct{}, closed chan struct{}) {
	ticker := time.NewTicker(tickerDuration)
	for {
		select {
		case <-ticker.C:
			if err := session.Ping(); err != nil {
				log.Errorf("session error: %v", err)
				session.Refresh()
				session.SetSocketTimeout(socketTimeout)
				session.SetSyncTimeout(syncTimeout)
			}
		case <-closing:
			session.Close()
			log.Info("Mongodb session has been closed")
			close(closed)
			return
		}
	}
}
