package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// DefaultCrdResourceGroup is the group name use in this package
	DefaultCrdResourceGroup = "resource.caicloud.io"
	// DefaultAPIVersion is default api version of this package
	DefaultAPIVersion = "v1beta1"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: DefaultCrdResourceGroup, Version: DefaultAPIVersion}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	// SchemeBuilder is builder of scheme ,We only register manually written functions here.
	// The registration of the generated functions takes place in the generated files. The
	// separation makes the code compile even when the generated files are missing.
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
		&Cluster{},
		&ClusterList{},
	)

	// Add common types
	scheme.AddKnownTypes(SchemeGroupVersion, &metav1.Status{})

	// Add the watch version that applies
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
