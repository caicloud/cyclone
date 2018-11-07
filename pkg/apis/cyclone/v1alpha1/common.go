package v1alpha1

// ResourceItem defines a resource
type ResourceItem struct {
	// Resource name
	Name string `json:"name"`
	// Resource type
	Type ResourceType
}

// ParameterItem defines a parameter
type ParameterItem struct {
	// Name of the parameter
	Name string `json:"name"`
	// Value of the parameter
	Value string `json:"value"`
}

// ArgumentValue defines a argument value
type ArgumentItem ParameterItem
