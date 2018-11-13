package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkflowTrigger describes trigger of an workflow, time schedule and webhook supported.
type WorkflowTrigger struct {
	// Metadata for the resource, like kind and apiversion
	metav1.TypeMeta `json:",inline"`
	// Metadata for the particular object, including name, namespace, labels, etc
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of this workflow trigger
	Spec WorkflowTriggerSpec `json:"spec"`
	// Status of this trigger
	Status WorkflowTriggerStatus `json:"status"`
}

// Type of workflow trigger
type TriggerType string

const (
	ScheduledTrigger = "Schedule"
	WebhookTrigger   = "Webhook"
)

// WorkflowTriggerSpec defines workflow trigger definition.
type WorkflowTriggerSpec struct {
	// Reference of the Workflow to trigger
	WorkflowRef *corev1.ObjectReference `json:"workflowRef"`
	// Type of this trigger, Schedule or Webhook
	Type TriggerType `json:"triggerType"`
	// Whether this trigger is enabled, if set to true, no workflow will be triggered
	Enabled bool `json:"enabled"`
	// Parameters of the trigger, for Schedule type trigger, "schedule"
	// parameter is required
	Parameters []ParameterItem `json:"parameters"`
}

// WorkflowTriggerStatus describes status of a workflow trigger
type WorkflowTriggerStatus struct {
	// How many times this trigger got triggered
	Count int `json:"count"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkflowTriggerList describes an array of WorkflowTrigger instances.
type WorkflowTriggerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []WorkflowTrigger `json:"items""`
}
