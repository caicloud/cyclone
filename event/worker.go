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

package event

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/docker"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/store"
	docker_client "github.com/fsouza/go-dockerclient"
)

// Worker is the type for Cyclone Worker.
type Worker struct {
	dockerHost  string
	dm          *docker.Manager
	containerID string
}

const (
	CYCLONE_SERVER_HOST = "CYCLONE_SERVER_HOST"
	WORKER_IMAGE        = "WORKER_IMAGE"
	WORK_DOCKER_HOST    = "WORK_DOCKER_HOST"

	// worker env setting
	WORKER_EVENTID = "WORKER_EVENTID"
	SERVER_HOST    = "SERVER_HOST"

	// worker time out
	WORKER_TIMEOUT = 7200 * time.Second

	WORK_REGISTRY_LOCATION = "WORK_REGISTRY_LOCATION"
	REGISTRY_USERNAME      = "REGISTRY_USERNAME"
	REGISTRY_PASSWORD      = "REGISTRY_PASSWORD"
	CONSOLE_WEB_ENDPOINT   = "CONSOLE_WEB_ENDPOINT"
	LOG_SERVER             = "LOG_SERVER"
	CLAIR_SERVER_IP        = "CLAIR_SERVER_IP"
	SERVER_GITLAB          = "SERVER_GITLAB"
	MEMORY_FOR_CONTAINER   = "MEMORY_FOR_CONTAINER"
	CPU_FOR_CONTAINER      = "CPU_FOR_CONTAINER"
)

var (
	ErrWorkerBusy = errors.New("Get worker docker host busy")
)

// RegistryCompose that compose the info about the registry
type RegistryCompose struct {
	// Registry's address, ie. cargo.caicloud.io
	RegistryLocation string `json:"registrylocation,omitempty"`
	// RegistryUsername used for operating the images
	RegistryUsername string `json:"registryusername,omitempty"`
	// RegistryPassword used for operating the images
	RegistryPassword string `json:"registrypassword,omitempty"`
}

// NewWorker new a worker
func NewWorker(event *api.Event) (*Worker, error) {
	dockerHostWorker, err := GetWorkerDockerHost(event)
	if err != nil {
		return nil, err
	}

	dockerManager, err := docker.NewManager(dockerHostWorker, "", registryWorker)
	//dockerManager, err := docker.NewManager(dockerHostWorker, certPathWorker, registryWorker)
	if err != nil {
		return nil, err
	}
	w := &Worker{
		dockerHost:  dockerHostWorker,
		dm:          dockerManager,
		containerID: event.WorkerInfo.ContainerID,
	}

	return w, nil
}

// LoadWorker load worker from event
func LoadWorker(event *api.Event) (*Worker, error) {
	if "" == event.WorkerInfo.ContainerID {
		return nil, fmt.Errorf("event with empty workerinfo")
	}

	dockerHostWorker := event.WorkerInfo.DockerHost
	dockerManager, err := docker.NewManager(dockerHostWorker, certPathWorker, registryWorker)
	if err != nil {
		return nil, err
	}
	w := &Worker{
		dockerHost:  dockerHostWorker,
		dm:          dockerManager,
		containerID: event.WorkerInfo.ContainerID,
	}

	return w, nil
}

// GetWorkerDockerHost get woker docker host LB according to node resource
func GetWorkerDockerHost(event *api.Event) (string, error) {
	if event.Operation == CreateVersionOps {
		event.WorkerInfo.UsedResource = event.Version.BuildResource
	} else {
		event.WorkerInfo.UsedResource.Memory = osutil.GetFloat64Env(MEMORY_FOR_CONTAINER, 536870912.0) //512M
		event.WorkerInfo.UsedResource.CPU = osutil.GetFloat64Env(CPU_FOR_CONTAINER, 512.0)
	}

	ds := store.NewStore()
	defer ds.Close()
	workerNodes, err := ds.FindSystemWorkerNodeByResource(&(event.WorkerInfo.UsedResource))
	if err != nil {
		log.Errorf("Get worker docker host err %v", err)
		return "", err
	}
	if len(workerNodes) == 0 {
		log.Errorf("Get worker docker host busy")
		return "", ErrWorkerBusy
	}

	err = resourceManager.ApplyResource(event)
	if err != nil {
		log.Errorf("apply resource err %v", err)
		return "", err
	}

	workerDockerHost := workerNodes[0].DockerHost
	log.Infof("Get worker docker host: %s", workerDockerHost)
	workerNodes[0].LeftResource.CPU -= event.WorkerInfo.UsedResource.CPU
	workerNodes[0].LeftResource.Memory -= event.WorkerInfo.UsedResource.Memory
	_, err = ds.UpsertWorkerNodeDocument(&workerNodes[0])
	if err != nil {
		log.Errorf("Update woker node err %v", err)
		return "", err
	}

	return workerDockerHost, nil
}

// DoWork create a container start do work
func (w *Worker) DoWork(event *api.Event) (err error) {
	coo := toBuildContainerConfig(event.EventID, int64(event.WorkerInfo.UsedResource.CPU),
		int64(event.WorkerInfo.UsedResource.Memory))
	w.containerID, err = w.dm.RunContainer(coo)
	if err != nil {
		w.dm.StopContainer(w.containerID)
		w.dm.RemoveContainer(w.containerID) // release resource of worker node
		ds := store.NewStore()
		defer ds.Close()
		nodes, errinfo := ds.FindWorkerNodesByDockerHost(event.WorkerInfo.DockerHost)
		if errinfo != nil || len(nodes) != 1 {
			log.Errorf("find worker node err: %v", err)
		} else {
			nodes[0].LeftResource.Memory += event.WorkerInfo.UsedResource.Memory
			nodes[0].LeftResource.CPU += event.WorkerInfo.UsedResource.CPU
			_, err := ds.UpsertWorkerNodeDocument(&(nodes[0]))
			if err != nil {
				log.Errorf("release worker node resource err: %v", err)
			}
		}
		return err
	}

	event.WorkerInfo.DockerHost = w.dockerHost
	event.WorkerInfo.ContainerID = w.containerID
	event.WorkerInfo.DueTime = time.Now().Add(time.Duration(WORKER_TIMEOUT))
	err = SaveEventToEtcd(event)
	log.Infof("save event worker info: %s, %v", w.containerID, err)
	go CheckWorkerTimeOut(*event)
	return nil
}

// Fire fire a worker, stop and remove the worker container
func (w *Worker) Fire() error {
	// stop worker container
	err := w.dm.StopContainer(w.containerID)
	if err != nil {
		log.Errorf("stop err: %v", err)
	}

	// remove worker container
	err = w.dm.RemoveContainer(w.containerID)
	if err != nil {
		log.Errorf("remove err: %v", err)
	}
	return err
}

// CheckWorkerTimeOut ensures that the events are not timed out.
func CheckWorkerTimeOut(e api.Event) {
	var eventCopy api.Event
	eventCopy = e
	event := &eventCopy
	if IsEventFinished(event) {
		return
	}

	now := time.Now()
	dueTime := event.WorkerInfo.DueTime
	var err error
	// has time out
	if !dueTime.After(now) {
		log.Infof("event has time out: %v", event)

		// save event to etcd
		event.Status = api.EventStatusFail
		SaveEventToEtcd(event)
		return
	}
	remain := event.WorkerInfo.DueTime.Sub(now)
	time.Sleep(remain)

	event, err = LoadEventFromEtcd(event.EventID)
	if err != nil {
		return
	}

	if !IsEventFinished(event) {
		log.Infof("event time out: %v", event)
		event.Status = api.EventStatusFail
		SaveEventToEtcd(event)
	}
}

// toContainerConfig creates CreateContainerOptions from BuildNode.
func toBuildContainerConfig(eventID api.EventID, cpu, memory int64) *docker_client.CreateContainerOptions {
	workerImage := osutil.GetStringEnv(WORKER_IMAGE, "cargo.caicloud.io/caicloud/cyclone-worker")
	serverHost := osutil.GetStringEnv(CYCLONE_SERVER_HOST, "http://127.0.0.1:7099")
	registryLocation := osutil.GetStringEnv(WORK_REGISTRY_LOCATION, "")
	registryUsername := osutil.GetStringEnv(REGISTRY_USERNAME, "")
	registryPassword := osutil.GetStringEnv(REGISTRY_PASSWORD, "")
	consoleWebEndpoint := osutil.GetStringEnv(CONSOLE_WEB_ENDPOINT, "http://127.0.0.1:3000")
	clairServerIP := osutil.GetStringEnv(CLAIR_SERVER_IP, "http://127.0.0.1:6060")
	gitlabServer := osutil.GetStringEnv("SERVER_GITLAB", "https://gitlab.com")
	logServer := osutil.GetStringEnv(LOG_SERVER, "ws://127.0.0.1:8000/ws")

	envEventID := fmt.Sprintf("%s=%s", WORKER_EVENTID, string(eventID))
	envServerHost := fmt.Sprintf("%s=%s", SERVER_HOST, serverHost)
	envregistryLocation := fmt.Sprintf("%s=%s", WORK_REGISTRY_LOCATION, registryLocation)
	envregistryUsername := fmt.Sprintf("%s=%s", REGISTRY_USERNAME, registryUsername)
	envregistryPassword := fmt.Sprintf("%s=%s", REGISTRY_PASSWORD, registryPassword)
	envconsoleWebEndpoint := fmt.Sprintf("%s=%s", CONSOLE_WEB_ENDPOINT, consoleWebEndpoint)
	envclairServerIP := fmt.Sprintf("%s=%s", CLAIR_SERVER_IP, clairServerIP)
	envgitlabServer := fmt.Sprintf("%s=%s", SERVER_GITLAB, gitlabServer)
	envLogServer := fmt.Sprintf("%s=%s", LOG_SERVER, logServer)

	config := &docker_client.Config{
		Image: workerImage,
		Env: []string{envEventID, envServerHost, envregistryLocation, envregistryUsername, envregistryPassword,
			envconsoleWebEndpoint, envclairServerIP, envgitlabServer, envLogServer},
	}

	hostConfig := &docker_client.HostConfig{
		Privileged: true,
		//NetworkMode: "host",
		AutoRemove: true,
		CPUShares:  cpu,
		Memory:     memory,
	}

	createContainerOptions := &docker_client.CreateContainerOptions{
		Config:     config,
		HostConfig: hostConfig,
	}
	return createContainerOptions
}

// traceScript is a helper script that is added
// to the build script to trace a command.
const traceScript = `
echo %s | base64 -d
%s
`

// writeCmds is a helper fuction that writes a slice
// of bash commands as a single script.
func writeCmds(cmds []string) string {
	var buf bytes.Buffer
	for _, cmd := range cmds {
		buf.WriteString(trace(cmd))
	}
	return buf.String()
}

// trace is a helper function that allows us to echo
// commands back to the console for debugging purposes.
func trace(cmd string) string {
	echo := fmt.Sprintf("$ %s\n", cmd)
	base := base64.StdEncoding.EncodeToString([]byte(echo))
	return fmt.Sprintf(traceScript, base, cmd)
}

// encode is a helper function that base64 encodes
// a shell command (or entire script)
func encode(script []byte) string {
	encoded := base64.StdEncoding.EncodeToString(script)
	return fmt.Sprintf("echo %s | base64 -d | /bin/sh", encoded)
}
