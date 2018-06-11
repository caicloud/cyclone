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
	"fmt"
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
var imageNamePrefix string

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

	if registry != nil {
		imageNamePrefix = getImagePrifix(registry.Server, registry.Repository)
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
	errChan := make(chan error)
	defer func() {
		errChan <- err
		sm.updateEventAfterStage(api.CodeCheckoutStageName, err)
	}()

	logFile, err := sm.startWatchLogs(api.CodeCheckoutStageName, "", errChan)
	if err != nil {
		logdog.Error(err.Error())
		return err
	}

	logs, err := scm.CloneRepos(token, stage, sm.performParams.Ref)
	if err != nil {
		logdog.Error(err.Error())
		logFile.WriteString(err.Error())
		return err
	} else {
		// Just one line of log, will add more detailed logs.
		logFile.WriteString(logs)
	}

	setCommits(stage)
	return nil
}

func (sm *stageManager) ExecPackage(builderImage *api.BuilderImage, buildInfo *api.BuildInfo, unitTestStage *api.UnitTestStage, packageStage *api.PackageStage) (err error) {
	errChan := make(chan error)
	defer func() {
		errChan <- err
		sm.updateEventAfterStage(api.PackageStageName, err)
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

	logFile, err := sm.startWatchLogs(api.PackageStageName, "", errChan)
	if err != nil {
		log.Infof("fail to watch package log as %v", err)
		return
	}

	// Execute unit test and package commands in the builder container.
	// Run stage script in container
	cmds := packageStage.Command
	// Run the unit test commands before package commands if there is unit test stage.
	if unitTestStage != nil {
		cmds = append(unitTestStage.Command, cmds...)
	}

	// Start and run the container from builder image.
	config := &docker_client.Config{
		Image:      builderImage.Image,
		Env:        convertEnvs(builderImage.EnvVars),
		OpenStdin:  true, // Open stdin to keep the container running after starts.
		WorkingDir: cloneDir,
		Cmd:        cmds,
	}

	cco := docker_client.CreateContainerOptions{
		Config:     config,
		HostConfig: hostConfig,
	}

	var cid string
	defer func() {
		if err := sm.dockerManager.RemoveContainer(cid); err != nil {
			log.Errorf("Fail to remove the container %s", cid)
		}
	}()

	cid, err = sm.dockerManager.StartContainer(cco, generateAuthConfig(sm.registry), logFile)
	if err != nil {
		return err
	}

	// Copy the build outputs if necessary.
	// Only need to copy the outputs not in the current workspace.
	// The outputs must be absolute path of files or folders.
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

func (sm *stageManager) ExecImageBuild(stage *api.ImageBuildStage) (builtImages []string, err error) {
	defer func() {
		sm.updateEventAfterStage(api.ImageBuildStageName, err)
	}()

	// New ImageBuildStageStatus to store task status.
	if event.PipelineRecord.StageStatus.ImageBuild == nil {
		event.PipelineRecord.StageStatus.ImageBuild = &api.ImageBuildStageStatus{}
	}
	imageBuildStatus := event.PipelineRecord.StageStatus.ImageBuild

	wg := sync.WaitGroup{}
	buildInfos := stage.BuildInfos
	imageBuildStatus.Tasks = make([]*api.ImageBuildTaskStatus, len(buildInfos))
	for i, _ := range buildInfos {
		wg.Add(1)

		go func(index int) {
			defer wg.Done()

			status := &api.ImageBuildTaskStatus{
				TaskStatus: api.TaskStatus{
					Name:      buildInfos[index].TaskName,
					Status:    api.Running,
					StartTime: time.Now(),
				},
			}
			imageBuildStatus.Tasks[index] = status

			buildInfo := buildInfos[index]
			if bErr := sm.buildImage(buildInfo, status); bErr != nil {
				log.Errorf("Fail to build image %s as %v", buildInfo.TaskName, bErr)
				err = bErr
			}
			// Update status again after tasks finish.
			imageBuildStatus.Tasks[index] = status
		}(i)
	}

	wg.Wait()
	if err != nil {
		return nil, err
	}

	for _, task := range imageBuildStatus.Tasks {
		if task.Status == api.Success {
			builtImages = append(builtImages, task.Image)
		} else {
			log.Errorf("Image build task %s's status is %s", task.Name, task.Status)
		}
	}

	return builtImages, err
}

func (sm *stageManager) buildImage(buildInfo *api.ImageBuildInfo, status *api.ImageBuildTaskStatus) (err error) {
	errChan := make(chan error)

	defer func() {
		// Update task status.
		status.EndTime = time.Now()
		if err == nil {
			status.Status = api.Success
		} else {
			status.Status = api.Failed
		}

		errChan <- err
	}()

	logFile, err := sm.startWatchLogs(api.ImageBuildStageName, buildInfo.TaskName, errChan)
	if err != nil {
		log.Infof("fail to watch image build log as %v", err)
		return
	}

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
			return
		}
	} else {
		if buildInfo.DockerfilePath != "" {
			opt.Dockerfile = strings.TrimPrefix(strings.TrimPrefix(buildInfo.DockerfilePath, buildInfo.ContextDir), "/")
		}
	}

	opt.Name = formatImageName(buildInfo.ImageName)

	if err = sm.dockerManager.BuildImage(opt); err != nil {
		return
	}

	status.Image = opt.Name

	return
}

func (sm *stageManager) ExecIntegrationTest(builtImages []string, stage *api.IntegrationTestStage) (err error) {
	errChan := make(chan error)
	defer func() {
		errChan <- err
		sm.updateEventAfterStage(api.IntegrationTestStageName, err)
	}()

	log.Infof("Exec integration test stage for pipeline record %s/%s/%s", sm.project, sm.pipeline, sm.recordID)

	logFile, err := sm.startWatchLogs(api.IntegrationTestStageName, "", errChan)
	if err != nil {
		log.Infof("fail to watch integration test log as %v", err)
		return
	}

	var serviceInfos map[string]string
	defer func() {
		var err error
		for s, cid := range serviceInfos {
			if err = sm.dockerManager.RemoveContainer(cid); err != nil {
				log.Errorf("Fail to remove container %s for the service %s", cid, s)
			}
		}
	}()

	// Start the services.
	serviceInfos, err = sm.StartServicesForIntegrationTest(builtImages, stage.Services)
	if err != nil {
		return err
	}

	testConfig := stage.Config

	included, testImage := getBuiltImageName(builtImages, testConfig.ImageName)
	if !included {
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
		Cmd:       testConfig.Command,
	}

	hostConfig := &docker_client.HostConfig{
		Links: serviceNames,
	}
	cco := docker_client.CreateContainerOptions{
		Config:     config,
		HostConfig: hostConfig,
	}

	var cid string
	defer func() {
		if err := sm.dockerManager.RemoveContainer(cid); err != nil {
			log.Errorf("Fail to remove the container %s", cid)
		}
	}()

	cid, err = sm.dockerManager.StartContainer(cco, generateAuthConfig(sm.registry), logFile)
	if err != nil {
		return err
	}

	return nil
}

func (sm *stageManager) StartServicesForIntegrationTest(builtImages []string, services []api.Service) (map[string]string, error) {
	serviceInfos := make(map[string]string)
	for _, svc := range services {

		_, imageName := getBuiltImageName(builtImages, svc.Image)
		// Start and run the container from builder image.
		config := &docker_client.Config{
			Image: imageName,
			Env:   convertEnvs(svc.EnvVars),
			Cmd:   svc.Command,
		}

		cco := docker_client.CreateContainerOptions{
			Name:   svc.Name,
			Config: config,
			// HostConfig: hostConfig,
		}

		cid, err := sm.dockerManager.StartContainer(cco, generateAuthConfig(sm.registry), nil)
		if err != nil {
			return nil, err
		}

		serviceInfos[svc.Name] = cid
	}

	return serviceInfos, nil
}

// if image already built in 'builtImages', return true and the built image name,
// otherwise return false and origin image name.
func getBuiltImageName(builtImages []string, image string) (bool, string) {
	for _, builtImage := range builtImages {

		// trim prefix
		tempBuiltImage := strings.TrimPrefix(builtImage, imageNamePrefix)
		tempImage := strings.TrimPrefix(image, imageNamePrefix)

		// is image name equal
		if strings.EqualFold(strings.Split(tempBuiltImage, ":")[0], strings.Split(tempImage, ":")[0]) {
			return true, builtImage
		}

	}

	return false, image
}

func (sm *stageManager) ExecImageRelease(builtImages []string, stage *api.ImageReleaseStage) (err error) {
	log.Infof("Exec image release stage for pipeline record %s/%s/%s", sm.project, sm.pipeline, sm.recordID)

	defer func() {
		sm.updateEventAfterStage(api.ImageReleaseStageName, err)
	}()

	if event.PipelineRecord.StageStatus.ImageRelease == nil {
		event.PipelineRecord.StageStatus.ImageRelease = &api.ImageReleaseStageStatus{}
	}
	imageReleaseStatus := event.PipelineRecord.StageStatus.ImageRelease

	wg := sync.WaitGroup{}
	policies := stage.ReleasePolicies
	imageReleaseStatus.Tasks = make([]*api.ImageReleaseTaskStatus, len(policies))
	for i, p := range policies {
		included, builtImage := getBuiltImageName(builtImages, p.ImageName)
		if included {
			wg.Add(1)

			go func(index int, image string) {
				defer wg.Done()

				status := &api.ImageReleaseTaskStatus{
					TaskStatus: api.TaskStatus{
						Name:      image,
						Status:    api.Running,
						StartTime: time.Now(),
					},
				}
				imageReleaseStatus.Tasks[index] = status

				if bErr := sm.releaseImage(image, status); bErr != nil {
					log.Errorf("Fail to release image %s as %v", image, bErr)
					err = bErr
				}

				// Update status again after tasks finish.
				imageReleaseStatus.Tasks[index] = status
			}(i, builtImage)
		}
	}

	wg.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (sm *stageManager) releaseImage(image string, status *api.ImageReleaseTaskStatus) (err error) {
	log.Infof("Release the built image %s", image)

	errChan := make(chan error)

	defer func() {
		// Update task status.
		status.EndTime = time.Now()
		if err == nil {
			status.Status = api.Success
		} else {
			status.Status = api.Failed
		}

		errChan <- err
	}()

	imageParts := strings.Split(image, ":")
	imageNameParts := strings.Split(imageParts[0], "/")
	imageName := imageNameParts[len(imageNameParts)-1]

	logFile, err := sm.startWatchLogs(api.ImageReleaseStageName, imageName, errChan)
	if err != nil {
		log.Infof("fail to watch image release log as %v", err)
		return
	}

	opts := docker_client.PushImageOptions{
		Name:         imageParts[0],
		Tag:          imageParts[1],
		OutputStream: logFile,
	}
	if err = sm.dockerManager.PushImage(opts, generateAuthConfig(sm.registry)); err != nil {
		log.Errorf("Fail to release the built image %s as %s", image, err.Error())
		return
	}

	status.Image = image

	return
}

func (sm *stageManager) updateEventAfterStage(stage api.PipelineStageName, err error) {
	if err != nil {
		updateEvent(sm.cycloneClient, event, stage, api.Failed, err)
	} else {
		updateEvent(sm.cycloneClient, event, stage, api.Success, nil)
	}
}

func (sm *stageManager) startWatchLogs(stage api.PipelineStageName, task string, errChan chan error) (*os.File, error) {
	if err := updateEvent(sm.cycloneClient, event, stage, api.Running, nil); err != nil {
		logdog.Errorf("fail to update event for stage %s as %v", stage, err)
		return nil, err
	}

	fileName := fmt.Sprintf(logFileNameTemplate, stage)
	if task != "" {
		fileName = fmt.Sprintf(logFileNameTemplate, fmt.Sprintf("%s-%s", stage, task))
	}

	logFile, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	if _, err = logFile.WriteString(generateStageStartLog(stage)); err != nil {
		logFile.Close()
		return nil, err
	}

	closeLog := make(chan struct{})
	go sm.cycloneClient.PushLogStream(sm.project, sm.pipeline, sm.recordID, stage, task, fileName, closeLog)

	go func() {
		err := <-errChan
		logFile.WriteString(generateStageFinishLog(stage, err))

		// Wait for a while to ensure finish log of stages are reported.
		time.Sleep(waitTime)

		close(closeLog)
		logFile.Close()
	}()

	return logFile, nil
}

func convertEnvs(envVars []api.EnvVar) []string {
	var envs []string
	for _, envVar := range envVars {
		envs = append(envs, fmt.Sprintf("%s=%s", envVar.Name, envVar.Value))
	}

	return envs
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

// formatImageName Ensure that the image name matching a format --- registry[/repository]/name:tag
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

	return imageNamePrefix + nameout
}

func getImagePrifix(registry, repository string) string {
	prefix := registry + "/"
	if repository != "" {
		prefix = prefix + repository + "/"
	}
	return prefix
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
