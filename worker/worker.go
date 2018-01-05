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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/cloud"
	"github.com/caicloud/cyclone/docker"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/store"
	"github.com/caicloud/cyclone/websocket"
	"github.com/caicloud/cyclone/worker/ci"
	"github.com/caicloud/cyclone/worker/ci/parser"
	"github.com/caicloud/cyclone/worker/handler"
	"github.com/caicloud/cyclone/worker/helper"
	worker_log "github.com/caicloud/cyclone/worker/log"
	"github.com/caicloud/cyclone/worker/vcs"
	"github.com/zoumo/logdog"
)

const (
	WorkerTimeout = 7200 * time.Second
	WaitTimes     = 5

	DefaultCaicloudYaml = "build:\n  dockerfile_name: Dockerfile"
)

const (
	ImageBuild = 1 << iota // 0001
	Integrate              // 0010
	Publish                // 0100
	Deploy                 // 1000
)

// Worker ...
type Worker struct {
	Config *Config
	Envs   *cloud.WorkerEnvs
}

// Run starts the worker
func (worker *Worker) Run() error {

	logdog.Info("worker start with", logdog.Fields{"config": worker.Config, "envs": worker.Envs})

	// Get job for cyclone server
	job, err := worker.getJob()
	if err != nil {
		job.PipelineRecord.Status = store.JobStatusFailed
		sendErr := worker.sendJob(job)
		if sendErr != nil {
			logdog.Errorf("set job result err: %v", err)
		}
		return err
	}
	logdog.Info("get job success", logdog.Fields{"job": job})

	// Handle job
	logdog.Info("handleJob ...")
	worker.handleJob(job)

	// Sent job for cyclone server
	err = worker.sendJob(job)
	if err != nil {
		logdog.Errorf("set job result err: %v", err)
		return err
	}
	logdog.Info("send job to server", logdog.Fields{"job": job})
	return nil
}

// getEvent used for getting event for cyclone server
func (worker *Worker) getJob() (*store.Job, error) {
	if worker.Envs.Job == nil {
		return nil, fmt.Errorf("No job founded in worker")
	}
	return worker.Envs.Job, nil
}

// handleEvent analysize the the operation in event, and do the relate operation
func (worker *Worker) handleJob(job *store.Job) {
	vcsManager := vcs.NewManager()

	logServer := worker.Envs.LogServer

	if err := worker_log.DialLogServer(logServer); err != nil {
		return
	}

	go worker_log.SendHeartBeat()
	defer worker_log.Disconnect()

	worker.createPipelineRecord(vcsManager, job)
}

// createVersion create version for service and push log to server via websocket
// Step1: clone code from VCS
// Step2: integretion
// Step3: publish
// Step4: deploy
func (worker *Worker) createPipelineRecord(vcsManager *vcs.Manager, job *store.Job) {
	registryLocation := worker.Envs.RegistryLocation
	registryUsername := worker.Envs.RegistryUsername
	registryPassword := worker.Envs.RegistryPassword

	dockerManager, err := docker.NewManager(worker.Config.DockerHost, "",
		api.RegistryCompose{registryLocation, registryUsername, registryPassword})
	if err != nil {
		return
	}
	ciManager, err := ci.NewManager(dockerManager)
	if err != nil {
		return
	}

	defer func() {
		// Push log file to cyclone.

		// wait util the send the log to kafka throuth cyclone server totally.

	}()

	destDir := vcsManager.GetCloneDir()
	job.PipelineRecord.Data["context-dir"] = destDir

	if err = vcsManager.CloneVersionRepository(job); err != nil {
		return
	}

	formatVersionName(job)

	setImageNameAndTag(dockerManager, job)

	replaceDockerfile(job, destDir)

	// Get the execution tree from the caicloud.yml.
	tree, err := ciManager.Parse(job)
	if err != nil {
		return
	}
	worker.yamlBuild(event, tree, dockerManager, ciManager)
	worker.build(job, dockerManager)
}

// replaceDockerfile create new dockerfile to replace the dockerfile in repo
func replaceDockerfile(job *store.Job, destDir string) {

	// Create dockerfile and caicloud.yml if need
	if job.Pipeline.Dockerfile != "" {
		path := destDir + "/Dockerfile"
		err := osutil.ReplaceFile(path, strings.NewReader(job.Pipeline.Dockerfile))
		if err != nil {
			return
		}
	}
}

func (worker *Worker) prebuild(job *store.Job) {
	if job.Pipeline.Build
}

func (worker *Worker) imageBuild(job *store.Job) {

}

func (worker *Worker) integrate(job *store.Job) {
	
}

func (worker *Worker) publish(job *store.Job) {
	
}

func (worker *Worker) deploy(job *store.Job) {
	
}

// build func uses for build with job.
func (worker *Worker) build(job *store.Job, dockerManager *docker.Manager) {
	operation := string(job.PipelineRecord.Operation)

	err := worker.prebuild(job, dockerManager)
	if err != nil {
		return
	}

	if operation&ImageBuild != 0 {
		err = worker.imageBuild(job, dockerManager)
		if err != nil {
			return
		}
	}

	if operation&Integrate != 0 {
		err = worker.integrate(job, dockerManager)
		if err != nil {
			return
		}
	}

	if operation&Publish != 0 {
		err = worker.publish(job, dockerManager)
		if err != nil {
			return
		}
	}

	if operation&Deploy != 0 {
		err = worker.deploy(job, dockerManager)
		if err != nil {
			return
		}
	}

	job.Status = api.EventStatusSuccess
	job.PipelineRecord = api.EventStatusSuccess
}

// yamlBuild func uses for build with yaml file.
func (worker *Worker) yamlBuild(event *api.Event, tree *parser.Tree, dockerManager *docker.Manager, ciManager *ci.Manager) {
	operation := string(event.Version.Operation)
	// Load the tree to a runner.Build.
	r, err := ciManager.LoadTree(event, tree)
	if err != nil {
		setEventFailStatus(event, err.Error())
		return
	}

	// Always run prebuild.
	if err = ciManager.ExecPreBuild(r); err != nil {
		setEventFailStatus(event, err.Error())
		return
	}

	// Run build if needed.
	if strings.Contains(operation, "imageBuild") {
		if err = ciManager.ExecBuild(r); err != nil {
			setEventFailStatus(event, err.Error())
			return
		}
	}

	// If need integration
	if strings.Contains(operation, "integration") {
		// Integration
		if err = helper.ExecIntegration(ciManager, r); err != nil {
			setEventFailStatus(event, err.Error())
			return
		}
	}

	// If need publish
	if strings.Contains(operation, "publish") {
		// Publish
		if err = helper.ExecPublish(ciManager, r); err != nil {
			setEventFailStatus(event, err.Error())
			return
		}
	}

	// If need deploy
	if strings.Contains(operation, "deploy") {
		// Deploy
		if err = helper.ExecDeploy(event, dockerManager, r, tree); err != nil {
			setEventFailStatus(event, err.Error())
			return
		}

		err = worker.sendEvent(*event)
		if err != nil {
			logdog.Errorf("set event result err: %v", err)
		}

		// Deploy Check
		helper.ExecDeployCheck(event, tree)

	}
	event.Status = api.EventStatusSuccess
}

// sendJob used for setting job for circe server
func (worker *Worker) sendJob(job store.Job) error {
	serverHost := worker.Envs.CycloneServer

	BaseURL := fmt.Sprintf("%s/api/%s", serverHost, api.APIVersion)
	httpHandler := handler.NewHTTPHandler(BaseURL)
	result := &api.SetEvent{
		Event: event,
	}
	var setResponse api.SetEventResponse
	var err error
	DueTime := time.Now().Add(time.Duration(WorkerTimeout))
	for DueTime.After(time.Now()) == true {
		err = httpHandler.SetEvent(eventID, result, &setResponse)
		if err != nil {
			if strings.Contains(err.Error(), "connection refused") {
				time.Sleep(time.Minute)
				continue
			}
			return err
		}
		logdog.Infof("set job result: %v", setResponse)
		return nil
	}
	return err
}

// formatPipelineRecordName replace the random name with default name '$commitID[:7]-$createTime' when name empty in create pipelineRecord
func formatPipelineRecordName(job *store.Job) {
	if job.PipelineRecord.Name == job.PipelineRecord.ID && job.PipelineRecord.Commit != "" {
		// report to server in sendJob
		job.PipelineRecord.Name = fmt.Sprintf("%s-%s", job.PipelineRecord.Commit[:7], job.PipelineRecord.CreateTime.Format("060102150405"))
	}
}

// setImageNameAndTag sets the image name and tag name of the event.
func setImageNameAndTag(dockerManager *docker.Manager, job *store.Job) {
	var imageName, tagName string
	imageName = job.Pipeline.ImageName
	names := strings.Split(strings.TrimSpace(imageName), ":")
	switch len(names) {
	case 1:
		imageName = names[0]
	case 2:
		imageName = names[0]
		tagName = names[1]
	default:
		logdog.Error("image name error", logdog.Fields{"imageName": imageName})
		imageName = ""
	}

	if imageName == "" {
		imageName = fmt.Sprintf("%s/%s/%s", dockerManager.Registry,
			strings.ToLower(job.Pipeline.Owner), strings.ToLower(job.Pipeline.Name))
	}

	if tagName == "" {
		pipelineRecord := job.PipelineRecord
		if pipelineRecord.Name == "" {
			tagName = fmt.Sprintf("%s", time.Now().Format("060102150405"))
			if pipelineRecord.Commit != "" {
				tagName = fmt.Sprintf("%s-%s", pipelineRecord.Commit[:7], tagName)
			}
		} else {
			tagName = pipelineRecord.Name
		}
	}

	job.PipelineRecord.Data["image-name"] = imageName
	job.PipelineRecord.Data["tag-name"] = tagName
}
