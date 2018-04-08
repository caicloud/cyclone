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

package stage

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	docker_client "github.com/fsouza/go-dockerclient"
	log "github.com/golang/glog"
	"github.com/zoumo/logdog"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/docker"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/pkg/pathutil"
	"github.com/caicloud/cyclone/pkg/worker/cycloneserver"
	"github.com/caicloud/cyclone/pkg/worker/scm"
)

// logFileNameTemplate ...
const (
	logFileNameTemplate = "/tmp/logs/%s.log"

	// waitTime represents the wait time for log pushing.
	waitTime = 3 * time.Second
)

var event *api.Event

type StageManager interface {
	SetRecordInfo(project, pipeline, recordID string)
	SetEvent(event *api.Event)
	ExecCodeCheckout(token string, stage *api.CodeCheckoutStage) error
	ExecPackage(*api.BuilderImage, *api.BuildInfo, *api.UnitTestStage, *api.PackageStage) error
	ExecImageBuild(stage *api.ImageBuildStage) ([]string, error)
	ExecIntegrationTest(builtImages []string, stage *api.IntegrationTestStage) error
	ExecImageRelease(builtImages []string, stage *api.ImageReleaseStage) error
}

type recordInfo struct {
	project  string
	pipeline string
	recordID string
}

type stageManager struct {
	recordInfo
	dockerManager *docker.DockerManager
	cycloneClient cycloneserver.CycloneServerClient
	performParams *api.PipelinePerformParams
	registry      *api.Registry
}

func NewStageManager(dockerManager *docker.DockerManager, cycloneClient cycloneserver.CycloneServerClient,
	registry *api.Registry, performParams *api.PipelinePerformParams) StageManager {
	err := pathutil.EnsureParentDir(logFileNameTemplate, os.ModePerm)
	if err != nil {
		log.Errorf(err.Error())
	}

	return &stageManager{
		dockerManager: dockerManager,
		cycloneClient: cycloneClient,
		performParams: performParams,
		registry:      registry,
	}
}

func (sm *stageManager) SetRecordInfo(project, pipeline, recordID string) {
	sm.project = project
	sm.pipeline = pipeline
	sm.recordID = recordID
}

func (sm *stageManager) SetEvent(e *api.Event) {
	event = e
}

// ExecCodeCheckout Checkout code and report real-time status to cycloue server.
func (sm *stageManager) ExecCodeCheckout(token string, stage *api.CodeCheckoutStage) (err error) {
	event.PipelineRecord.StageStatus.CodeCheckout = &api.CodeCheckoutStageStatus{
		GeneralStageStatus: api.GeneralStageStatus{
			Status:    api.Running,
			StartTime: time.Now(),
		},
	}
	sm.cycloneClient.SendEvent(event)

	closeLog := make(chan struct{})
	defer func() {
		if err != nil {
			event.PipelineRecord.StageStatus.CodeCheckout.Status = api.Failed
			event.PipelineRecord.Status = api.Failed
			event.PipelineRecord.ErrorMessage = fmt.Sprintf("code checkout fail : %v", err)
		} else {
			event.PipelineRecord.StageStatus.CodeCheckout.Status = api.Success
		}

		event.PipelineRecord.StageStatus.CodeCheckout.EndTime = time.Now()
		sm.cycloneClient.SendEvent(event)

		time.Sleep(waitTime)
		close(closeLog)
	}()

	logs, err := scm.CloneRepos(token, stage, sm.performParams.Ref)
	if err != nil {
		logdog.Error(err.Error())
		return err
	}

	fileName := fmt.Sprintf(logFileNameTemplate, api.CodeCheckoutStageName)
	logFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer func() {
		logFile.WriteString(generateStageFinishLog(api.CodeCheckoutStageName, err))
		logFile.Close()
	}()
	logFile.WriteString(generateStageStartLog(api.CodeCheckoutStageName))

	// Just one line of log, will add more detailed logs.
	logFile.WriteString(logs)

	go sm.cycloneClient.PushLogStream(sm.project, sm.pipeline, sm.recordID, api.CodeCheckoutStageName, "", fileName, closeLog)

	setCommits(stage)
	return nil
}

func (sm *stageManager) ExecPackage(builderImage *api.BuilderImage, buildInfo *api.BuildInfo, unitTestStage *api.UnitTestStage, packageStage *api.PackageStage) (err error) {
	event.PipelineRecord.StageStatus.Package = &api.GeneralStageStatus{
		Status:    api.Running,
		StartTime: time.Now(),
	}
	sm.cycloneClient.SendEvent(event)

	closeLog := make(chan struct{})
	defer func() {
		if err != nil {
			event.PipelineRecord.StageStatus.Package.Status = api.Failed
			event.PipelineRecord.Status = api.Failed
			event.PipelineRecord.ErrorMessage = fmt.Sprintf("package fail : %v", err)
		} else {
			event.PipelineRecord.StageStatus.Package.Status = api.Success
		}
		event.PipelineRecord.StageStatus.Package.EndTime = time.Now()
		sm.cycloneClient.SendEvent(event)

		time.Sleep(waitTime)
		close(closeLog)
	}()

	// Trick: bind the docker sock file to container to support
	// docker operation in the container.
	enterpoint := []byte(sm.dockerManager.EndPoint)[7:]
	log.Infof("enterpoint is %s", string(enterpoint))
	pathenterpoint := fmt.Sprintf("%s:%s", string(enterpoint), "/var/run/docker.sock")

	cloneDir := scm.GetCloneDir()
	hostConfig := &docker_client.HostConfig{
		Binds: []string{fmt.Sprintf("%s:%s", cloneDir, cloneDir), pathenterpoint},
	}

	// Mount the cache volume.
	if sm.performParams.CacheDependency && buildInfo != nil && buildInfo.CacheDependency && buildInfo.BuildTool != nil {
		var bindVolume string
		switch buildInfo.BuildTool.Name {
		case api.MavenBuildTool:
			bindVolume = "/root/.m2:/root/.m2"
		case api.NPMBuildTool:
			bindVolume = "/root/.npm:/root/.npm"
		default:
			return fmt.Errorf("Not support build tool %s, only supports: %s, %s", buildInfo.BuildTool.Name, api.MavenBuildTool, api.NPMBuildTool)
		}

		hostConfig.Binds = append(hostConfig.Binds, bindVolume)
	}

	// Start and run the container from builder image.
	config := &docker_client.Config{
		Image:      builderImage.Image,
		Env:        convertEnvs(builderImage.EnvVars),
		OpenStdin:  true, // Open stdin to keep the container running after starts.
		WorkingDir: cloneDir,
		// Entrypoint: []string{"/bin/sh", "-e", "-c"},
	}

	cco := docker_client.CreateContainerOptions{
		Config:     config,
		HostConfig: hostConfig,
	}
	cid, err := sm.dockerManager.StartContainer(cco, generateAuthConfig(sm.registry))
	if err != nil {
		return err
	}

	defer func() {
		if err := sm.dockerManager.RemoveContainer(cid); err != nil {
			log.Errorf("Fail to remove the container %s", cid)
		}
	}()

	// Execute unit test and package commands in the builder container.
	// Run stage script in container
	cmds := packageStage.Command
	// Run the unit test commands before package commands if there is unit test stage.
	if unitTestStage != nil {
		cmds = append(unitTestStage.Command, cmds...)
	}

	fileName := fmt.Sprintf(logFileNameTemplate, api.PackageStageName)
	logFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer func() {
		logFile.WriteString(generateStageFinishLog(api.PackageStageName, err))
		logFile.Close()
	}()
	logFile.WriteString(generateStageStartLog(api.PackageStageName))

	go sm.cycloneClient.PushLogStream(sm.project, sm.pipeline, sm.recordID, api.PackageStageName, "", fileName, closeLog)

	eo := docker.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Container:    cid,
		OutputStream: logFile,
		ErrorStream:  logFile,
	}

	// Run the commands one by one.
	for _, cmd := range cmds {
		eo.Cmd = strings.Split(cmd, " ")
		err = sm.dockerManager.ExecInContainer(eo)
		if err != nil {
			return err
		}
	}

	// Copy the build outputs if necessary.
	// Only need to copy the outputs not in the current workspace. The outputs must be absolute path of files or folders.
	cloneDir = cloneDir + "/"
	for _, ot := range packageStage.Outputs {
		if strings.HasPrefix(ot, "./") {
			ot = strings.TrimPrefix(ot, "./")
			ot = cloneDir + ot
		} else if !strings.HasPrefix(ot, "/") {
			ot = cloneDir + ot
		}

		opt := docker.CopyFromContainerOptions{
			Container:     cid,
			HostPath:      cloneDir,
			ContainerPath: ot,
		}
		if err = sm.dockerManager.CopyFromContainer(opt); err != nil {
			return err
		}
	}

	return nil
}

func (sm *stageManager) ExecImageBuild(stage *api.ImageBuildStage) ([]string, error) {
	var err error
	event.PipelineRecord.StageStatus.ImageBuild = &api.GeneralStageStatus{
		Status:    api.Running,
		StartTime: time.Now(),
	}
	sm.cycloneClient.SendEvent(event)

	closeLog := make(chan struct{})
	defer func() {
		if err != nil {
			event.PipelineRecord.StageStatus.ImageBuild.Status = api.Failed
			event.PipelineRecord.Status = api.Failed
			event.PipelineRecord.ErrorMessage = fmt.Sprintf("image build fail : %v", err)
		} else {
			event.PipelineRecord.StageStatus.ImageBuild.Status = api.Success
		}
		event.PipelineRecord.StageStatus.ImageBuild.EndTime = time.Now()
		sm.cycloneClient.SendEvent(event)

		time.Sleep(waitTime)
		close(closeLog)
	}()

	builtImages := []string{}
	wg := sync.WaitGroup{}
	for _, buildInfo := range stage.BuildInfos {
		wg.Add(1)
		go func(buildInfo *api.ImageBuildInfo) {
			defer wg.Done()

			if image, bErr := sm.buildImage(buildInfo, closeLog); bErr != nil {
				log.Errorf("Fail to build image %s as %v", buildInfo.TaskName, bErr)
				err = bErr
			} else {
				builtImages = append(builtImages, image)
			}
		}(buildInfo)
	}

	wg.Wait()
	return builtImages, err
}

func (sm *stageManager) buildImage(buildInfo *api.ImageBuildInfo, close chan struct{}) (image string, err error) {
	fileName := fmt.Sprintf(logFileNameTemplate, api.ImageBuildStageName)
	if buildInfo.TaskName != "" {
		fileName = fmt.Sprintf(logFileNameTemplate, fmt.Sprintf("%s-%s", api.ImageBuildStageName, buildInfo.TaskName))
	}

	logFile, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer logFile.Close()

	defer func() {
		logFile.WriteString(generateStageFinishLog(api.ImageBuildStageName, err))
		logFile.Close()
	}()
	logFile.WriteString(generateStageStartLog(api.ImageBuildStageName))

	go sm.cycloneClient.PushLogStream(sm.project, sm.pipeline, sm.recordID, api.ImageBuildStageName, buildInfo.TaskName, fileName, close)

	authOpts := docker_client.AuthConfigurations{
		Configs: make(map[string]docker_client.AuthConfiguration),
	}
	if sm.registry != nil {
		authOpt := docker_client.AuthConfiguration{
			Username: sm.registry.Username,
			Password: sm.registry.Password,
		}

		authOpts.Configs[sm.registry.Server] = authOpt
	}

	// Image build ptions, need AuthConfigs for auth to pull images in Dockerfiles.
	opt := docker_client.BuildImageOptions{
		AuthConfigs:    authOpts,
		RmTmpContainer: true,
		Memswap:        -1,
		OutputStream:   logFile,
	}

	cloneDir := scm.GetCloneDir()
	opt.ContextDir = cloneDir
	if buildInfo.ContextDir != "" {
		opt.ContextDir = cloneDir + "/" + buildInfo.ContextDir
	}

	opt.Dockerfile = "Dockerfile"
	if buildInfo.Dockerfile != "" {
		if err = osutil.ReplaceFile(opt.ContextDir+"/Dockerfile", strings.NewReader(buildInfo.Dockerfile)); err != nil {
			return "", err
		}
	} else {
		if buildInfo.DockerfilePath != "" {
			opt.Dockerfile = strings.TrimPrefix(strings.TrimPrefix(buildInfo.DockerfilePath, buildInfo.ContextDir), "/")
		}
	}
	opt.Name = formatImageName(buildInfo.ImageName)

	if err = sm.dockerManager.BuildImage(opt); err != nil {
		return "", err
	}

	return opt.Name, nil
}

func (sm *stageManager) ExecIntegrationTest(builtImages []string, stage *api.IntegrationTestStage) (err error) {
	event.PipelineRecord.StageStatus.IntegrationTest = &api.GeneralStageStatus{
		Status:    api.Running,
		StartTime: time.Now(),
	}
	sm.cycloneClient.SendEvent(event)

	closeLog := make(chan struct{})
	defer func() {
		if err != nil {
			event.PipelineRecord.StageStatus.IntegrationTest.Status = api.Failed
			event.PipelineRecord.Status = api.Failed
			event.PipelineRecord.ErrorMessage = fmt.Sprintf("integration test fail : %v", err)
		} else {
			event.PipelineRecord.StageStatus.IntegrationTest.Status = api.Success
		}
		event.PipelineRecord.StageStatus.IntegrationTest.EndTime = time.Now()
		sm.cycloneClient.SendEvent(event)

		time.Sleep(waitTime)
		close(closeLog)
	}()

	log.Infof("Exec integration test stage for pipeline record %s/%s/%s", sm.project, sm.pipeline, sm.recordID)

	// Start the services.
	serviceInfos, err := sm.StartServicesForIntegrationTest(stage.Services)
	if err != nil {
		return err
	}
	defer func() {
		var err error
		for s, cid := range serviceInfos {
			if err = sm.dockerManager.RemoveContainer(cid); err != nil {
				log.Errorf("Fail to remove container %s for the service %s", cid, s)
			}
		}
	}()

	testConfig := stage.Config
	var testImage string
	for _, builtImage := range builtImages {
		if strings.Contains(builtImage, testConfig.ImageName) {
			testImage = builtImage
			break
		}
	}

	if testImage == "" {
		err = fmt.Errorf("image %s in integration test config is not the built images %v", testConfig.ImageName, builtImages)
		log.Error(err.Error())
		return err
	}

	var serviceNames []string
	for name := range serviceInfos {
		serviceNames = append(serviceNames, name)
	}

	// Start the built image.
	config := &docker_client.Config{
		Image:     testImage,
		Env:       convertEnvs(testConfig.EnvVars),
		OpenStdin: true, // Open stdin to keep the container running after starts.
	}

	hostConfig := &docker_client.HostConfig{
		Links: serviceNames,
	}
	cco := docker_client.CreateContainerOptions{
		Config:     config,
		HostConfig: hostConfig,
	}
	cid, err := sm.dockerManager.StartContainer(cco, generateAuthConfig(sm.registry))
	if err != nil {
		return err
	}
	defer func() {
		if err := sm.dockerManager.RemoveContainer(cid); err != nil {
			log.Errorf("Fail to remove the container %s", cid)
		}
	}()

	fileName := fmt.Sprintf(logFileNameTemplate, api.IntegrationTestStageName)
	logFile, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer func() {
		logFile.WriteString(generateStageFinishLog(api.ImageBuildStageName, err))
		logFile.Close()
	}()
	logFile.WriteString(generateStageStartLog(api.ImageBuildStageName))

	go sm.cycloneClient.PushLogStream(sm.project, sm.pipeline, sm.recordID, api.IntegrationTestStageName, "", fileName, closeLog)

	eo := docker.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Container:    cid,
		OutputStream: logFile,
		ErrorStream:  logFile,
	}

	// Run the commands one by one.
	for _, cmd := range testConfig.Command {
		eo.Cmd = strings.Split(cmd, " ")
		if err = sm.dockerManager.ExecInContainer(eo); err != nil {
			return err
		}
	}

	return nil
}

func (sm *stageManager) StartServicesForIntegrationTest(services []api.Service) (map[string]string, error) {
	serviceInfos := make(map[string]string)
	for _, svc := range services {
		// Start and run the container from builder image.
		config := &docker_client.Config{
			Image:      svc.Image,
			Env:        convertEnvs(svc.EnvVars),
			Entrypoint: svc.Command,
		}

		cco := docker_client.CreateContainerOptions{
			Name:   svc.Name,
			Config: config,
			// HostConfig: hostConfig,
		}
		cid, err := sm.dockerManager.StartContainer(cco, generateAuthConfig(sm.registry))
		if err != nil {
			return nil, err
		}

		serviceInfos[svc.Name] = cid
	}

	return serviceInfos, nil
}

func (sm *stageManager) ExecImageRelease(builtImages []string, stage *api.ImageReleaseStage) (err error) {
	event.PipelineRecord.StageStatus.ImageRelease = &api.ImageReleaseStageStatus{
		GeneralStageStatus: api.GeneralStageStatus{
			Status:    api.Running,
			StartTime: time.Now(),
		},
	}
	sm.cycloneClient.SendEvent(event)

	closeLog := make(chan struct{})
	defer func() {
		if err != nil {
			event.PipelineRecord.StageStatus.ImageRelease.Status = api.Failed
			event.PipelineRecord.Status = api.Failed
			event.PipelineRecord.ErrorMessage = fmt.Sprintf("build images fail : %v", err)
		} else {
			event.PipelineRecord.StageStatus.ImageRelease.Status = api.Success
		}
		event.PipelineRecord.StageStatus.ImageRelease.EndTime = time.Now()
		sm.cycloneClient.SendEvent(event)

		time.Sleep(waitTime)
		close(closeLog)
	}()

	log.Infof("Exec image release stage for pipeline record %s/%s/%s", sm.project, sm.pipeline, sm.recordID)

	fileName := fmt.Sprintf(logFileNameTemplate, api.ImageReleaseStageName)
	logFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer logFile.Close()

	go sm.cycloneClient.PushLogStream(sm.project, sm.pipeline, sm.recordID, api.ImageReleaseStageName, "", fileName, closeLog)

	for _, p := range stage.ReleasePolicy {
		for _, builtImage := range builtImages {
			imageParts := strings.Split(builtImage, ":")
			if strings.EqualFold(imageParts[0], strings.Split(p.ImageName, ":")[0]) {
				log.Infof("Release the built image %s", builtImage)
				opts := docker_client.PushImageOptions{
					Name:         imageParts[0],
					Tag:          imageParts[1],
					OutputStream: logFile,
				}

				if err = sm.dockerManager.PushImage(opts, generateAuthConfig(sm.registry)); err != nil {
					log.Errorf("Fail to release the built image %s as %s", builtImage, err.Error())
					return err
				}

				setReleaseImages(builtImage)
			}
		}
	}

	return nil
}

func convertEnvs(envVars []api.EnvVar) []string {
	var envs []string
	for _, envVar := range envVars {
		envs = append(envs, fmt.Sprintf("%s=%s", envVar.Name, envVar.Value))
	}

	return envs
}

func watchLogs(filePath string, lines chan []byte, stop chan bool) error {
	log.Infof("watch log file: %s", filePath)
	logFile, err := os.Open(filePath)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	defer logFile.Close()

	buf := bufio.NewReader(logFile)

	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			log.Infoln("ticker")
			lines <- []byte("hello abc")
			line, errRead := buf.ReadBytes('\n')
			if errRead != nil {
				if errRead == io.EOF {
					return nil
				}
				log.Errorf("watch log file as errs: %s", errRead.Error())
				return errRead
			}
			log.Infof("log:%s", line)
			lines <- line
		case <-stop:
			close(lines)
			return nil
		}
	}
}

// replace the record name with default name '$commitID[:7]-$createTime' when name empty in create version
func formatPipelineRecordName(id string) {
	if event.PipelineRecord.Name == "" && id != "" {
		if len(id) > 7 {
			event.PipelineRecord.Name = fmt.Sprintf("%s-%s", id[:7], event.PipelineRecord.StartTime.Format("060102150405"))
		} else {
			event.PipelineRecord.Name = fmt.Sprintf("%s-%s", id, event.PipelineRecord.StartTime.Format("060102150405"))
		}

	}
}

func setCommit(commitLog *api.CommitLog, main bool) {
	if main {
		event.PipelineRecord.StageStatus.CodeCheckout.Commits.MainRepo = commitLog
	} else {
		event.PipelineRecord.StageStatus.CodeCheckout.Commits.DepRepos =
			append(event.PipelineRecord.StageStatus.CodeCheckout.Commits.DepRepos, commitLog)
	}

}

func setCommits(codeSources *api.CodeCheckoutStage) {
	commitLog, errl := scm.GetCommitLog(codeSources.MainRepo, "")
	if errl != nil {
		log.Warningf("get commit log fail %s", errl.Error())
	}
	formatPipelineRecordName(commitLog.ID)

	setCommit(&commitLog, true)
	for _, repo := range codeSources.DepRepos {
		commitLog, errl := scm.GetCommitLog(&repo.CodeSource, repo.Folder)
		if errl != nil {
			log.Warningf("get commit log fail %s", errl.Error())
		}

		setCommit(&commitLog, false)
	}
}

func setReleaseImages(image string) {
	if event.PipelineRecord.StageStatus.ImageRelease.Images == nil {
		event.PipelineRecord.StageStatus.ImageRelease.Images = []string{}
	}
	event.PipelineRecord.StageStatus.ImageRelease.Images = append(event.PipelineRecord.StageStatus.ImageRelease.Images, image)
}

/* formatImageName Ensure that the image name including a tag.
//  input        output
//  test:v1      test:v1
//  test         test:{rocordName}
*/
func formatImageName(namein string) string {
	var nameout string
	tname := strings.TrimSpace(namein)
	names := strings.Split(tname, ":")
	switch len(names) {
	case 1:
		nameout = names[0] + ":" + event.PipelineRecord.Name
	case 2:
		nameout = tname
	default:
		logdog.Error("image name error", logdog.Fields{"imageName": namein})
		nameout = tname
	}

	return nameout
}

// generateAuthConfig generates auth config for Docker registry.
func generateAuthConfig(registry *api.Registry) docker_client.AuthConfiguration {
	var auth docker_client.AuthConfiguration
	if registry != nil {
		auth = docker_client.AuthConfiguration{
			ServerAddress: registry.Server,
			Username:      registry.Username,
			Password:      registry.Password,
		}
	}

	return auth
}
