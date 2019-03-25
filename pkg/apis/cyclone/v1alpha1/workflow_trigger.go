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

	// TriggerTypeSCM indicates SCM trigger
	TriggerTypeSCM TriggerType = "SCM"

	// TriggerTypeWebhook indicates webhook trigger
	TriggerTypeWebhook TriggerType = "Webhook"
)

// WorkflowTriggerSpec defines workflow trigger definition.
type WorkflowTriggerSpec struct {
	// Type of this trigger, Cron or Webhook
	Type TriggerType `json:"type"`
	// Parameters of the trigger to run workflow
	Parameters []ParameterItem `json:"parameters"`
	// Cron represents cron trigger config.
	Cron CronTrigger `json:"cron,omitempty"`
	// SCM represents webhook trigger config.
	SCM SCMTrigger `json:"scm,omitempty"`
	// Whether this trigger is disabled, if set to true, no workflow will be triggered
	Disabled bool `json:"disabled"`
	// Spec to run the workflow
	WorkflowRunSpec `json:",inline"`
}

// CronTrigger represents the cron trigger policy.
type CronTrigger struct {
	Schedule string `json:"schedule"`
}

// WorkflowTriggerStatus describes status of a workflow trigger.
type WorkflowTriggerStatus struct {
	// Count represents triggered times.
	Count int `json:"count"`
}

// SCMTrigger represents the SCM trigger policy.
type SCMTrigger struct {
	// Secret represents the secret of integrated SCM.
	Secret string `json:"secret"`
	// Repo represents full repo name without server address.
	Repo string `json:"repo"`
	// SCMTriggerPolicy represents trigger policies for SCM events.
	SCMTriggerPolicy `json:",inline"`
}

// SCMTriggerPolicy represents trigger policies for SCM events.
// Supports 4 events: push, tag release, pull request and pull request comment.
type SCMTriggerPolicy struct {
	// Push represents trigger policy for push events.
	Push SCMTriggerPush `json:"push"`
	// TagRelease represents trigger policy for tag release events.
	TagRelease SCMTriggerTagRelease `json:"tagRelease"`
	// PullRequest represents trigger policy for pull request events.
	PullRequest SCMTriggerPullRequest `json:"pullRequest"`
	// PullRequestComment represents trigger policy for pull request comment events.
	PullRequestComment SCMTriggerPullRequestComment `json:"pullRequestComment"`
}

// SCMTriggerBasic represents basic config for SCM trigger policy.
type SCMTriggerBasic struct {
	// Enabled represents whether enable this policy.
	Enabled bool `json:"enabled"`
}

// SCMTriggerTagRelease represents trigger policy for tag release events.
type SCMTriggerTagRelease struct {
	SCMTriggerBasic `json:",inline"`
}

// SCMTriggerPush represents trigger policy for push events.
type SCMTriggerPush struct {
	SCMTriggerBasic `json:",inline"`
	// Branches represents the branch lists to filter push events.
	Branches []string `json:"branches"`
}

// SCMTriggerPullRequest represents trigger policy for pull request events.
type SCMTriggerPullRequest struct {
	SCMTriggerBasic `json:",inline"`
}

// SCMTriggerPullRequestComment represents trigger policy for pull request comment events.
type SCMTriggerPullRequestComment struct {
	SCMTriggerBasic `json:",inline"`
	// Comments represents the comment lists to filter pull request comment events.
	Comments []string `json:"comments"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkflowTriggerList describes an array of WorkflowTrigger instances.
type WorkflowTriggerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []WorkflowTrigger `json:"items"`
}
