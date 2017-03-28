package cloud

import "time"

var (
	// DefaultCloudPingTimeout ...
	DefaultCloudPingTimeout = time.Duration(5 * time.Second)
)

// Cloud defines a cloud provider which provision workers for cyclone
type Cloud interface {

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
	GetOptions() Options
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

// // Event ...
// type Event struct {
// 	// ID ...
// 	ID     string
// 	Quota  Quota
// 	Worker WorkerInfo
// }
