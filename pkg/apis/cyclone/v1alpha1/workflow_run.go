package v1alpha1

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkflowRun describes one workflow run, giving concrete runtime parameters and
// recording workflow run status.
type WorkflowRun struct {
	// Metadata for the resource, like kind and apiversion
	metav1.TypeMeta `json:",inline"`
	// Metadata for the particular object, including name, namespace, labels, etc
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Workflow run specification
	Spec WorkflowRunSpec `json:"spec"`
	// Status of workflow execution
	Status WorkflowRunStatus `json:"status"`
}

// WorkflowRunSpec defines workflow run specification.
type WorkflowRunSpec struct {
	// Stages in the workflow to start execution
	StartStages []string `json:"startStages"`
	// Stages in the workflow to end execution
	EndStages []string `json:"endStages"`
	// ServiceAccount used in the workflow execution
	ServiceAccount string `json:"serviceAccount"`
	// Resource parameters
	Resources []ParameterConfig `json:"resources"`
	// Stage parameters
	Stages []ParameterConfig `json:"stages"`
}

// ParameterConfig configures parameters of a resource or a stage.
type ParameterConfig struct {
	// Whose parameters to configure
	Name string `json:"name"`
	// Parameters
	Parameters []ParameterItem `json:"parameters"`
}

// WorkflowRunStatus records workflow running status.
type WorkflowRunStatus struct {
	// Status of all stages
	Stages map[string]StageStatus
	// Overall conditions
	Conditions []Condition
}

// StageStatus describes status of a stage execution.
type StageStatus struct {
	// Conditions of a stage
	Conditions []Condition
	// Key-value outputs of this stage
	Outputs []KeyValue
}

// ConditionType is a camel-cased condition type.
type ConditionType string

// Conditions defines a readiness condition for Cyclone resource.
// See: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#typical-status-properties
// +k8s:deepcopy-gen=true
type Condition struct {
	// Type of condition.
	// +required
	Type ConditionType `json:"type"`

	// Status of the condition, one of True, False, Unknown.
	// +required
	Status corev1.ConditionStatus `json:"status"`

	// LastTransitionTime is the last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// WorkflowRunList describes an array of WorkflowRun instances.
type WorkflowRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []WorkflowRun `json:"items""`
}
