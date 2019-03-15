package cd

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	// ConfigEnvKey is environment variable key for config
	ConfigEnvKey = "_CONFIG_"
	// DeploymentTypeDeployment indicate deployment type 'deployment'
	DeploymentTypeDeployment = "deployment"
)

// Config configures for a CD stage
type Config struct {
	Cluster    *ClusterInfo    `json:"cluster"`
	Deployment *DeploymentInfo `json:"deployment"`
	Images     []*ImageInfo    `json:"images"`
}

// ClusterInfo describes cluster information.
type ClusterInfo struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// DeploymentInfo describes deployment information, for the moment, only 'deployment' type supported.
type DeploymentInfo struct {
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	Name      string `json:"name"`
}

// ImageInfo describes which image to update
type ImageInfo struct {
	Container string `json:"container"`
	Image     string `json:"image"`
}

// LoadConfig loads CD configs from environment variable.
func LoadConfig() (*Config, error) {
	value := os.Getenv(ConfigEnvKey)
	if len(value) == 0 {
		return nil, fmt.Errorf("no config found from environment variable '%s'", ConfigEnvKey)
	}

	config := &Config{}
	err := json.Unmarshal([]byte(value), config)
	if err != nil {
		return nil, fmt.Errorf("unmarshal config error: %v", err)
	}

	return config, nil
}
