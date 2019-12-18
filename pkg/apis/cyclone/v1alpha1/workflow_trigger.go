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
	// for GitHub, GitLab, Bitbucket, it is ordinarily in format of 'owner/repo-name',
	// for SVN, it is stored of the RepoUUID of the SVN repo. you can get the SVN repo
	// UUID by command:
	//
	// 'svn info --show-item repos-uuid --username {user} --password {password} --non-interactive
	// --trust-server-cert-failures unknown-ca,cn-mismatch,expired,not-yet-valid,other
	// --no-auth-cache {remote-svn-address}'
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
	// PostCommit represents trigger policy for post commit events.
	PostCommit SCMTriggerPostCommit `json:"postCommit"`
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

	// Branches represents the pr target branches list to filter PullRequest events.
	Branches []string `json:"branches"`
}

// SCMTriggerPullRequestComment represents trigger policy for pull request comment events.
type SCMTriggerPullRequestComment struct {
	SCMTriggerBasic `json:",inline"`
	// Comments represents the comment lists to filter pull request comment events.
	Comments []string `json:"comments"`
}

// SCMTriggerPostCommit represents trigger policy for post commit events.
type SCMTriggerPostCommit struct {
	SCMTriggerBasic `json:",inline"`

	// RootURL represents SVN repository root url, this root is retrieved by
	//
	// 'svn info --show-item repos-root-url --username {user} --password {password} --non-interactive
	// --trust-server-cert-failures unknown-ca,cn-mismatch,expired,not-yet-valid,other
	// --no-auth-cache {remote-svn-address}'
	//
	// e.g: http://192.168.21.97/svn/caicloud
	RootURL string `json:"rootURL"`

	// WorkflowURL represents repository url of the workflow that the wrokflowTrigger related to,
	// Cyclone will checkout code from this URL while executing WorkflowRun.
	// e.g: http://192.168.21.97/svn/caicloud/cyclone
	WorkflowURL string `json:"workflowURL"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkflowTriggerList describes an array of WorkflowTrigger instances.
type WorkflowTriggerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []WorkflowTrigger `json:"items"`
}
