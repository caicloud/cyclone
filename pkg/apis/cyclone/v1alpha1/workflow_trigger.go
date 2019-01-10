package v1alpha1

import (
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

// TriggerType defines type of workflow trigger
type TriggerType string

const (
	// ScheduledTrigger indicates scheduled trigger
	ScheduledTrigger = "Schedule"
	// WebhookTrigger indicates webhook trigger
	WebhookTrigger = "Webhook"
)

// WorkflowTriggerSpec defines workflow trigger definition.
type WorkflowTriggerSpec struct {
	// Type of this trigger, Schedule or Webhook
	Type TriggerType `json:"triggerType"`
	// Parameters of the trigger, for Schedule type trigger, "schedule"
	// parameter is required
	Parameters []ParameterItem `json:"parameters"`
	// Whether this trigger is disabled, if set to true, no workflow will be triggered
	Disabled bool `json:"disabled"`
	// Spec to run the workflow
	WorkflowRunSpec `json:",inline"`
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
	Items           []WorkflowTrigger `json:"items"`
}
