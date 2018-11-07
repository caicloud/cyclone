package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	Stages StageItem `json:"stageItem"`
}

// StageItem describes a stage in a workflow.
type StageItem struct {
	// Name of stage
	Name string `json:"name"`
	// Stages that this stage depends on
	Depends []string `json:"depends"`
}
