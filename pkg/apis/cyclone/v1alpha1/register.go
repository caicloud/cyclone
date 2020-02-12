package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupName is the group name use in this package
const GroupName = "cyclone.dev"

// Version is version of the CRD
const Version = "v1alpha1"

// APIVersion ...
const APIVersion = GroupName + "/" + Version

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

var (
	// SchemeBuilder ...
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	// AddToScheme ...
	AddToScheme = SchemeBuilder.AddToScheme
)

// GroupResource ...
func GroupResource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// Adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Resource{},
		&ResourceList{},
		&Workflow{},
		&WorkflowList{},
		&WorkflowRun{},
		&WorkflowRunList{},
		&Stage{},
		&StageList{},
		&WorkflowTrigger{},
		&WorkflowTriggerList{},
		&Project{},
		&ProjectList{},
		&ExecutionCluster{},
		&ExecutionClusterList{},
	)
	// Add the watch version that applies
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
