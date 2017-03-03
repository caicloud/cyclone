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
	"strings"
	"time"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/docker"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/websocket"
	"github.com/caicloud/cyclone/worker/ci"
	"github.com/caicloud/cyclone/worker/ci/parser"
	"github.com/caicloud/cyclone/worker/handler"
	"github.com/caicloud/cyclone/worker/helper"
	worker_log "github.com/caicloud/cyclone/worker/log"
	"github.com/caicloud/cyclone/worker/vcs"
)

const (
	WORKER_EVENTID = "WORKER_EVENTID"
	SERVER_HOST    = "SERVER_HOST"

	WORK_REGISTRY_LOCATION = "WORK_REGISTRY_LOCATION"
	REGISTRY_USERNAME      = "REGISTRY_USERNAME"
	REGISTRY_PASSWORD      = "REGISTRY_PASSWORD"
	KAFKA_SERVER_IP        = "KAFKA_SERVER_IP"
	LOG_SERVER             = "LOG_SERVER"

	WORKER_TIMEOUT = 7200 * time.Second
	WAIT_TIMES     = 5

	DefaultCaicloudYaml = "build:\n  dockerfile_name: Dockerfile"
)

func main() {
	// Get event for circe server
	event, err := getEvent()
	if err != nil {
		log.Errorf("get event err: %v", err)
		event.Status = api.EventStatusFail
		event.ErrorMessage = err.Error()
		err = sendEvent(event)
		if err != nil {
			log.Errorf("set event result err: %v", err)
		}
		return
	}

	// Handle event
	handleEvent(&event)

	// Sent event for circe server
	err = sendEvent(event)
	if err != nil {
		log.Errorf("set event result err: %v", err)
	}

	return
}

// getEvent used for getting event for circe server
func getEvent() (api.Event, error) {
	eventID := osutil.GetStringEnv(WORKER_EVENTID, "")
	serverHost := osutil.GetStringEnv(SERVER_HOST, "http://127.0.0.1:7099")

	BaseURL := fmt.Sprintf("%s/api/%s", serverHost, api.APIVersion)
	httpHandler := handler.NewHTTPHandler(BaseURL)

	var getResponse api.GetEventResponse
	err := httpHandler.GetEvent(eventID, &getResponse)
	if err != nil {
		return api.Event{}, err
	}

	event := getResponse.Event
	log.Infof("get event: %+v", event)

	return event, nil
}

// handleEvent analysize the the operation in event, and do the relate operation
func handleEvent(event *api.Event) {
	vcsManager := vcs.NewManager()

	logServer := osutil.GetStringEnv(LOG_SERVER, "ws://127.0.0.1:8000/ws")
	err := worker_log.DialLogServer(logServer)
	if nil != err {
		log.Errorf("dail log server err: %v", err)
	} else {
		go worker_log.SendHeartBeat()
		defer worker_log.Disconnect()
	}

	switch event.Operation {
	case "create-service":
		createService(vcsManager, event)

	case "create-version":
		createVersion(vcsManager, event)

	default:
		event.Status = api.EventStatusFail
		event.ErrorMessage = "unkwon operation"
		log.ErrorWithFields("Operation failed", log.Fields{"event": event})
	}
}

// createService verify repo validity for creating service
func createService(vcsManager *vcs.Manager, event *api.Event) {
	// err := vcsManager.CloneServiceRepository(event)
	err := vcsManager.CheckRepoValid(event)
	if err != nil {
		event.Status = api.EventStatusFail
		event.ErrorMessage = err.Error()
		log.ErrorWithFields("Operation failed", log.Fields{"event": event})
	} else {
		event.Status = api.EventStatusSuccess
	}
}

// createVersion create version for service and push log to server via websocket
// Step1: clone code from VCS
// Step2: integretion
// Step3: publish
// Step4: deploy
func createVersion(vcsManager *vcs.Manager, event *api.Event) {
	registryLocation := osutil.GetStringEnv(WORK_REGISTRY_LOCATION, "")
	registryUsername := osutil.GetStringEnv(REGISTRY_USERNAME, "")
	registryPassword := osutil.GetStringEnv(REGISTRY_PASSWORD, "")

	dockerManager, err := docker.NewManager("unix:///var/run/docker.sock", "",
		api.RegistryCompose{registryLocation, registryUsername, registryPassword})
	if err != nil {
		return
	}
	ciManager, err := ci.NewManager(dockerManager)
	if err != nil {
		return
	}

	err = worker_log.CreateFileBuffer(event.EventID)
	if err != nil {
		event.Status = api.EventStatusFail
		event.ErrorMessage = err.Error()
		log.ErrorWithFields("Operation failed", log.Fields{"event": event})
		return
	}
	output := worker_log.Output
	ch := make(chan interface{})
	defer func() {
		// Push log file to cyclone.
		if err := helper.PushLogToCyclone(event); err != nil {
			event.Status = api.EventStatusFail
			event.ErrorMessage = err.Error()
			log.ErrorWithFields("Operation failed", log.Fields{"event": event})
		}
		output.Close()
		worker_log.SetWatchLogFileSwitch(output.Name(), false)
		// wait util the send the log to kafka throuth cyclone server totally.
		for i := 0; i < WAIT_TIMES; i++ {
			if isChanClosed(ch) {
				break
			}
			time.Sleep(time.Second * 2)
		}
	}()

	topicLog := websocket.CreateTopicName(string(event.Operation), event.Service.UserID,
		event.Service.ServiceID, event.Version.VersionID)
	go worker_log.WatchLogFile(output.Name(), topicLog, ch)

	destDir := vcsManager.GetCloneDir(&event.Service, &event.Version)
	event.Data["context-dir"] = destDir
	event.Data["image-name"] = fmt.Sprintf("%s/%s/%s", dockerManager.Registry,
		strings.ToLower(event.Service.Username), strings.ToLower(event.Service.Name))
	event.Data["tag-name"] = event.Version.Name

	if err = vcsManager.CloneVersionRepository(event); err != nil {
		event.Status = api.EventStatusFail
		event.ErrorMessage = err.Error()
		log.ErrorWithFields("Operation failed", log.Fields{"event": event})
		return
	}

	replaceDockerfile(event, destDir)
	replaceCaicloudYaml(event, destDir)

	// Get the execution tree from the caicloud.yml.
	tree, err := ciManager.Parse(event)
	if err != nil {
		event.Status = api.EventStatusFail
		event.ErrorMessage = err.Error()
		log.ErrorWithFields("Operation failed", log.Fields{"event": event})
		return

	}
	yamlBuild(event, tree, dockerManager, ciManager)
}

// replaceDockerfile create new dockerfile to replace the dockerfile in repo
func replaceDockerfile(event *api.Event, destDir string) {

	// Create dockerfile and caicloud.yml if need
	if event.Service.Dockerfile != "" {
		path := destDir + "/Dockerfile"
		err := osutil.ReplaceFile(path, strings.NewReader(event.Service.Dockerfile))
		if err != nil {
			worker_log.InsertStepLog(event, worker_log.CloneRepository, worker_log.Stop, err)
			return
		}
	}
}

// replaceCaicloudYaml create new yaml to replace the yaml in repo
// default yaml file is caicloud.yml
// if you set your own yaml config name, cyclone will use it.
// if there is no yaml found, cyclone will create a default yaml file with only build step.
func replaceCaicloudYaml(event *api.Event, destDir string) {

	yamlFile := destDir + "/" + ci.DefaultYamlFile
	if event.Service.YAMLConfigName != "" {
		yamlFile = destDir + "/" + event.Service.YAMLConfigName
	}

	if event.Service.CaicloudYaml != "" || !osutil.IsFileExists(yamlFile) {
		content := DefaultCaicloudYaml
		if event.Service.CaicloudYaml != "" {
			content = event.Service.CaicloudYaml
		}
		err := osutil.ReplaceFile(yamlFile, strings.NewReader(content))
		if err != nil {
			worker_log.InsertStepLog(event, worker_log.CloneRepository, worker_log.Stop, err)
			return
		}
	}
}

// yamlBuild func uses for build with yaml file.
func yamlBuild(event *api.Event, tree *parser.Tree, dockerManager *docker.Manager, ciManager *ci.Manager) {
	operation := string(event.Version.Operation)
	// Load the tree to a runner.Build.
	r, err := ciManager.LoadTree(event, tree)
	if err != nil {
		event.Status = api.EventStatusFail
		event.ErrorMessage = err.Error()
		log.ErrorWithFields("Operation failed", log.Fields{"event": event})
		return
	}

	// Build image
	if err = helper.ExecBuild(ciManager, r); err != nil {
		event.Status = api.EventStatusFail
		event.ErrorMessage = err.Error()
		log.ErrorWithFields("Operation failed", log.Fields{"event": event})
		return
	}

	// If need integration
	if strings.Contains(operation, "integration") {
		// Integration
		if err = helper.ExecIntegration(ciManager, r); err != nil {
			event.Status = api.EventStatusFail
			event.ErrorMessage = err.Error()
			log.ErrorWithFields("Operation failed", log.Fields{"event": event})
			return
		}
	}

	// If need publish
	if strings.Contains(operation, "publish") {
		// Publish
		if err = helper.ExecPublish(ciManager, r); err != nil {
			event.Status = api.EventStatusFail
			event.ErrorMessage = err.Error()
			log.ErrorWithFields("Operation failed", log.Fields{"event": event})
			return
		}
	}

	// If need deploy
	if strings.Contains(operation, "deploy") {
		// Deploy
		if err = helper.ExecDeploy(event, dockerManager, r, tree); err != nil {
			event.Status = api.EventStatusFail
			event.ErrorMessage = err.Error()
			log.ErrorWithFields("Operation failed", log.Fields{"event": event})
			return
		}

		err = sendEvent(*event)
		if err != nil {
			log.Errorf("set event result err: %v", err)
		}

		// Deploy Check
		helper.ExecDeployCheck(event, tree)

	}
	event.Status = api.EventStatusSuccess
}

// sendEvent used for setting event for circe server
func sendEvent(event api.Event) error {
	eventID := osutil.GetStringEnv(WORKER_EVENTID, "")
	serverHost := osutil.GetStringEnv(SERVER_HOST, "http://127.0.0.1:7099")

	BaseURL := fmt.Sprintf("%s/api/%s", serverHost, api.APIVersion)
	httpHandler := handler.NewHTTPHandler(BaseURL)
	result := &api.SetEvent{
		Event: event,
	}
	var setResponse api.SetEventResponse
	var err error
	DueTime := time.Now().Add(time.Duration(WORKER_TIMEOUT))
	for DueTime.After(time.Now()) == true {
		err = httpHandler.SetEvent(eventID, result, &setResponse)
		if err != nil {
			if strings.Contains(err.Error(), "connection refused") {
				time.Sleep(time.Minute)
				continue
			}
			return err
		} else {
			log.Infof("set event result: %v", setResponse)
			return nil
		}
	}
	return err
}

func isChanClosed(ch chan interface{}) bool {
	if len(ch) == 0 {
		select {
		case _, ok := <-ch:
			return !ok
		}
	}
	return false
}
