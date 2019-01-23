package v1alpha1

import (
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Project defines a project which holds a lot of common default information of some workflow under it.
type Project struct {
	// Metadata for the resource, like kind and apiversion
	meta_v1.TypeMeta `json:",inline"`
	// Metadata for the particular object, including name, namespace, labels, etc
	meta_v1.ObjectMeta `json:"metadata,omitempty"`
	// Spec is the Workflow specification
	Spec ProjectSpec `json:"spec"`
}

// ProjectSpec defines project specification.
type ProjectSpec struct {
	// Services contains default value of various type of integrations.
	Services []ServiceItem `json:"services"`

	// Quota is the default quota of the workflow under it,
	// eg map[core_v1.ResourceName]string{"requests.cpu": "2", "requests.memory": "4Gi"}
	Quota map[core_v1.ResourceName]string `json:"quota"`
}

// ServiceItem describes default value of a type of integrations
type ServiceItem struct {
	// Type is the integration type
	Type string `json:"type"`
	// Name is the default value of the corresponding type of integration
	Name string `json:"name"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProjectList describes an array of Project instances.
type ProjectList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []Project `json:"items"`
}
