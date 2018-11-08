package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkflowParam describes global workflow runtime parameters
type WorkflowParam struct {
	// Metadata for the resource, like kind and apiversion
	metav1.TypeMeta `json:",inline"`
	// Metadata for the particular object, including name, namespace, labels, etc
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Workflow param specification
	Spec WorkflowParamSpec `json:"spec"`
}

// WorkflowParamSpec describes global parameters a workflow should run with.
type WorkflowParamSpec struct {
	// PVC to be used for workflow
	PVC string `json:"pvc"`
	// Cluster where the workflow will run on
	Cluster string `json:"cluster"`
}

// WorkflowParamList describes an array of WorkflowParam instances.
type WorkflowParamList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []WorkflowParam `json:"items""`
}
