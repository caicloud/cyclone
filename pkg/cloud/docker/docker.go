/*
Copyright 2018 caicloud authors. All rights reserved.

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

package docker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-connections/tlsconfig"
	log "github.com/golang/glog"

	"github.com/caicloud/cyclone/cmd/worker/options"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/cloud"
	"github.com/caicloud/cyclone/pkg/worker/scm"
)

type dockerCloud struct {
	client *dockerclient.Client
}

func init() {
	if err := cloud.RegistryCloudProvider(api.CloudTypeDocker, NewDockerCloud); err != nil {
		log.Errorln(err)
	}
}

func NewDockerCloud(c *api.Cloud) (cloud.Provider, error) {
	if c.Type != api.CloudTypeDocker {
		err := fmt.Errorf("fail to new Docker cloud as cloud type %s is not %s", c.Type, api.CloudTypeDocker)
		log.Error(err)
		return nil, err
	}

	var cd *api.CloudDocker
	if c.Docker == nil {
		err := fmt.Errorf("Docker cloud %s is empty", c.Name)
		log.Error(err)
		return nil, err
	} else {
		cd = c.Docker
	}

	var httpClient *http.Client
	if cd.CertPath != "" {
		opts := tlsconfig.Options{
			CAFile:             filepath.Join(cd.CertPath, "ca.pem"),
			CertFile:           filepath.Join(cd.CertPath, "cert.pem"),
			KeyFile:            filepath.Join(cd.CertPath, "key.pem"),
			InsecureSkipVerify: cd.Insecure,
		}

		tlsc, err := tlsconfig.Client(opts)
		if err != nil {
			return nil, err
		}

		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:     tlsc,
				TLSHandshakeTimeout: cloud.DefaultCloudPingTimeout,
			},
		}
	} else {
		proto, addr, _, err := dockerclient.ParseHost(cd.Host)
		if err != nil {
			return nil, err
		}
		transport := new(http.Transport)
		sockets.ConfigureTransport(transport, proto, addr)
		transport.ResponseHeaderTimeout = cloud.DefaultCloudPingTimeout
		httpClient = &http.Client{
			Transport: transport,
		}
	}

	version := os.Getenv("DOCKER_API_VERSION")
	if version == "" {
		version = dockerclient.DefaultVersion
	}

	client, err := dockerclient.NewClient(cd.Host, version, httpClient, nil)
	if err != nil {
		return nil, err
	}

	return &dockerCloud{client}, nil
}

func (c *dockerCloud) Resource() (*options.Resource, error) {
	info, err := c.client.Info(context.Background())
	if err != nil {
		return nil, err
	}

	resource := options.NewResource()
	resource.Limit[options.ResourceLimitsCPU] = options.NewDecimalQuantity(info.NCPU)
	resource.Limit[options.ResourceLimitsMemory] = options.MustParseMemory(float64(info.MemTotal))

	// Used
	// count all used resource from running containers
	cs, err := c.client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}
	usedNanoCPU := 0.0 // NanoCPU
	usedMemory := 0.0
	for _, container := range cs {
		data, _ := c.client.ContainerInspect(context.Background(), container.ID)
		usedNanoCPU += float64(data.HostConfig.NanoCPUs)
		usedMemory += float64(data.HostConfig.Memory)
	}
	// NanoCPU in units of 1e-9 CPUs.
	resource.Used[options.ResourceLimitsCPU] = options.MustParseCPU(usedNanoCPU / 1e9)
	resource.Used[options.ResourceLimitsMemory] = options.MustParseMemory(usedMemory)

	return resource, nil
}

func (c *dockerCloud) CanProvision(quota options.Quota) (bool, error) {
	if quota == nil {
		return false, errors.New("need a valid quota")
	}

	resource, err := c.Resource()
	if err != nil {
		return false, err
	}

	if resource.Limit.IsZero() {
		return true, nil
	}

	if resource.Limit.Enough(resource.Used, quota) {
		return true, nil
	}

	return false, nil
}

func (c *dockerCloud) Provision(info *api.WorkerInfo, opts *options.WorkerOptions) (*api.WorkerInfo, error) {
	image := opts.WorkerImage
	eventID := opts.EventID
	config := &container.Config{
		Image:      image,
		Env:        buildDockerEnv(eventID, *opts),
		WorkingDir: scm.GetCloneDir(),
		Labels: map[string]string{
			"cyclone":    "worker",
			"cyclone/id": eventID,
		},
	}

	hostConfig := &container.HostConfig{
		Privileged: true,
	}

	ctx := context.Background()
	_, _, err := c.client.ImageInspectWithRaw(ctx, image)
	if err != nil {
		if dockerclient.IsErrImageNotFound(err) {
			if _, err = c.client.ImagePull(ctx, image, types.ImagePullOptions{}); err != nil {
				log.Error(err)
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	container, err := c.client.ContainerCreate(ctx, config, hostConfig, nil, info.Name)
	if err != nil {
		log.Error("fail to create container as %v", err)
		return nil, err
	}

	cid := container.ID
	if err = c.client.ContainerStart(ctx, cid, types.ContainerStartOptions{}); err != nil {
		return nil, c.client.ContainerRemove(ctx, cid, types.ContainerRemoveOptions{})
	}

	now := time.Now()
	info.StartTime = now
	info.DueTime = info.StartTime.Add(time.Duration(cloud.WorkerTimeout))

	return info, nil
}

func (c *dockerCloud) Ping() error {
	if _, err := c.client.Info(context.Background()); err != nil {
		log.Error("fail to ping Docker cloud as %v", err)
		return err
	}

	return nil
}

func (c *dockerCloud) TerminateWorker(cid string) error {
	cro := types.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	}

	return c.client.ContainerRemove(context.Background(), cid, cro)
}

// TODO (robin) Only need to pass worker envs as the param.
func buildDockerEnv(id string, opts options.WorkerOptions) []string {

	buildEnv := func(name string, value string) string {
		return fmt.Sprintf("%s=%s", name, value)
	}

	env := []string{
		buildEnv(options.EventID, id),
		buildEnv(options.CycloneServer, opts.CycloneServer),
		buildEnv(options.ConsoleWebEndpoint, opts.ConsoleWebEndpoint),
		buildEnv(options.RegistryLocation, opts.RegistryLocation),
		buildEnv(options.RegistryUsername, opts.RegistryUsername),
		buildEnv(options.RegistryPassword, opts.RegistryPassword),
		buildEnv(options.GitlabURL, opts.GitlabURL),
		buildEnv(options.WorkerImage, opts.WorkerImage),
		buildEnv(options.LimitCPU, opts.Quota[options.ResourceLimitsCPU].String()),
		buildEnv(options.LimitMemory, opts.Quota[options.ResourceLimitsMemory].String()),
	}

	return env
}

func (c *dockerCloud) ListWorkers() ([]api.WorkerInstance, error) {
	pods := []api.WorkerInstance{}

	ctx := context.Background()
	args := filters.NewArgs()
	args.Add("label", "cyclone=worker")
	opts := types.ContainerListOptions{
		Filters: args,
	}

	cycloneWorkers, err := c.client.ContainerList(ctx, opts)
	if err != nil {
		log.Errorf("list cyclone workers errorerr:%v", err)
		return pods, err
	}

	for _, worker := range cycloneWorkers {
		t := time.Unix(worker.Created, 0)
		// the origial name has extra forward slash in name
		name := strings.TrimSuffix(strings.TrimPrefix(worker.Names[0], "/"), "/")
		pod := api.WorkerInstance{
			Name:           name,
			Status:         worker.Status,
			CreationTime:   t,
			LastUpdateTime: t,
		}
		pods = append(pods, pod)
	}

	return pods, nil
}
