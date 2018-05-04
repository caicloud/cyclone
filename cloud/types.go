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

import "time"

var (
	// DefaultCloudPingTimeout ...
	DefaultCloudPingTimeout = time.Duration(5 * time.Second)
)

// CloudProvider defines a cloud provider which provision workers for cyclone
type CloudProvider interface {

	// Name returns cloud name
	Name() string

	// Kind returns cloud type, such as kubernetes
	Kind() string

	// Ping returns nil if cloud is accessible
	Ping() error

	// Resource returns the limit and used quotas of the cloud
	Resource() (*Resource, error)

	// CanProvision returns true if the cloud can provision a worker meetting the quota
	CanProvision(need Quota) (bool, error)

	// Provision returns a worker if the cloud can provison,
	// but worker is not running. you should call worker.Do() to do the work
	// Provision should call ConProvision firstly
	Provision(id string, opts WorkerOptions) (Worker, error)

	// LoadWorker rebuilds a worker from worker info
	LoadWorker(WorkerInfo) (Worker, error)

	//
	GetCloud() Cloud
}

// CloudType represents cloud type, supports Docker and Kubernetes.
type CloudType string

const (
	// CloudTypeDocker represents the Docker cloud type.
	CloudTypeDocker CloudType = "Docker"

	// CloudTypeKubernetes represents the Kubernetes cloud type.
	CloudTypeKubernetes CloudType = "Kubernetes"
)

type CloudDocker struct {
	Host     string `json:"host,omitempty" bson:"host,omitempty"`
	CertPath string `json:"certPath,omitempty" bson:"certPath,omitempty"`
}

type CloudKubernetes struct {
	Host        string `json:"host,omitempty" bson:"host,omitempty"`
	InCluster   bool   `json:"inCluster,omitempty" bson:"inCluster,omitempty"`
	BearerToken string `json:"bearerToken,omitempty" bson:"bearerToken,omitempty"`
}

// Cloud represents clouds for workers.
// TODO (robin) Move this to pkg/api
type Cloud struct {
	ID         string           `bson:"_id,omitempty" json:"id,omitempty"`
	Type       CloudType        `bson:"type,omitempty" json:"type,omitempty"`
	Name       string           `json:"name,omitempty" bson:"name,omitempty"`
	Insecure   bool             `json:"insecure,omitempty" bson:"insecure,omitempty"`
	Docker     *CloudDocker     `json:"docker,omitempty" bson:"docker,omitempty"`
	Kubernetes *CloudKubernetes `json:"kubernetes,omitempty" bson:"kubernetes,omitempty"`
}

// Options is options for all kinds of clouds
type Options struct {
	// mongodb id
	ID string `bson:"_id,omitempty" json:"-"`

	// common options
	Kind     string `json:"kind,omitempty" bson:"kind,omitempty"`
	Name     string `json:"name,omitempty" bson:"name,omitempty"`
	Host     string `json:"host,omitempty" bson:"host,omitempty"`
	Insecure bool   `json:"insecure,omitempty" bson:"insecure,omitempty"`

	// docker cloud
	DockerCertPath string `json:"dockerCertPath,omitempty" bson:"dockerCertPath,omitempty"`

	// k8s cloud
	K8SInCluster   bool   `json:"k8sInCluster,omitempty" bson:"k8sInCluster,omitempty"`
	K8SNamespace   string `json:"k8sNamespace,omitempty" bson:"k8sNamespace,omitempty"`
	K8SBearerToken string `json:"k8sBearerToken,omitempty" bson:"k8sBearerToken,omitempty"`
}

// Worker is the truly excutor to deal with build jobs
type Worker interface {
	// Do starts the worker and does the work
	Do() error

	// GetWorkerInfo returns worker's infomation
	GetWorkerInfo() WorkerInfo

	// IsTimeout returns true if worker is timeout
	// and returns the time left until it is due
	IsTimeout() (bool, time.Duration)

	// Terminate terminates the worker and destroy it
	Terminate() error
}
