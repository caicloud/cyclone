package types

import (
	"time"
)

type Registry struct {
	Metadata *RegistryMetadata `json:"metadata"`
	Spec     *RegistrySpec     `json:"spec"`
	Status   *RegistryStatus   `json:"status"`
}

type RegistryMetadata struct {
	Name           string    `json:"name"`
	Alias          string    `json:"alias"`
	CreationTime   time.Time `json:"creationTime"`
	LastUpdateTime time.Time `json:"lastUpdateTime"`
}

type RegistrySpec struct {
	Host     string `json:"host"`
	Domain   string `json:"domain"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegistryStatus struct {
	ProjectCount    *ProjectCount    `json:"projectCount"`
	RepositoryCount *RepositoryCount `json:"repositoryCount"`
	StorageStatics  *StorageStatics  `json:"storageStatics"`
	Healthy         bool             `json:"healthy"`
}

type ProjectCount struct {
	Public  int64 `json:"public"`
	Private int64 `json:"private"`
}

type RepositoryCount struct {
	Public  int64 `json:"public"`
	Private int64 `json:"private"`
}

type StorageStatics struct {
	Used  string `json:"used"`
	Total string `json:"total"`
}

// =================================================================================================

type CreateRegistryReq struct {
	Metadata *RegistryMetadata `json:"metadata"`
	Spec     *RegistrySpec     `json:"spec"`
}

// =================================================================================================

type UpdateRegistryReq struct {
	Metadata *RegistryMetadata `json:"metadata"`
	Spec     *RegistrySpec     `json:"spec"`
}
