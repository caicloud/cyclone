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

package cloud

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	apiclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/reference"
	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/zoumo/logdog"
)

// DockerCloud is a docker cloud
type DockerCloud struct {
	name           string
	host           string
	insecure       bool
	dockerCertPath string
	client         *apiclient.Client
}

// NewDockerCloud returns a new Cloud from CloudOption
func NewDockerCloud(opts Options) (Cloud, error) {
	if opts.Name == "" {
		return nil, errors.New("DockerCloud: Invalid cloud name")
	}
	if opts.Host == "" {
		return nil, errors.New("DockerCloud: Invalid cloud host")
	}

	cloud := &DockerCloud{
		name:           opts.Name,
		host:           opts.Host,
		insecure:       opts.Insecure,
		dockerCertPath: opts.DockerCertPath,
	}

	var httpClient *http.Client

	if opts.DockerCertPath != "" {
		// TLS
		options := tlsconfig.Options{
			CAFile:             filepath.Join(opts.DockerCertPath, "ca.pem"),
			CertFile:           filepath.Join(opts.DockerCertPath, "cert.pem"),
			KeyFile:            filepath.Join(opts.DockerCertPath, "key.pem"),
			InsecureSkipVerify: opts.Insecure,
		}
		tlsc, err := tlsconfig.Client(options)
		if err != nil {
			return nil, err
		}

		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:     tlsc,
				TLSHandshakeTimeout: DefaultCloudPingTimeout,
			},
		}
	} else {
		// Normal
		proto, addr, _, err := apiclient.ParseHost(cloud.host)
		if err != nil {
			return nil, err
		}
		transport := new(http.Transport)
		sockets.ConfigureTransport(transport, proto, addr)
		transport.ResponseHeaderTimeout = DefaultCloudPingTimeout
		httpClient = &http.Client{
			Transport: transport,
		}
	}

	version := os.Getenv("DOCKER_API_VERSION")
	if version == "" {
		version = apiclient.DefaultVersion
	}

	client, err := apiclient.NewClient(cloud.host, version, httpClient, nil)
	if err != nil {
		return nil, err
	}
	cloud.client = client
	return cloud, nil
}

// Client returns docker client
func (cloud *DockerCloud) Client() *apiclient.Client {
	return cloud.client
}

// Name returns cloud name.
func (cloud *DockerCloud) Name() string {
	return cloud.name
}

// Kind returns cloud type.
func (cloud *DockerCloud) Kind() string {
	return KindDockerCloud
}

// Ping returns nil if cloud is accessible
func (cloud *DockerCloud) Ping() error {
	_, err := cloud.client.Info(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// Resource returns the limit and used quotas of the cloud
func (cloud *DockerCloud) Resource() (*Resource, error) {
	info, err := cloud.client.Info(context.Background())
	if err != nil {
		return nil, err
	}
	resource := &Resource{
		Limit: ZeroQuota.DeepCopy(),
		Used:  ZeroQuota.DeepCopy(),
	}

	// Limit
	resource.Limit[ResourceLimitsCPU] = NewDecimalQuantity(info.NCPU)
	resource.Limit[ResourceLimitsMemory] = MustParseMemory(float64(info.MemTotal))

	// Used
	// count all used resource from running containers
	cs, err := cloud.client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}
	usedNanoCPU := 0.0 // NanoCPU
	usedMemory := 0.0
	for _, container := range cs {
		data, _ := cloud.client.ContainerInspect(context.Background(), container.ID)
		usedNanoCPU += float64(data.HostConfig.NanoCPUs)
		usedMemory += float64(data.HostConfig.Memory)
	}
	// NanoCPU in units of 1e-9 CPUs.
	resource.Used[ResourceLimitsCPU] = MustParseCPU(usedNanoCPU / 1e9)
	resource.Used[ResourceLimitsMemory] = MustParseMemory(usedMemory)

	return resource, nil
}

// CanProvision returns true if the cloud can provision a worker meetting the quota
func (cloud *DockerCloud) CanProvision(need Quota) (bool, error) {

	if need == nil {
		return false, errors.New("CanProvision: need a valid quota")
	}

	resource, err := cloud.Resource()
	if err != nil {
		return false, err
	}
	if resource.Limit.IsZero() {
		return true, nil
	}
	if resource.Limit.Enough(resource.Used, need) {
		return true, nil
	}
	return false, nil
}

// Provision returns a worker if the cloud can provison
func (cloud *DockerCloud) Provision(id string, wopts WorkerOptions) (Worker, error) {

	ok, err := cloud.CanProvision(wopts.Quota)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, ErrNoEnoughResource
	}

	config := &container.Config{
		Image:      wopts.WorkerEnvs.WorkerImage,
		Env:        buildDockerEnv(id, wopts),
		WorkingDir: WorkingDir,
	}

	hostConfig := &container.HostConfig{
		Privileged: true,
		// AutoRemove: true,
		Resources: wopts.Quota.ToDockerQuota(),
	}

	worker := &DockerWorker{
		name:       id,
		config:     config,
		hostConfig: hostConfig,
		cloud:      cloud,
	}

	return worker, nil
}

// LoadWorker rebuilds a worker from worker info
func (cloud *DockerCloud) LoadWorker(info WorkerInfo) (Worker, error) {

	if cloud.Kind() != info.CloudKind {
		return nil, fmt.Errorf("DockerCloud: can not load worker with another cloud kind %s", info.CloudKind)
	}

	worker := &DockerWorker{
		name:        info.Name,
		cloud:       cloud,
		containerID: info.ContainerID,
		createTime:  info.CreateTime,
		dueTime:     info.DueTime,
	}

	return worker, nil
}

// GetOptions ...
func (cloud *DockerCloud) GetOptions() Options {
	return Options{
		Name:           cloud.name,
		Kind:           cloud.Kind(),
		Host:           cloud.host,
		Insecure:       cloud.insecure,
		DockerCertPath: cloud.dockerCertPath,
	}
}

// --------------------------------------------------------------------------------

// DockerWorker ...
type DockerWorker struct {
	name        string
	containerID string
	config      *container.Config
	hostConfig  *container.HostConfig
	cloud       *DockerCloud
	createTime  time.Time
	dueTime     time.Time
}

// runContainer creates and run a container, returns the container id
func (worker *DockerWorker) runContainer(ctx context.Context) (string, error) {

	// add default tag if need
	_, ref, err := reference.ParseIDOrReference(worker.config.Image)
	if err != nil {
		return "", err
	}
	if ref != nil {
		ref = reference.WithDefaultTag(ref)
		if ref, ok := ref.(reference.NamedTagged); ok {
			worker.config.Image = ref.String()
		}
	}

	// create
	created, err := worker.cloud.Client().ContainerCreate(ctx, worker.config, worker.hostConfig, nil, worker.name)
	if err != nil {
		// if image not found, pull image and retry
		if apiclient.IsErrImageNotFound(err) && ref != nil {
			resp, err := worker.cloud.Client().ImagePull(ctx, worker.config.Image, types.ImagePullOptions{})
			if err != nil {
				return "", err
			}
			defer resp.Close()
			// TODO add log
			logdog.Debug("Unable to find image locally. Pulling...", logdog.Fields{"image": ref.String()})
			err = jsonmessage.DisplayJSONMessagesStream(resp, os.Stderr, 0, false, nil)
			if err != nil {
				return "", err
			}
			// create again
			var retryErr error
			created, retryErr = worker.cloud.Client().ContainerCreate(ctx, worker.config, worker.hostConfig, nil, worker.name)
			if retryErr != nil {
				return "", retryErr
			}
		} else {
			return "", err
		}
	}

	// get id
	containerID := created.ID
	worker.containerID = containerID

	// start
	err = worker.cloud.Client().ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		// force remove
		worker.Terminate()
		return "", err
	}

	// add time
	worker.createTime = time.Now()
	worker.dueTime = worker.createTime.Add(time.Duration(WorkerTimeout))

	return containerID, nil
}

// Do starts the worker and do the work
func (worker *DockerWorker) Do() error {
	ctx := context.Background()

	_, err := worker.runContainer(ctx)
	if err != nil {
		return err
	}

	return nil
}

// GetWorkerInfo returns worker's infomation
func (worker *DockerWorker) GetWorkerInfo() WorkerInfo {
	return WorkerInfo{
		CloudName:   worker.cloud.Name(),
		CloudKind:   worker.cloud.Kind(),
		CreateTime:  worker.createTime,
		DueTime:     worker.dueTime,
		ContainerID: worker.containerID,
	}
}

// IsTimeout returns true if worker is timeout
// and returns the time left until it is due
func (worker *DockerWorker) IsTimeout() (bool, time.Duration) {
	now := time.Now()
	if now.After(worker.dueTime) {
		return true, time.Duration(0)
	}
	return false, worker.dueTime.Sub(now)
}

// Terminate terminates the worker and destroy it
func (worker *DockerWorker) Terminate() error {
	if worker.containerID == "" {
		return nil
	}
	ctx := context.TODO()
	logdog.Debug("worker terminating...", logdog.Fields{"cloud": worker.cloud.Name(), "kind": worker.cloud.Kind(), "containerID": worker.containerID})

	if Debug {
		readCloser, err := worker.cloud.Client().ContainerLogs(
			ctx,
			worker.containerID,
			types.ContainerLogsOptions{ShowStderr: true, ShowStdout: true},
		)
		if err != nil {
			logdog.Error("Can not read log from container", logdog.Fields{
				"cloud":       worker.cloud.Name(),
				"kind":        worker.cloud.Kind(),
				"containerID": worker.containerID,
				"err":         err,
			})
		} else {
			defer readCloser.Close()
			content, _ := ioutil.ReadAll(readCloser)
			logdog.Debug(string(content))
		}
	}

	err := worker.cloud.Client().ContainerRemove(
		ctx,
		worker.containerID,
		types.ContainerRemoveOptions{
			Force:         true,
			RemoveVolumes: true,
			// RemoveLinks:   true,
		})

	return err
}

func buildDockerEnv(id string, opts WorkerOptions) []string {

	buildEnv := func(name string, value string) string {
		return fmt.Sprintf("%s=%s", name, value)
	}

	env := []string{
		// special
		buildEnv(WorkerEventID, id),

		buildEnv(CycloneServer, opts.WorkerEnvs.CycloneServer),
		buildEnv(ConsoleWebEndpoint, opts.WorkerEnvs.ConsoleWebEndpoint),
		buildEnv(RegistryLocation, opts.WorkerEnvs.RegistryLocation),
		buildEnv(RegistryUsername, opts.WorkerEnvs.RegistryUsername),
		buildEnv(RegistryPassword, opts.WorkerEnvs.RegistryPassword),
		buildEnv(ClairDisable, strconv.FormatBool(opts.WorkerEnvs.ClairDisable)),
		buildEnv(ClairServer, opts.WorkerEnvs.ClairServer),
		buildEnv(GitlabURL, opts.WorkerEnvs.GitlabURL),
		buildEnv(LogServer, opts.WorkerEnvs.LogServer),
		buildEnv(WorkerImage, opts.WorkerEnvs.WorkerImage),
		buildEnv(LimitCPU, opts.Quota[ResourceLimitsCPU].String()),
		buildEnv(LimitMemory, opts.Quota[ResourceLimitsMemory].String()),
	}

	return env
}
