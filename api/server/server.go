/*
Copyright 2016 caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/golang/glog"

	"github.com/caicloud/cyclone/cmd/worker/options"
	"github.com/caicloud/cyclone/pkg/api/v1/descriptor"
	"github.com/caicloud/cyclone/pkg/cloud"
	"github.com/caicloud/cyclone/pkg/event"
	"github.com/caicloud/cyclone/pkg/store"
	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/plugins/metrics"
	"github.com/caicloud/nirvana/plugins/profiling"
	"github.com/zoumo/logdog"

	"github.com/caicloud/cyclone/pkg/api/v1/handler"
)

const (
	DefaultPort = 7099
)

// APIServer ...
type APIServer struct {
	Config        *APIServerOptions
	WorkerOptions *options.WorkerOptions
}

// PrepareRun prepare for apiserver running
func (s *APIServer) PrepareRun() (*PreparedAPIServer, error) {
	closing := make(chan struct{})

	// init database
	mclosed, err := store.Init(s.Config.MongoDBHost, s.Config.MongoGracePeriod, closing, s.Config.SaltKey)
	if err != nil {
		return nil, err
	}

	go background(closing, mclosed)

	// init event manager
	event.Init(s.WorkerOptions)

	if err = cloud.InitCloud(s.Config.CloudAutoDiscovery); err != nil {
		log.Error(err)
		return nil, err
	}

	return &PreparedAPIServer{s}, nil
}

// PreparedAPIServer is a prepared api server
type PreparedAPIServer struct {
	*APIServer
}

// Run start a api server
func (s *PreparedAPIServer) Run(stopCh <-chan struct{}) error {
	dataStore := store.NewStore()
	defer dataStore.Close()

	// Initialize the V1 API Handler.
	if err := handler.InitHandler(dataStore, s.Config.RecordRotationThreshold); err != nil {
		logdog.Fatal(err)
		return err
	}

	port := DefaultPort
	if s.Config.CyclonePort != 0 {
		port = s.Config.CyclonePort
	}

	log.Infof("cyclone starts listening on %v", port)
	log.Infof("cyclone starts listening on %v", s.Config.CycloneAddrTemplate)

	config := nirvana.NewDefaultConfig()
	//nirvana.IP(f.Address)(config)
	nirvana.Port(uint16(port))(config)
	config.Configure(
		metrics.Path("/metrics"),
		profiling.Path("/debug/pprof/"),
		profiling.Contention(true),
	)

	config.Configure(nirvana.Descriptor(descriptor.Descriptor()))

	log.Infof("API service listening on %s:%d", s.Config.CycloneAddrTemplate, port)
	if err := nirvana.NewServer(config).Serve(); err != nil {
		log.Fatal(err)
	}

	log.Error("Server stopped")
	<-stopCh
	return nil
}

// background must be a daemon goroutine for Cyclone server.
// It can catch system signal and send signal to other goroutine before program exits.
func background(closing, mclosed chan struct{}) {
	closed := []chan struct{}{mclosed}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigs
	log.Info("capture system signal, will close \"closing\" channel")
	close(closing)
	for _, c := range closed {
		<-c
	}
	log.Info("exit the process with 0")
	os.Exit(0)
}
