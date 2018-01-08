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
	"time"

	"k8s.io/client-go/pkg/util/wait"

	"github.com/caicloud/cyclone/api/rest"
	"github.com/caicloud/cyclone/cloud"
	"github.com/caicloud/cyclone/etcd"
	"github.com/caicloud/cyclone/event"
	loghttp "github.com/caicloud/cyclone/http"
	"github.com/caicloud/cyclone/kafka"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/server/router"
	newstore "github.com/caicloud/cyclone/pkg/store"
	"github.com/caicloud/cyclone/store"
	"github.com/caicloud/cyclone/websocket"
	restful "github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"github.com/zoumo/logdog"
	mgo "gopkg.in/mgo.v2"
)

// APIServer ...
type APIServer struct {
	Config        *APIServerOptions
	WorkerOptions *cloud.WorkerOptions
	dbSession     *mgo.Session
}

// PrepareRun prepare for apiserver running
func (s *APIServer) PrepareRun() (*PreparedAPIServer, error) {

	s.InitLog()
	cloud.Debug = s.Config.Debug
	logdog.Debugf("Debug mode: %t", s.Config.Debug)

	// init api doc
	if s.Config.ShowAPIDoc {
		// Open http://localhost:7099/apidocs and enter http://localhost:7099/apidocs.json in the api input field.
		config := swagger.Config{
			WebServices:    restful.RegisteredWebServices(), // you control what services are visible.
			WebServicesUrl: fmt.Sprintf(s.Config.CycloneAddrTemplate, s.Config.CyclonePort),
			ApiPath:        "/apidocs.json",

			// Optionally, specify where the UI is located.
			SwaggerPath:     "/apidocs/",
			SwaggerFilePath: "./node_modules/swagger-ui/dist",
		}
		swagger.InstallSwaggerService(config)
	}

	// init database
	err := s.InitStore()
	if err != nil {
		return nil, err
	}

	// init event manager
	err = s.initEventManager()
	if err != nil {
		return nil, err
	}

	// init rest api server
	rest.Initialize()

	return &PreparedAPIServer{s}, nil
}

// InitLog initializes log
func (s *APIServer) InitLog() {
	if s.Config.LogForceColor {
		logdog.ForceColor = s.Config.LogForceColor
	}

	// init debug log
	if s.Config.Debug {
		logdog.ApplyOptions(logdog.DebugLevel)
		log.SetLogLevel(log.DebugLevel)
	} else {
		logdog.ApplyOptions(logdog.InfoLevel)
	}
}

// InitStore init database connection. fix me: change function name
func (s *APIServer) InitStore() error {
	// init mongodb
	var err error

	// dail mongo session
	// wait mongodb set up
	wait.Poll(time.Second, s.Config.MongoGracePeriod, func() (bool, error) {
		s.dbSession, err = mgo.Dial(s.Config.MongoDBHost)
		return err == nil, nil
	})
	if err != nil {
		logdog.Fatalf("Unable connect to mongodb addr %s", s.Config.MongoDBHost)
		return err
	}

	logdog.Debugf("connect to mongodb addr: %s", s.Config.MongoDBHost)
	s.dbSession.SetMode(mgo.Strong, true)

	// TODO (robin) Remove the old store and use the new one.
	store.Init(s.dbSession)
	newstore.Init(s.dbSession)

	return nil
}

// FIXME
func (s *APIServer) initEventManager() error {
	etcd.Init([]string{s.Config.ETCDHost})
	event.Init(s.WorkerOptions, s.Config.CloudAutoDiscovery)

	return nil
}

// PreparedAPIServer is a prepared api server
type PreparedAPIServer struct {
	*APIServer
}

// Run start a api server
func (s *PreparedAPIServer) Run(stopCh <-chan struct{}) error {

	// TODO: PostStartHooks

	// start log server
	go s.StartLogServer()

	dataStore := newstore.NewStore()
	defer dataStore.Close()

	// Initialize the V1 API.
	if err := router.InitRouters(dataStore); err != nil {
		logdog.Fatal(err)
		return err
	}

	// start server
	server := &http.Server{Addr: fmt.Sprintf(":%d", s.Config.CyclonePort), Handler: restful.DefaultContainer}
	logdog.Infof("cyclone server listening on %d", s.Config.CyclonePort)
	logdog.Fatal(server.ListenAndServe())
	// <-stopCh
	return nil
}

// StartLogServer run a http log server and a websocket server
func (s *PreparedAPIServer) StartLogServer() {

	// FIXME: start loghttp server
	go loghttp.Server()

	// start websocket log server
	err := kafka.Dail([]string{s.Config.KafkaHost})
	if nil != err {
		log.Error(err.Error())
	}
	defer kafka.Close()

	websocket.LoadServerConfig()
	err = websocket.StartServer()
	if err != nil {
		log.Error(err.Error())
	}
}
