package v1alpha1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cmd_api "k8s.io/client-go/tools/clientcmd/api"
)

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExecutionCluster defines a cluster to run workflows.
type ExecutionCluster struct {
	// Metadata for the resource, like kind and apiversion
	meta_v1.TypeMeta `json:",inline"`
	// Metadata for the particular object, including name, namespace, labels, etc
	meta_v1.ObjectMeta `json:"metadata,omitempty"`
	// Spec is the Workflow specification
	Spec ExecutionClusterSpec `json:"spec"`
}

// ExecutionClusterSpec defines execution cluster specification.
type ExecutionClusterSpec struct {
	// Credential is the credential info of the cluster
	Credential ClusterCredential `json:"credential"`
}

// ClusterCredential contains credential info about cluster
type ClusterCredential struct {
	// Server represents the address of cluster.
	Server string `json:"server"`
	// User is a user of the cluster.
	User string `json:"user"`
	// Password is the password of the corresponding user.
	Password string `json:"password"`
	// BearerToken is the credential to access cluster.
	BearerToken string `json:"bearerToken"`
	// TLSClientConfig is the config about TLS
	TLSClientConfig *TLSClientConfig `json:"tlsClientConfig,omitempty"`
	// KubeConfig is the config about kube config
	KubeConfig *cmd_api.Config `json:"kubeConfig,omitempty"`
}

// TLSClientConfig contains settings to enable transport layer security
type TLSClientConfig struct {
	// Server should be accessed without verifying the TLS certificate. For testing only.
	Insecure bool `json:"insecure,omitempty" bson:"insecure"`

	// CAFile is the trusted root certificates for server
	CAFile string `json:"caFile,omitempty" bson:"caFile"`

	// CAData holds PEM-encoded bytes (typically read from a root certificates bundle).
	// CAData takes precedence over CAFile
	CAData []byte `json:"caData,omitempty" bson:"caData"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExecutionClusterList describes an array of ExecutionCluster instances.
type ExecutionClusterList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []ExecutionCluster `json:"items"`
}
