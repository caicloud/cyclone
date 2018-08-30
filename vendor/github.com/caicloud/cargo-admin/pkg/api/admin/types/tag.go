package types

import (
	"time"
)

type Tag struct {
	Metadata *TagMetadata `json:"metadata"`
	Spec     *TagSpec     `json:"spec"`
	Status   *TagStatus   `json:"status"`
}

type TagMetadata struct {
	Name         string    `json:"name"`
	CreationTime time.Time `json:"creationTime"`
}

type TagSpec struct {
	Image string `json:"image"`
}

type TagStatus struct {
	Author               string           `json:"author"`
	VulnerabilitiesCount int              `json:"vulnerabilitiesCount"`
	Vulnerabilities      []*Vulnerability `json:"vulnerabilities"`
}

type Vulnerability struct {
	Name        string `json:"name"`
	Package     string `json:"package"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Severity    string `json:"severity"`
	Version     string `json:"version"`
	Fixed       string `json:"fixedVersion,omitempty"`
}
