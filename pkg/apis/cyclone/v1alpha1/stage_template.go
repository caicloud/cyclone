package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StageTemplate defines a template of Stage.
type StageTemplate struct {
	// Metadata for the resource, like kind and apiversion
	metav1.TypeMeta `json:",inline"`
	// Metadata for the particular object, including name, namespace, labels, etc
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Stage template specification
	Spec StageTemplateSpec `json:"spec"`
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
	Resources []ResourceItem `json:"resource,omitempty"`
	// Artifacts to output
	Artifacts []ArtifactItem `json:"artifacts,omitempty"`
}

// StatusRef defines how to judge CRD status completion.
type StatusRef struct {
	// Path of field in the CRD spec that determines status of CRD workload
	Path string `json:"path"`
	// Value indicates CRD workload completion
	ComplatedValue string `json:"complatedValue"`
}

// CRDSpec defines CRD workload specification.
type CRDSpec struct {
	// Specification of the CRD
	Manifest string `json:"manifest"`
	// How to judge CRD workload is completed
	StatusRef StatusRef `json:"statusRef"`
}

// PodWorkload describes pod type workload, a complete pod spec is included.
type PodWorkload struct {
	// Stage inputs
	Inputs Inputs `json:"inputs,omitempty"`
	// Stage outputs
	Outputs Outputs `json:"outputs,omitempty"`
	// Stage workload specification
	Spec corev1.PodSpec `json:"spec"`
}

// CRDWorkload describes crd type workload.
type CRDWorkload struct {
	// Stage workload specification
	Spec CRDSpec `json:"spec"`
}

// StageTemplateSpec defines stage template specification.
// Exact one workload should be specified.
type StageTemplateSpec struct {
	// Arguments of the template
	Arguments []Argument `json:"arguments,omitempty"`
	// Pod kind workload
	Pod PodWorkload `json:"pod,omitempty"`
	// CRD kind workload
	CRD CRDWorkload `json:"crd,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StageTemplateList describes an array of StageTemplate instances.
type StageTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []StageTemplate `json:"items""`
}
