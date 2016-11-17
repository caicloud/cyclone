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

package main

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/api/rest"
	"github.com/caicloud/cyclone/etcd"
	"github.com/caicloud/cyclone/event"
	cyclonehttp "github.com/caicloud/cyclone/http"
	"github.com/caicloud/cyclone/kafka"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/pkg/wait"
	"github.com/caicloud/cyclone/store"
	"github.com/caicloud/cyclone/websocket"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"github.com/spf13/pflag"
	"gopkg.in/mgo.v2"
)

// Here, we define all flags used in cyclone.
var (
	debug  = pflag.Bool("debug", false, "Debug mode, default to false")
	apiDoc = pflag.Bool("show-api-doc", true, "show the api doc at http://<cyclone instance>/apidocs/#/api/v0.1")

	consoleWebEndpoint string
)

const (
	MONGO_DB_IP      = "MONGO_DB_IP"
	DOCKER_HOST      = "DOCKER_HOST"
	DOCKER_CERT_PATH = "DOCKER_CERT_PATH"

	// The docker registry to pull&push images, and its login username & password.
	REGISTRY_LOCATION = "REGISTRY_LOCATION"
	REGISTRY_USERNAME = "REGISTRY_USERNAME"
	REGISTRY_PASSWORD = "REGISTRY_PASSWORD"

	// Do we want to enable caicloud authserver, default to 'true'. This is used
	// for debugging/testing. Note this is used for caicloud auth, not docker
	// registry - we always deal with registry auth in cyclone.
	ENABLE_CAICLOUD_AUTH = "ENABLE_CAICLOUD_AUTH"

	// Where the docker hosts exsit
	//DOCKER_HOST_DIR = "DOCKER_HOST_DIR"
	// Console web endpoint environment.
	//CONSOLE_WEB_ENDPOINT = "CONSOLE_WEB_ENDPOINT"

	// Env variables name about EmailNotifier.
	//SMTP_SERVER   = "SMTP_SERVER"
	//SMTP_PORT     = "SMTP_PORT"
	//SMTP_USERNAME = "SMTP_USERNAME"
	//SMTP_PASSWORD = "SMTP_PASSWORD"

	KAFKA_SERVER_IP = "KAFKA_SERVER_IP"

	ETCD_SERVER_IP = "ETCD_SERVER_IP"
)

const (
	// DefaultRegistry is the default docker registry, if user don't have REGISTRY_LOCATION defined,
	// cyclone would use DefaultRegistry instead.
	DefaultRegistry = "cargo.caicloud.io"

	// MongoGracePeriod is the grace period waiting for mongodb to be running.
	MongoGracePeriod = 30 * time.Second

	cyclonePort            = 7099
	cycloneAddressTemplate = "http://localhost:%v"
)

func main() {
	log.Info("Cyclone server start")
	pflag.Parse()

	// init system
	setLogLevel()
	setMaxProcesses()

	// init DB
	var session *mgo.Session
	dailMongo(session)
	defer session.Close()

	// init event manager
	initEventManger()

	// init log server
	go initLogServer()
	go cyclonehttp.Server()

	// init api
	initAPIServer()
	initAPIDoc()
	startAPIServer()
}

// setLogLevel set the log level by input flag
// if exist --debug flag the log level will be Debug,
// else the log under Debug level will not output.
func setLogLevel() {
	// set log level to debug level
	if *debug == true {
		log.SetLogLevel(log.DebugLevel)
		log.Debug("Debug mode: True")
	}
}

// setMaxProcesses set man process with the CPU number.
func setMaxProcesses() {
	//set max processes
	nCPUNumber := runtime.NumCPU()
	runtime.GOMAXPROCS(nCPUNumber)
}

// dailMongo dail mongo server by ENV.
func dailMongo(session *mgo.Session) {
	// Get the IP of mongodb
	mongoIP := osutil.GetStringEnv(MONGO_DB_IP, "localhost")
	// Create mongo session with grace period.
	var err error

	err = wait.Poll(time.Second, MongoGracePeriod, func() (done bool, err error) {
		session, err = mgo.Dial(mongoIP)
		return err == nil, nil
	})
	if err != nil {
		log.Fatal("Error dailing mongodb")
	}

	session.SetMode(mgo.Strong, true)

	// Init data store
	store.Init(session)
}

// initAPIServer init restful api server.
func initAPIServer() {
	// Get docker deamon's endpoint and cert path.
	//endpoint := osutil.MustGetStringEnv(DOCKER_HOST, "unix:///var/run/docker.sock")

	// Get option for enabling caicloud auth.
	enableCaicloudAuth := osutil.GetStringEnv(ENABLE_CAICLOUD_AUTH, "false")

	// Initialize rest endpoints and all cyclone managers.
	rest.Initialize(enableCaicloudAuth)

}

// initAPIDoc init api doc server according to the input flag.
func initAPIDoc() {
	if *apiDoc == true {
		// Open http://localhost:7099/apidocs and enter http://localhost:7099/apidocs.json in the api input field.
		config := swagger.Config{
			WebServices:    restful.RegisteredWebServices(), // you control what services are visible.
			WebServicesUrl: fmt.Sprintf(cycloneAddressTemplate, cyclonePort),
			ApiPath:        "/apidocs.json",

			// Optionally, specify where the UI is located.
			SwaggerPath:     "/apidocs/",
			SwaggerFilePath: "./node_modules/swagger-ui/dist",
		}
		swagger.InstallSwaggerService(config)
	}
}

// startAPIServer start api server.
func startAPIServer() {
	// Start listening on port 7099 for incomming connections.
	server := &http.Server{Addr: fmt.Sprintf(":%d", cyclonePort), Handler: restful.DefaultContainer}
	log.Infof("cyclone server listening on %d", cyclonePort)
	log.Fatal(server.ListenAndServe())
}

// initLogServer init log server.
func initLogServer() {
	kafkaIP := osutil.GetStringEnv(KAFKA_SERVER_IP, "127.0.0.1:9092")
	err := kafka.Dail([]string{kafkaIP})
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

// initEventManger init event manager.
func initEventManger() {
	etcdIP := osutil.GetStringEnv(ETCD_SERVER_IP, "http://127.0.0.1:2379")
	etcd.Init([]string{etcdIP})

	certPath := osutil.GetStringEnv(DOCKER_CERT_PATH, "")

	// Get the username and password to access the docker registry.
	registryLocation := osutil.GetStringEnv(REGISTRY_LOCATION, DefaultRegistry)
	registryUsername := osutil.GetStringEnv(REGISTRY_USERNAME, "")
	registryPassword := osutil.GetStringEnv(REGISTRY_PASSWORD, "")

	registry := api.RegistryCompose{
		RegistryLocation: registryLocation,
		RegistryUsername: registryUsername,
		RegistryPassword: registryPassword,
	}
	event.Init(certPath, registry)
}
