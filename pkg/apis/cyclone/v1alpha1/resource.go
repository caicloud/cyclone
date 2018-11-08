package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Resource represents a resource used in workflow
type CycloneResource struct {
	// Metadata for the resource, like kind and apiversion
	metav1.TypeMeta `json:",inline"`
	// Metadata for the particular object, including name, namespace, labels, etc
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Resource specification
	Spec ResourceSpec `json:"spec"`
}

// Resource type
type ResourceType string

const (
	ImageResourceType   = "Image"
	GitResourceType     = "Git"
	KVResourceType      = "KV"
	PVResourceType      = "PV"
	GeneralResourceType = "General"
)

// Resource pull policy
type ResourcePullPolicy string

const (
	// Always pull resource. Old data would be removed if exist.
	PullAlways = "Always"
	// If old data exists, take advantage of it when pull resource,
	// incremental pull is performed.
	PullIfNotExist = "IfNotExist"
)

// ResourceSpec describes a resource
type ResourceSpec struct {
	// Image to resolve this kind of resource.
	Resolver string `json:"resolver,omitempty"`
	// Resource type, e.g. image, git, kv, general.
	Type ResourceType `json:"type"`
	// Whether persist resource to a dedicated storage (PV).
	Persistent bool `json:"persistent"`
	// Whether to pull resource when there already be data
	PullPolicy ResourcePullPolicy `json:"pullPolicy"`
	// Parameters of the resource
	Parameters []ParameterItem `json:"parameters"`
}

// CargoList describes an array of CycloneResource instances.
type CycloneResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []CycloneResource `json:"items""`
}
