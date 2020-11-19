package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Stage defines a workflow stage. Only one of Spec and Template should be specified.
type Stage struct {
	// Metadata for the resource, like kind and apiversion
	metav1.TypeMeta `json:",inline"`
	// Metadata for the particular object, including name, namespace, labels, etc
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Stage specification
	Spec StageSpec `json:"spec,omitempty"`
}

// StageSpec defines stage specification.
// Exact one workload should be specified.
type StageSpec struct {
	// Pod kind workload
	Pod *PodWorkload `json:"pod,omitempty"`
	// Delegation kind workload, this stage would be executed externally.
	Delegation *DelegationWorkload `json:"delegation,omitempty"`
}

// PodWorkload describes pod type workload, a complete pod spec is included.
type PodWorkload struct {
	// Stage inputs
	Inputs Inputs `json:"inputs,omitempty"`
	// Stage outputs
	Outputs Outputs `json:"outputs,omitempty"`
	// Stage workload specification
	Spec corev1.PodSpec `json:"spec"`
	// Stage workload metadata
	Meta *PodWorkloadMeta `json:"metadata,omitempty"`
}

// PodWorkloadMeta describes extra labels or annotations that should be added to the PodWorkload.
type PodWorkloadMeta struct {
	// Labels is a map of string keys and values that can be used to organize and categorize
	// (scope and select) objects.
	Labels map[string]string `json:"labels,omitempty"`
	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	Annotations map[string]string `json:"annotations,omitempty"`
}

// DelegationWorkload describes workload delegated to external services.
type DelegationWorkload struct {
	// Type identifies what kind of workload this is, for example 'notification', Cyclone doesn't need to understand it.
	Type string `json:"type"`
	// URL of the target service. Cyclone would send POST request to this URL.
	URL string `json:"url"`
	// Config is a json string that configure how to run this workload, it's interpreted by external services.
	Config string `json:"config"`
}

// Argument defines a argument.
type Argument struct {
	Name    string `json:"name"`
	Default string `json:"default"`
}

// Inputs defines stage inputs.
type Inputs struct {
	// Resources used as input
	Resources []ResourceItem `json:"resources,omitempty"`
	// Parameters used as input
	Arguments []ArgumentValue `json:"arguments,omitempty"`
	// Artifacts to output
	Artifacts []ArtifactItem `json:"artifacts,omitempty"`
}

// Outputs defines stage output.
type Outputs struct {
	// Resources used as output
	Resources []ResourceItem `json:"resources,omitempty"`
	// Artifacts to output
	Artifacts []ArtifactItem `json:"artifacts,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StageList describes an array of Stage instances.
type StageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Stage `json:"items"`
}
