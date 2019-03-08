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
	// TriggerTypeCron indicates cron trigger
	TriggerTypeCron TriggerType = "Cron"

	// TriggerTypeWebhook indicates webhook trigger
	WebhookTrigger TriggerType = "Webhook"
)

// WorkflowTriggerSpec defines workflow trigger definition.
type WorkflowTriggerSpec struct {
	// Type of this trigger, Cron or Webhook
	Type TriggerType `json:"triggerType"`
	// Parameters of the trigger to run workflow
	Parameters []ParameterItem `json:"parameters"`
	// CronTrigger represents cron trigger config.
	Cron CronTrigger `json:"cron,omitempty"`
	// Whether this trigger is disabled, if set to true, no workflow will be triggered
	Disabled bool `json:"disabled"`
	// Spec to run the workflow
	WorkflowRunSpec `json:",inline"`
}

// CronTrigger represents the cron trigger policy.
type CronTrigger struct {
	Schedule string `json:schedule`
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
