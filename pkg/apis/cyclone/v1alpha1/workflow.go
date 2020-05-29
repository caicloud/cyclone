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

	// Notification represents the notification config of workflowrun result.
	Notification Notification `json:"notification,omitempty"`

	// GlobalVariables are global variables that can be used in stages or resources parameters. For example, a
	// global variable 'IMAGE_TAG' set here can be used in resource parameters as '${variables.IMAGE_TAG}. Format
	// for the variable reference is ${variables.<variable_name>}
	GlobalVariables []GlobalVariable `json:"globalVariables,omitempty"`
}

// GlobalVariable defines a global variable, For the moment we support three kinds of value:
// - concrete string, for example: 'latest'
// - $(random:<length>), random string with given length, for example: $(random:5)
// - $(timenow:<format>), now time with given time format, for example: $(timenow:RFC1123)
type GlobalVariable struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// NotificationPolicy represents the policy to send notifications.
type NotificationPolicy string

const (
	// NotificationPolicyAlways represents always sending notifications no matter what workflow results are.
	NotificationPolicyAlways NotificationPolicy = "Always"
	// NotificationPolicySuccess represents sending notifications only when workflows succeed.
	NotificationPolicySuccess NotificationPolicy = "Success"
	// NotificationPolicyFailure represents sending notifications only when workflows fail.
	NotificationPolicyFailure NotificationPolicy = "Failure"
)

// Notification represents notifications for workflowrun results.
type Notification struct {
	// Policy represents the policy to send notifications.
	Policy NotificationPolicy `json:"policy"`
	// Receivers represents the receivers of notifications.
	Receivers []NotificationReceiver `json:"receivers"`
}

// NotificationType represents the way to send notifications.
type NotificationType string

const (
	// NotificationTypeEmail represents sending notifications by email.
	NotificationTypeEmail NotificationPolicy = "Email"
	// NotificationTypeSlack represents sending notifications by Slack.
	NotificationTypeSlack NotificationPolicy = "Slack"
	// NotificationTypeWebhook represents sending notifications by webhook.
	NotificationTypeWebhook NotificationPolicy = "Webhook"
)

// NotificationReceiver represents the receiver of notifications.
type NotificationReceiver struct {
	// Type represents the way to send notifications.
	Type NotificationType `json:"type"`
	// Addresses represents the addresses to receive notifications.
	Addresses []string `json:"addresses"`
}

// StageItem describes a stage in a workflow.
type StageItem struct {
	// Name of stage
	Name string `json:"name"`
	// Input artifacts that this stage needed, we bind the artifacts source here.
	Artifacts []ArtifactItem `json:"artifacts"`
	// Stages that this stage depends on.
	Depends []string `json:"depends"`
	// Trivial indicates whether this stage is critical in the workflow. If set to true, it means the workflow
	// can tolerate failure of this stage. In this case, all other stages can continue to execute and the overall
	// status of the workflow execution can still be succeed.
	Trivial bool `json:"trivial"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkflowList describes an array of Workflow instances.
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Workflow `json:"items"`
}
