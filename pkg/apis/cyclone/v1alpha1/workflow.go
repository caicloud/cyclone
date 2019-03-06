package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Workflow defines a workflow
type Workflow struct {
	// Metadata for the resource, like kind and apiversion
	metav1.TypeMeta `json:",inline"`
	// Metadata for the particular object, including name, namespace, labels, etc
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Workflow specification
	Spec WorkflowSpec `json:"spec"`
}

// WorkflowSpec defines workflow specification.
type WorkflowSpec struct {
	Resources *corev1.ResourceRequirements
	Stages    []StageItem `json:"stages"`
}

// StageItem describes a stage in a workflow.
type StageItem struct {
	// Name of stage
	Name string `json:"name"`
	// Input artifacts that this stage needed, we bind the artifacts source here.
	Artifacts []ArtifactItem `json:"artifacts"`
	// Stages that this stage depends on
	Depends []string `json:"depends"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkflowList describes an array of Workflow instances.
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Workflow `json:"items"`
}
