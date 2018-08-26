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
	log "github.com/golang/glog"

	"github.com/caicloud/cyclone/cmd/worker/options"
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
	Client  cycloneserver.CycloneServerClient
	Options *options.WorkerOptions
}

// NewWorker returns a new APIServer with config
func NewWorker(opts *options.WorkerOptions) *Worker {
	client := cycloneserver.NewClient(opts.CycloneServer)

	return &Worker{
		Client:  client,
		Options: opts,
	}
}

// Run starts the worker
func (worker *Worker) Run() error {
	log.Info("worker start with options: %v", worker.Options)

	// Get event for cyclone server
	eventID := worker.Options.EventID
	event, err := worker.Client.GetEvent(eventID)
	if err != nil {
		log.Errorf("fail to get event %s as %s", eventID, err.Error())
		return err
	}

	// Handle event
	log.Infof("start to handle event %s", eventID)
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

	opts := worker.Options
	registryLocation := opts.RegistryLocation
	registryUsername := opts.RegistryUsername
	registryPassword := opts.RegistryPassword

	dockerManager, err := docker.NewDockerManager(opts.DockerHost, registryLocation, registryUsername, registryPassword)
	if err != nil {
		log.Error(err.Error())
		return
	}

	// Init StageStatus
	if event.PipelineRecord.StageStatus == nil {
		event.PipelineRecord.StageStatus = &api.StageStatus{}
	}

	// TODO(robin) Seperate unit test and package stage.
	stageManager := stage.NewStageManager(dockerManager, worker.Client, project.Registry, performParams)
	stageManager.SetRecordInfo(project.Name, pipeline.Name, event.ID)
	stageManager.SetEvent(event)

	// Execute the code checkout stage.
	err = stageManager.ExecCodeCheckout(project.SCM.Token, build.Stages.CodeCheckout)
	if err != nil {
		log.Error(err.Error())
		return
	}

	// Execute the package stage, this stage is required and can not be skipped.
	err = stageManager.ExecPackage(build.BuilderImage, build.BuildInfo, build.Stages.UnitTest, build.Stages.Package)
	if err != nil {
		log.Error(err.Error())
		return
	}

	// The built images from image build stage.
	var builtImages []string

	// Execute the image build stage if necessary.
	if _, ok := stageSet[api.ImageBuildStageName]; ok {
		builtImages, err = stageManager.ExecImageBuild(build.Stages.ImageBuild)
		if err != nil {
			log.Error(err.Error())
			return
		}
	}

	// Execute the integration test stage if necessary.
	if _, ok := stageSet[api.IntegrationTestStageName]; ok {
		err = stageManager.ExecIntegrationTest(builtImages, build.Stages.IntegrationTest)
		if err != nil {
			log.Error(err.Error())
			return
		}
	}

	// Execute the image release stage if necessary.
	if _, ok := stageSet[api.ImageReleaseStageName]; ok {
		err = stageManager.ExecImageRelease(builtImages, build.Stages.ImageRelease)
		if err != nil {
			log.Error(err.Error())
			return
		}
	}

	if event.PipelineRecord.PerformParams.CreateSCMTag {
		codesource := event.Pipeline.Build.Stages.CodeCheckout.MainRepo
		err = scm.NewTagFromLatest(codesource, project.SCM, event.PipelineRecord.Name, event.PipelineRecord.PerformParams.Description)
		if err != nil {
			log.Errorf("new tag from latest fail : %v", err)
			err = stage.UpdateEvent(worker.Client, event, api.CreateScmTagStageName, api.Failed, err)
			if err != nil {
				log.Errorf("fail to update event for stage %s as %v", api.CreateScmTagStageName, err)
				return
			}
			return
		}
	}

	log.Info("success: ")

	// update event.PipelineRecord.Status from running to success
	event.PipelineRecord.Status = api.Success

	// Sent event for cyclone server
	err = worker.Client.SendEvent(event)
	if err != nil {
		log.Errorf("set event result err: %v", err)
		return
	}
	log.Infof("send event %s to server", event)
}

func convertPerformStageSet(stages []api.PipelineStageName) map[api.PipelineStageName]struct{} {
	stageSet := make(map[api.PipelineStageName]struct{})
	for _, stage := range stages {
		stageSet[stage] = struct{}{}
	}

	return stageSet
}
