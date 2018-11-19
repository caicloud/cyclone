package v1alpha1

// ResourceItem defines a resource
type ResourceItem struct {
	// Resource name
	Name string `json:"name"`
	// Path that this resource should be mounted in container.
	Path string `json:"path"`
}

// ArtifactItem defines an artifact
type ArtifactItem struct {
	// Artifact name
	Name string `json:"name"`
	// Path of the artifact
	Path string `json:"path"`
	// Optional, used for output artifacts to specify which container containers the artifact.
	Container string `json:"container"`
}

// ParameterItem defines a parameter
type ParameterItem struct {
	// Name of the parameter
	Name string `json:"name"`
	// Value of the parameter
	Value string `json:"value"`
}

// ArgumentValue defines a argument value
type ArgumentValue ParameterItem

// KeyValue defines a key-value pair
type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
