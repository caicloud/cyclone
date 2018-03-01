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

package worker

import (
	"fmt"
	"strings"
	"time"

	"github.com/zoumo/logdog"

	"github.com/caicloud/cyclone/cloud"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/docker"
	"github.com/caicloud/cyclone/pkg/worker/cycloneserver"
	_ "github.com/caicloud/cyclone/pkg/worker/scm/provider"
	"github.com/caicloud/cyclone/pkg/worker/stage"
)

const (
	WorkerTimeout = 7200 * time.Second
)

// Worker ...
type Worker struct {
	Client cycloneserver.CycloneServerClient
	Config *Config
	Envs   *cloud.WorkerEnvs
}

// Run starts the worker
func (worker *Worker) Run() error {
	logdog.Info("worker start with", logdog.Fields{"config": worker.Config, "envs": worker.Envs})

	// Get event for cyclone server
	eventID := worker.Config.ID
	event, err := worker.Client.GetEvent(eventID)
	if err != nil {
		logdog.Errorf("fail to get event %s as %s", eventID, err.Error())

		// TODO update event status
		//		event.PipelineRecord.Status = api.Failed
		//		sendErr := worker.sendEvent(event)
		//		if sendErr != nil {
		//			logdog.Errorf("set event result err: %v", err)
		//		}
		return err
	}
	logdog.Info("get event success", logdog.Fields{"event": event})

	// Handle event
	logdog.Info("handleEvent ...")
	worker.HandleEvent(event)

	// Sent event for cyclone server
	err = worker.sendEvent(event)
	if err != nil {
		logdog.Errorf("set event result err: %v", err)
		return err
	}
	logdog.Info("send event to server", logdog.Fields{"event": event})
	return nil
}

// Steps:
// 1. Git clone code
// 2. New Docker manager
// 3. Run pipeline stages
// 4. Create the tag in SCM
func (worker *Worker) HandleEvent(event *api.Event) {
	project := event.Project
	pipeline := event.Pipeline
	performParams := event.PipelineRecord.PerformParams
	performStages := performParams.Stages
	stageSet := convertPerformStageSet(performStages)
	build := pipeline.Build

	registryLocation := worker.Envs.RegistryLocation
	registryUsername := worker.Envs.RegistryUsername
	registryPassword := worker.Envs.RegistryPassword

	dockerManager, err := docker.NewDockerManager(worker.Config.DockerHost, registryLocation, registryUsername, registryPassword)
	if err != nil {
		logdog.Error(err.Error())
		return
	}

	// TODO(robin) Seperate unit test and package stage.
	stageManager := stage.NewStageManager(dockerManager, worker.Client)
	stageManager.SetRecordInfo(project.Name, pipeline.Name, event.ID)

	// Init StageStatus
	if event.PipelineRecord.StageStatus == nil {
		event.PipelineRecord.StageStatus = &api.StageStatus{}
	}

	// Execute the code checkout stage,
	err = execCodeCheckout(worker, stageManager, event)
	if err != nil {
		logdog.Error(err.Error())
		return
	}

	// Execute the package stage, this stage is required and can not be skipped.
	err = stageManager.ExecPackage(build.BuilderImage, build.Stages.UnitTest, build.Stages.Package)
	if err != nil {
		logdog.Error(err.Error())
		return
	}

	// The built images from image build stage.
	var builtImages []string

	// Execute the image build stage if necessary.
	if _, ok := stageSet[api.ImageBuildStageName]; ok {
		builtImages, err = stageManager.ExecImageBuild(build.Stages.ImageBuild)
		if err != nil {
			logdog.Error(err.Error())
			return
		}
	}

	// Execute the integration test stage if necessary.
	if _, ok := stageSet[api.IntegrationTestStageName]; ok {
		err = stageManager.ExecIntegrationTest(builtImages, build.Stages.IntegrationTest)
		if err != nil {
			logdog.Error(err.Error())
			return
		}
	}

	// Execute the integration test stage if necessary.
	if _, ok := stageSet[api.ImageReleaseStageName]; ok {
		err = stageManager.ExecImageRelease(builtImages, build.Stages.ImageRelease)
		if err != nil {
			logdog.Error(err.Error())
			return
		}
	}

	logdog.Info("success: ")

	// update event.PipelineRecord.Status from running to success
	event.PipelineRecord.Status = api.Success
}

func convertPerformStageSet(stages []api.PipelineStageName) map[api.PipelineStageName]struct{} {
	stageSet := make(map[api.PipelineStageName]struct{})
	for _, stage := range stages {
		stageSet[stage] = struct{}{}
	}

	return stageSet
}

// sendEvent used for setting event for circe server.
func (worker *Worker) sendEvent(event *api.Event) error {
	DueTime := time.Now().Add(time.Duration(WorkerTimeout))
	for DueTime.After(time.Now()) == true {
		response, err := worker.Client.SetEvent(event)
		if err != nil {
			logdog.Warnf("set event failed: %v", err)
			if strings.Contains(err.Error(), "connection refused") {
				time.Sleep(time.Minute)
				continue
			}
			return err
		}
		logdog.Infof("set event result: %v", response)
		return nil
	}
	return nil
}

// formatVersionName replace the record name with default name '$commitID[:7]-$createTime' when name empty in create version.
func formatVersionName(id string, event *api.Event) {
	if event.PipelineRecord.Name == "" && id != "" {
		// report to server in sendEvent.
		version := fmt.Sprintf("%s-%s", id[:7], event.PipelineRecord.StartTime.Format("060102150405"))
		event.PipelineRecord.Name = version
		event.PipelineRecord.StageStatus.CodeCheckout.Version = version
	}
}

// execCodeCheckout Checkout code and report real-time status to cycloue server.
func execCodeCheckout(worker *Worker, stageManager stage.StageManager, event *api.Event) error {
	project := event.Project
	build := event.Pipeline.Build
	event.PipelineRecord.StageStatus.CodeCheckout = &api.CodeCheckoutStageStatus{
		GeneralStageStatus: api.GeneralStageStatus{
			Status:    api.Running,
			StartTime: time.Now(),
		},
	}
	go worker.sendEvent(event)

	commitID, err := stageManager.ExecCodeCheckout(project.SCM.Token, build.Stages.CodeCheckout)
	if err != nil {
		logdog.Error(err.Error())
		event.PipelineRecord.StageStatus.CodeCheckout.Status = api.Failed
		event.PipelineRecord.StageStatus.CodeCheckout.EndTime = time.Now()
		go worker.sendEvent(event)
		return err
	}

	event.PipelineRecord.StageStatus.CodeCheckout.Status = api.Success
	event.PipelineRecord.StageStatus.CodeCheckout.EndTime = time.Now()
	// Format version name when code checkout success.
	formatVersionName(commitID, event)
	go worker.sendEvent(event)
	return nil
}
