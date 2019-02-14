package config

import (
	"encoding/json"
	"fmt"

	"github.com/caicloud/nirvana/log"
	core_v1 "k8s.io/api/core/v1"
)

const (
	// ConfigFileKey is key of config file in ConfigMap
	ConfigFileKey = "cyclone-server.json"
)

// CycloneServerConfig configures Cyclone Server
type CycloneServerConfig struct {
	// Logging configuration, such as log level.
	Logging LoggingConfig `json:"logging"`

	// CycloneServerHost represents the host for cyclone server to serve on
	CycloneServerHost string `json:"cyclone_server_host"`
	// CycloneServerPort represents the port for cyclone server to serve on
	CycloneServerPort uint16 `json:"cyclone_server_port"`

	// DefaultPVCConfig represents the config of pvc for default tenant
	DefaultPVCConfig PVCConfig `json:"default_pvc_config"`

	// WorkerNamespaceQuota describes the resource quota of the namespace which will be used to run workflows,
	// eg map[core_v1.ResourceName]string{"cpu": "2", "memory": "4Gi"}
	WorkerNamespaceQuota map[core_v1.ResourceName]string `json:"worker_namespace_quota"`
}

// PVCConfig contains the PVC information
type PVCConfig struct {
	// StorageClass represents the strorageclass used to create pvc
	StorageClass string `json:"storage_class"`

	// Size represents the capacity of the pvc, unit supports 'Gi' or 'Mi'
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#capacity
	Size string `json:"size"`
}

// LoggingConfig configures logging
type LoggingConfig struct {
	Level string `json:"level"`
}

// Config is Workflow Controller config instance
var Config CycloneServerConfig

// LoadConfig loads configuration from ConfigMap
func LoadConfig(cm *core_v1.ConfigMap) error {
	data, ok := cm.Data[ConfigFileKey]
	if !ok {
		return fmt.Errorf("ConfigMap '%s' doesn't have data key '%s'", cm.Name, ConfigFileKey)
	}
	err := json.Unmarshal([]byte(data), &Config)
	if err != nil {
		log.Warningf("Unmarshal config data %v error: %v", data, err)
		return err
	}

	if !validate(&Config) {
		return fmt.Errorf("validate config failed")
	}

	log.Infof("cyclone server config: %v", Config)
	return nil
}

// validate validates some required configurations.
func validate(config *CycloneServerConfig) bool {
	if config.CycloneServerHost == "" {
		log.Warningf("CycloneServerHost not configured, will use default value '0.0.0.0'")
		config.CycloneServerHost = "0.0.0.0"
	}

	if config.CycloneServerPort == 0 {
		log.Warningf("CycloneServerPort not configured, will use default value '7099'")
		config.CycloneServerPort = 7099
	}

	return true
}
