/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package models

// Link describes a self-link for resource
type Link struct {
	// Name is name of the resource
	Name string `json:"name"`
	// Namespace is namespace of the resource. Optional
	Namespace string `json:"namespace"`
	// SelfLink is uri of the resource
	SelfLink string `json:"selfLink"`
}
