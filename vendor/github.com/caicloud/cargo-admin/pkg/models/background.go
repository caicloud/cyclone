package models

import (
	"time"

	"github.com/caicloud/nirvana/log"
)

const (
	tickerDuration = time.Second * 5
)

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
