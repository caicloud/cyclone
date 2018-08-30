package types

import (
	"time"
)

type DockerAccount struct {
	Metadata *DockerAccountMetadata `json:"metadata"`
	Spec     *DockerAccountSpec     `json:"spec"`
	Status   *DockerAccountStatus   `json:"status"`
}

type DockerAccountMetadata struct {
	CreationTime time.Time `json:"creationTime"`
}
type DockerAccountSpec struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
}
type DockerAccountStatus struct{}
