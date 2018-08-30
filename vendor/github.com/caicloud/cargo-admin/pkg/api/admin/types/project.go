package types

import (
	"time"
)

type ProjectMetadata struct {
	Name           string    `json:"name"`
	CreationTime   time.Time `json:"creationTime"`
	LastUpdateTime time.Time `json:"lastUpdateTime"`
}

type ProjectSpec struct {
	IsPublic            bool      `json:"isPublic"`
	IsProtected         bool      `json:isProtected`
	Registry            string    `json:"registry"`
	Description         string    `json:"description"`
	LastImageUpdateTime time.Time `json:"lastImageUpdateTime"`
}

type ProjectStatus struct {
	Synced           bool `json:synced`
	RepositoryCount  int  `json:"repositoryCount"`
	ReplicationCount int  `json:"replicationCount"`
}

type Project struct {
	Metadata *ProjectMetadata `json:"metadata"`
	Spec     *ProjectSpec     `json:"spec"`
	Status   *ProjectStatus   `json:"status"`
}

// =================================================================================================

type CreateProjectReq struct {
	Metadata *ProjectMetadata `json:"metadata"`
	Spec     *ProjectSpec     `json:"spec"`
}

// =================================================================================================

type UpdateProjectReq struct {
	Spec *ProjectSpec `json:"spec"`
}
