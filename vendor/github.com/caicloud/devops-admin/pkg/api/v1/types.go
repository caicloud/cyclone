/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package v1

const (
	// APIVersion is the version of API.
	APIVersion = "/api/v1"
)

// Workspace represents the isolated space for your work.
type Workspace struct {
	ID             string `bson:"_id" json:"-"`
	Name           string `bson:"name" json:"name"`
	Description    string `bson:"description" json:"description"`
	Owner          string `bson:"owner" json:"owner"`
	Tenant         string `bson:"tenant" json:"-"`
	CycloneProject string `bson:"cycloneProject" json:"-"`
	CreationTime   string `bson:"creationTime" json:"creationTime"`
	LastUpdateTime string `bson:"lastUpdateTime" json:"lastUpdateTime"`
}

// ListMeta represents metadata that list resources must have.
type ListMeta struct {
	Total int `json:"total"`
}

// ListResponse represents a collection of some resources.
type ListResponse struct {
	Meta  ListMeta    `json:"metadata"`
	Items interface{} `json:"items"`
}

// ErrorResponse represents response of error.
type ErrorResponse struct {
	Message string `json:"message,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Details string `json:"details,omitempty"`
}