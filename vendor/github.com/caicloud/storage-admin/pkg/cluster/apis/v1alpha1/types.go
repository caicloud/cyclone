package v1alpha1

import (
	"time"
)

type ProviderType string

const (
	Baremetal ProviderType = "caicloud-baremetal"
	Anchnet   ProviderType = "caicloud-anchnet"
	Ansible   ProviderType = "caicloud-ansible"
	Aliyun    ProviderType = "caicloud-aliyun"
	Fake      ProviderType = "caicloud-fake"
)

type ClusterStatus string

const (
	ClusterPreparing ClusterStatus = "preparing"
	ClusterRunning   ClusterStatus = "running"
	ClusterPausing   ClusterStatus = "pausing"
	ClusterPaused    ClusterStatus = "paused"
	ClusterResuming  ClusterStatus = "resuming"
	ClusterScalingUp ClusterStatus = "scaling_up"
	ClusterDeleted   ClusterStatus = "deleted"
	ClusterDeleting  ClusterStatus = "deleting"
	ClusterFailed    ClusterStatus = "failed"
)

type NodeStatus string

const (
	NodeRunning NodeStatus = "running"
	NodeFailed  NodeStatus = "failed"
	NodeDeleted NodeStatus = "deleted"
)

type ResourceStatus string

const (
	ResourceProvided     ResourceStatus = "provided"
	ResourceBeingDeleted ResourceStatus = "deleting"
	ResourceDeleted      ResourceStatus = "deleted"
)

//Delete clusters if suspended more than 2 days
const SuspendedTimeLimit time.Duration = 48 * time.Hour

// ClusterInfo is the type that will eventually be saved to database. It contains all information
// of a cluster, including its configuration, historical configuration, access information, cluster
// status etc.
type ClusterInfo struct {
	// A unique id of the cluster. Supposed to be the same as the requestId of ClusterCreationRequest.
	ClusterId string `bson:"_id,omitempty" json:"_id,omitempty"`
	// Indicate whether this cluster is extraneous.
	Adopted bool `bson:"adopted,omitempty" json:"adopted,omitempty"`
	// Indicate whether this cluster uses a self signed cert
	SelfSigned bool `bson:"self_signed,omitempty" json:"self_signed,omitempty"`
	// The deployed cluster configuration.
	Config ClusterConfig `bson:"config,omitempty" json:"config,omitempty"`
	// The history configuration information for a cluster. This is used for accounting. A user may
	// resize the cluster later. Note: New config will be pushed to the end of the array so current
	// cluster config will always be the last one in the array.
	ConfigHistory []ClusterConfig `bson:"config_history,omitempty" json:"config_history,omitempty"`
	// The user who own the cluster.
	Uid string `bson:"uid,omitempty" json:"uid,omitempty"`
	// The status of the cluster.
	Status ClusterStatus `bson:"status" json:"status"`
	// The status of the resource allocated to the cluster.
	CurResourceStatus ResourceStatus `bson:"cur_resource_status" json:"cur_resource_status"`
	// Name of the key that is used by the cluster.
	CloudKey string `bson:"cloud_key,omitempty" json:"cloud_key,omitempty"`
	// The information specific to the K8s installed on the cluster.
	K8s K8sInfo `bson:"k8s_info,omitempty" json:"k8s_info,omitempty"`
	// The list of operation ids performed on the cluster.
	// E.g. Create->update->update->Delete.
	OperationIds []string `bson:"op_ids,omitempty" json:"op_ids,omitempty"`
	// The timestamp when the record is created.
	CreateTime time.Time `bson:"create_time,omitempty" json:"create_time,omitempty"`
	// The timestamp when the cluster is deleted.
	DeleteTime time.Time `bson:"delete_time,omitempty" json:"delete_time,omitempty"`
	// Token for the cluster to access caicloud system.
	ClusterToken string `bson:"cluster_token,omitempty" json:"cluster_token,omitempty"`
	// Flag for shared cluster
	Ng bool `bson:"ng,omitempty" json:"ng,omitempty"`
	// ID of the group, indicating ACL policy
	Gid string `bson:"gid,omitempty" json:"gid,omitempty"`
	// The Id of the secret that is associated with this cluster. Currently this is only used
	// in aliyun.
	SecretId string `bson:"secret_id,omitempty" json:"secret_id,omitempty"`
	// Flag for control cluster
	ControlCluster bool `bson:"control_cluster,omitempty" json:"control_cluster,omitempty"`
	// The cpu overcommit ratio of the overall cluster. This will restrict the sum of cpu limit of all namespaces
	CpuOvercommitRatio float64 `bson:"cpu_overcommit_ratio,omitempty" json:"cpu_overcommit_ratio,omitempty"`
	// The memory overcommit ratio of the overall cluster. This will restrict the sum of memory limit of all namespaces
	MemoryOvercommitRatio float64 `bson:"memory_overcommit_ratio,omitempty" json:"memory_overcommit_ratio,omitempty"`
}

type K8sInfo struct {
	// The version of caicloud kubernetes release used for the cluster.
	CaicloudVersion string `bson:"caicloud_version,omitempty" json:"caicloud_version,omitempty"`
	// The useranme to access the K8 apiserver.
	User string `bson:"user,omitempty" json:"user,omitempty"`
	// The password to access the K8 apiserver.
	Password string `bson:"password,omitempty" json:"password,omitempty"`
	// Base domain name(caicloudapp.com) and host name(epic-caicloud-2015-cluster) will be
	// stored in different fields. While it is easy to get the full domain name by concatenating
	// these two fields, it is a bit tricky to originate each part from the full domain name.
	// The DNS host name of apiserver.
	HostName string `bson:"host_name,omitempty" json:"host_name,omitempty"`
	// The base domain name of apiserver.
	BaseDomainName string `bson:"base_domain_name,omitempty" json:"base_domain_name,omitempty"`
	// The apiserver endpoint ip.
	// If there are more than one master, a load balancer ip is used.
	EndpointIp string `bson:"endpoint_ip,omitempty" json:"endpoint_ip,omitempty"`
	// The apiserver endpoint port.
	EndpointPort string `bson:"endpoint_port,omitempty" json:"endpoint_port,omitempty"`
}

// ClusterConfig is configuration of a cluster, including its name, configuration of individual
// machine, etc.
type ClusterConfig struct {
	// The user provided name of the cluster.
	ClusterName string `bson:"cluster_name,omitempty" json:"cluster_name,omitempty"`
	// The provider of cloud resources. E.g. anchnet.
	Provider ProviderType `bson:"provider,omitempty" json:"provider,omitempty"`
	// The configuration of each master in the cluster
	Masters []MachineConfig `bson:"master_config" json:"master_config"`
	// The configuration of each node in the cluster
	Nodes []MachineConfig `bson:"node_config" json:"node_config"`
	// The timestamp when this configuration is used.
	Timestamp time.Time `bson:"timestamp,omitempty" json:"timestamp,omitempty"`
	// Whether the current config is chargable to Caicloud. A config is not chargable because
	// the cluster is suspended or other reasons like self-hosting cluster.
	Chargable bool `bson:"chargable,omitempty" json:"chargable,omitempty"`
}

// MachineConfig is the configuration of a machine (instance). Some of the field might
// be empty. e.g. In case we are bringing up cluster on private cloud, Mem, Cpu and
// Bandwidth are left empty.
type MachineConfig struct {
	// Id of a node.
	Id string `bson:"_id,omitempty" json:"_id,omitempty"`
	// Status of the node
	Status NodeStatus `bson:"status" json:"status"`
	// The memory size of each instance in the cluster (in MB).
	Mem int `bson:"mem,omitempty" json:"mem,omitempty"`
	// Number of cores in each instance.
	Cpu int `bson:"cpu,omitempty" json:"cpu,omitempty"`
	// The storage size of each instance in the cluster (in Bytes).
	Storage uint64 `bson:"storage,omitempty" json:"storage,omitempty"`
	// Network bandwith in MB/s.
	Bandwidth int `bson:"bw,omitempty" json:"bw,omitempty"`
	// Ip of this node, this can be either passed in by user or reported by kube-up.
	Ip string `bson:"ip,omitempty" json:"ip,omitempty"`
	// Internal ip of this node. Keep this the same with Ip if the machine only have external ip.
	InternalIp string `bson:"internal_ip,omitempty" json:"internal_ip,omitempty"`
	// ssh info in the form of username:password
	SSHInfo string `bson:"ssh_info,omitempty" json:"ssh_info,omitempty"`
	// sudoer's info in the form of username:password. This is used when bringing up a cluster.
	SudoInfo string `bson:"sudo_info,omitempty"`
	// k8s instance name
	InstanceName string `bson:"instance_name,omitempty" json:"instance_name,omitempty"`
	// The timestamp when this node is created
	CreateTime time.Time `bson:"create_time,omitempty" json:"create_time,omitempty"`
}

// Response type for cluster list request.
type ClusterListResponse struct {
	// The token used to poll the cluster status from UI IFF request is accepted.
	Clusters []ClusterInfo `json:"clusters,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors,
	ErrorMessage string `json:"error_msg,omitempty"`
}

// Response type for cluster info request.
type ClusterInfoResponse struct {
	// The token used to poll the cluster status from UI if request is accepted.
	Cluster ClusterInfo `json:"cluster,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors,
	ErrorMessage string `json:"error_msg,omitempty"`
}

type ClusterAuthResponse struct {
	ErrorMessage string `json:"error_msg,omitempty"`
}

type OvercommitRatio struct {
	CpuOvercommitRatio    float64 `json:"cpu_overcommit_ratio,omitempty"`
	MemoryOvercommitRatio float64 `json:"memory_overcommit_ratio,omitempty"`
}

type ClusterGetOvercommitRatioResponse struct {
	CpuOvercommitRatio    float64 `json:"cpu_overcommit_ratio,omitempty"`
	MemoryOvercommitRatio float64 `json:"memory_overcommit_ratio,omitempty"`
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}

type ClusterSetOvercommitRatioResponse struct {
	// Return the error message IFF not successful. This is used to provide user-facing errors.
	ErrorMessage string `json:"error_msg,omitempty"`
}
