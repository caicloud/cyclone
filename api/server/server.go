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
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	restful "github.com/emicklei/go-restful"
	swagger "github.com/emicklei/go-restful-swagger12"
	log "github.com/golang/glog"
	"github.com/zoumo/logdog"

	"github.com/caicloud/cyclone/cmd/worker/options"
	"github.com/caicloud/cyclone/pkg/cloud"
	"github.com/caicloud/cyclone/pkg/event"
	"github.com/caicloud/cyclone/pkg/server/router"
	"github.com/caicloud/cyclone/pkg/store"
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

	// Initialize the V1 API.
	if err := router.InitRouters(dataStore, s.Config.RecordRotationThreshold); err != nil {
		logdog.Fatal(err)
		return err
	}

	// init api doc
	if s.Config.ShowAPIDoc {
		// Open http://localhost:7099/apidocs and enter http://localhost:7099/apidocs.json in the api input field.
		config := swagger.Config{
			WebServices:    restful.DefaultContainer.RegisteredWebServices(), // you control what services are visible.
			WebServicesUrl: fmt.Sprintf(s.Config.CycloneAddrTemplate, s.Config.CyclonePort),
			ApiPath:        "/apidocs.json",

			// Optionally, specify where the UI is located.
			SwaggerPath:     "/apidocs/",
			SwaggerFilePath: "./node_modules/swagger-ui/dist",
		}
		swagger.InstallSwaggerService(config)
	}

	// start server
	server := &http.Server{Addr: fmt.Sprintf(":%d", s.Config.CyclonePort), Handler: restful.DefaultContainer}
	logdog.Infof("cyclone server listening on %d", s.Config.CyclonePort)
	logdog.Fatal(server.ListenAndServe())
	// <-stopCh
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
