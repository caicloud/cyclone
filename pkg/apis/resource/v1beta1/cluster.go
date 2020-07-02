package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Cluster describes a cluster with kubernetes and addon
//
// Cluster are non-namespaced; the id of the cluster
// according to etcd is in ObjectMeta.Name.
type Cluster struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines a specification of a cluster.
	// Provisioned by an administrator.
	Spec ClusterSpec `json:"spec"`

	// Status represents the current information/status for the cluster.
	// Populated by the system.
	Status ClusterStatus `json:"status"`
}

// ClusterSpec is a description of a cluster.
type ClusterSpec struct {
	// DisplayName is the human-readable name for cluster.
	DisplayName string `json:"displayName"`

	IsControlCluster bool   `json:"isControlCluster"`
	IsHighAvailable  bool   `json:"isHighAvailable"`
	MastersVIP       string `json:"mastersVIP"`

	// MastersEIP is the access endpoint of baremetal cluster, which can been access outof vpc
	MastersEIP string `json:"mastersEIP"`

	Masters []string `json:"masters"`
	Nodes   []string `json:"nodes"`
	Etcds   []string `json:"etcds"`
}

// ClusterStatus represents information about the status of a cluster.
type ClusterStatus struct{}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterList is a collection of clusters
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of Clusters
	Items []Cluster `json:"items" protobuf:"bytes,2,rep,name=items"`
}
