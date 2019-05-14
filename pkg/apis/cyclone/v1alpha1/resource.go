package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Resource represents a resource used in workflow
type Resource struct {
	// Metadata for the resource, like kind and apiversion
	metav1.TypeMeta `json:",inline"`
	// Metadata for the particular object, including name, namespace, labels, etc
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Resource specification
	Spec ResourceSpec `json:"spec"`
}

// ResourceType defines resource type
type ResourceType string

const (
	// ImageResourceType represents image in docker registry
	ImageResourceType = "Image"
	// GitResourceType represents git repo in SCM
	GitResourceType = "Git"
	// SvnResourceType represents svn repo in SCM
	SvnResourceType = "Svn"
	// GeneralResourceType represents general resource
	GeneralResourceType = "General"
)

// ResourcePullPolicy indicates resource pull policy
type ResourcePullPolicy string

const (
	// PullAlways indicates always pull resource. Old data would be removed if exist.
	PullAlways = "Always"
	// PullIfNotExist performs incremental pull if old data exists.
	PullIfNotExist = "IfNotExist"
)

// ResourceSpec describes a resource
type ResourceSpec struct {
	// Image to resolve this kind of resource.
	Resolver string `json:"resolver,omitempty"`
	// Resource type, e.g. image, git, kv, general.
	Type ResourceType `json:"type"`
	// Persistent resource to PVC.
	Persistent *Persistent `json:"persistent"`
	// Parameters of the resource
	Parameters []ParameterItem `json:"parameters"`
}

// Persistent describes persistent parameters for the resource.
type Persistent struct {
	// Name of the PVC to hold the resource
	PVC string `json:"pvc"`
	// Path of resource in the PVC
	Path string `json:"path"`
	// Whether to pull resource when there already be data
	PullPolicy ResourcePullPolicy `json:"pullPolicy"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResourceList describes an array of Resource instances.
type ResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Resource `json:"items"`
}
