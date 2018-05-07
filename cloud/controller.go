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
	"errors"
	"fmt"

	"github.com/caicloud/cyclone/pkg/register"
	"github.com/zoumo/logdog"
)

const (
	// TODO (robin) User CloudType instead, remove when refactor cloud.

	// KindDockerCloud ...
	KindDockerCloud = "docker"
	// KindK8SCloud ...
	KindK8SCloud = "kubernetes"
)

var (
	// Debug ...
	Debug = false
	// CloudFactory ...
	CloudFactory = register.NewRegister()
)

func init() {
	RegisterConstructor(string(CloudTypeDocker), NewDockerCloud)
	RegisterConstructor(string(CloudTypeKubernetes), NewK8SCloud)
}

// Constructor ...
type Constructor func(Options) (CloudProvider, error)

// RegisterConstructor ...
func RegisterConstructor(kind string, cc Constructor) {
	CloudFactory.Register(kind, cc)
}

// GetConstructor ...
func GetConstructor(kind string) Constructor {
	v := CloudFactory.Get(kind)
	if v == nil {
		return nil
	}
	return v.(Constructor)
}

// Controller ...
// TODO cloud controller sync all cloud infomation
type Controller struct {
	Clouds map[string]CloudProvider
	// WorkerOptions *WorkerOptions
	provisionErr *ErrCloudProvision
}

// NewController creates a new CloudController
func NewController() *Controller {
	return &Controller{
		Clouds: make(map[string]CloudProvider),
		// WorkerOptions: NewWorkerOptions(),
		provisionErr: NewErrCloudProvision(),
	}
}

// AddClouds creates clouds from CloudOptions and cache it
func (cc *Controller) AddClouds(clouds ...Cloud) error {
	for _, c := range clouds {
		kind := string(c.Type)
		constructor := GetConstructor(kind)
		if constructor == nil {
			err := fmt.Errorf("CloudController: no cloud constructor found for kind[%s]", kind)
			logdog.Error(err)
			return err
		}

		opt := convertCloudToOption(c)
		cloud, err := constructor(opt)
		if err != nil {
			logdog.Error("CloudController: create cloud error", logdog.Fields{"err": err})
			return err
		}
		cc.Clouds[c.Name] = cloud
		logdog.Debug("CloudController: add cloud successfully", logdog.Fields{"name": c.Name, "kind": kind})
	}
	return nil
}

// GetCloud returns a cloud by cloud name
func (cc *Controller) GetCloud(name string) (CloudProvider, bool) {
	cloud, ok := cc.Clouds[name]
	return cloud, ok
}

// DeleteCloud returns a cloud by cloud name
func (cc *Controller) DeleteCloud(name string) {
	delete(cc.Clouds, name)
}

// Provision asks all Clouds in CloudController to provision a worker
// you should compute needed quota by yourself,
func (cc *Controller) Provision(id string, opts WorkerOptions) (Worker, error) {

	if cc.provisionErr == nil {
		cc.provisionErr = NewErrCloudProvision()
	}

	if len(cc.Clouds) == 0 {
		return nil, errors.New("Cloud needs to be registered first")
	}

	if opts.Quota.IsZero() {
		opts.Quota = DefaultQuota.DeepCopy()
	}

	for _, cloud := range cc.Clouds {
		// cloud.Provision will call CanProvision, no need to check twice
		worker, err := cloud.Provision(id, opts)
		if err != nil {
			cc.provisionErr.Add(cloud.Name(), err)
			continue
		}

		logdog.Debug("Provision: success", logdog.Fields{"cloud": cloud.Name(), "cloudKind": cloud.Kind()})

		return worker, nil
	}

	return nil, cc.provisionErr.Err()
}

// LoadWorker ...
func (cc *Controller) LoadWorker(info WorkerInfo) (Worker, error) {
	cloud, ok := cc.GetCloud(info.CloudName)
	if !ok {
		return nil, fmt.Errorf("CloudControler: no cloud named %s registered before", info.CloudName)
	}
	return cloud.LoadWorker(info)
}

// Resources returns all clouds quotas
func (cc *Controller) Resources() (map[string]*Resource, error) {
	resources := make(map[string]*Resource)

	total := NewResource()
	for name, cloud := range cc.Clouds {
		// waitGroup.Add(1)
		res, err := cloud.Resource()
		if err != nil {
			// maybe some clouds are offline, ignore them
			logdog.Error("CloudCtroller: can not get resources from cloud", logdog.Fields{"cloud": cloud.Name(), "err": err})
			continue
		}
		total.Add(res)
		resources[name] = res

	}

	resources["_total"] = total
	return resources, nil
}

func convertCloudToOption(c Cloud) (opt Options) {
	switch c.Type {
	case CloudTypeDocker:
		opt = Options{
			Name:           c.Name,
			Kind:           string(CloudTypeDocker),
			Insecure:       c.Insecure,
			Host:           c.Docker.Host,
			DockerCertPath: c.Docker.CertPath,
		}
	case CloudTypeKubernetes:
		opt = Options{
			Name:           c.Name,
			Kind:           string(CloudTypeKubernetes),
			Insecure:       c.Insecure,
			Host:           c.Kubernetes.Host,
			K8SInCluster:   c.Kubernetes.InCluster,
			K8SBearerToken: c.Kubernetes.BearerToken,
		}
	}

	return opt
}
