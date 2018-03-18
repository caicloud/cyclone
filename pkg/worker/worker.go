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

	"github.com/zoumo/logdog"

	"github.com/caicloud/cyclone/cloud"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/docker"
	"github.com/caicloud/cyclone/pkg/scm"
	_ "github.com/caicloud/cyclone/pkg/scm/provider"
	"github.com/caicloud/cyclone/pkg/worker/cycloneserver"
	_ "github.com/caicloud/cyclone/pkg/worker/scm/provider"

	"github.com/caicloud/cyclone/pkg/worker/stage"
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
		return err
	}
	logdog.Info("get event success", logdog.Fields{"event": event})

	// Handle event
	logdog.Info("handleEvent ...")
	worker.HandleEvent(event)

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

	// Init StageStatus
	if event.PipelineRecord.StageStatus == nil {
		event.PipelineRecord.StageStatus = &api.StageStatus{}
	}

	// TODO(robin) Seperate unit test and package stage.
	stageManager := stage.NewStageManager(dockerManager, worker.Client, performParams)
	stageManager.SetRecordInfo(project.Name, pipeline.Name, event.ID)
	stageManager.SetEvent(event)

	// Execute the code checkout stage,
	err = stageManager.ExecCodeCheckout(project.SCM.Token, build.Stages.CodeCheckout)
	if err != nil {
		logdog.Error(err.Error())
		return
	}

	// Execute the package stage, this stage is required and can not be skipped.
	err = stageManager.ExecPackage(build.BuilderImage, build.BuildInfo, build.Stages.UnitTest, build.Stages.Package)
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

	// Execute the image release stage if necessary.
	if _, ok := stageSet[api.ImageReleaseStageName]; ok {
		err = stageManager.ExecImageRelease(builtImages, build.Stages.ImageRelease)
		if err != nil {
			logdog.Error(err.Error())
			return
		}
	}

	if event.PipelineRecord.PerformParams.CreateSCMTag {
		codesource := event.Pipeline.Build.Stages.CodeCheckout.CodeSources[0]
		err = scm.NewTagFromLatest(codesource, event.PipelineRecord.Name, event.PipelineRecord.PerformParams.Description, project.SCM.Token)
		if err != nil {
			logdog.Errorf("new tag from latest fail : %v", err)
			event.PipelineRecord.Status = api.Failed
			event.PipelineRecord.ErrorMessage = fmt.Sprintf("generate tag fail: %v", err)
			err = worker.Client.SendEvent(event)
			if err != nil {
				logdog.Errorf("set event result err: %v", err)
				return
			}
			return
		}
	}

	logdog.Info("success: ")

	// update event.PipelineRecord.Status from running to success
	event.PipelineRecord.Status = api.Success

	// Sent event for cyclone server
	err = worker.Client.SendEvent(event)
	if err != nil {
		logdog.Errorf("set event result err: %v", err)
		return
	}
	logdog.Info("send event to server", logdog.Fields{"event id": event.ID})
}

func convertPerformStageSet(stages []api.PipelineStageName) map[api.PipelineStageName]struct{} {
	stageSet := make(map[api.PipelineStageName]struct{})
	for _, stage := range stages {
		stageSet[stage] = struct{}{}
	}

	return stageSet
}
