package types

import (
	"time"
)

type Repository struct {
	Metadata *RepositoryMetadata `json:"metadata"`
	Spec     *RepositorySpec     `json:"spec"`
	Status   *RepositoryStatus   `json:"status"`
}

type RepositoryMetadata struct {
	Name           string    `json:"name"`
	CreationTime   time.Time `json:"creationTime"`
	LastUpdateTime time.Time `json:"lastUpdateTime"`
}

type RepositorySpec struct {
	Project     string `json:"project"`
	FullName    string `json:"fullName"`
	Description string `json:"description"`
}

type RepositoryStatus struct {
	TagCount  int64 `json:"tagCount"`
	PullCount int64 `json:"pullCount"`
}

type UpdateRepositoryReq struct {
	Spec *RepositorySpec `json:"spec"`
}
