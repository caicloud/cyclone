package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	// ImageResourceType represents image in docker registry
	ImageResourceType = "Image"
	// GitResourceType represents git repo in SCM
	GitResourceType = "Git"
	// SvnResourceType represents svn repo in SCM
	SvnResourceType = "Svn"
	// HTTPResourceType represents operating resources by http protocol
	HTTPResourceType = "Http"
)

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
	Type string `json:"type"`
	// Persistent resource to PVC.
	Persistent *Persistent `json:"persistent"`
	// Parameters of the resource
	Parameters []ParameterItem `json:"parameters"`
	// SupportedOperations defines what operations the resource type supported,
	// it's only used to register a resource type. When you create a resource for
	// workflow, just ignore it.
	SupportedOperations []string `json:"operations"`
	// IntegrationBind binds the resource type to integration (represent a external data source).
	// It's used when define a resource type.
	IntegrationBind *IntegrationBind `json:"bind,omitempty"`
}

// IntegrationBind describes bindings between a resource type and a integration type.
type IntegrationBind struct {
	// IntegrationType is type of integration to bind for this resource type, for example 'DockerRegistry'
	IntegrationType string `json:"integrationType"`
	// ParamBindings binds parameters of one resource type to the integration. It's a map with keys being
	// parameter names of the resource type, and values being the parameter names of the integration type.
	ParamBindings map[string]string `json:"paramBindings"`
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
