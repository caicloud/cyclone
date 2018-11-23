package controller

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"

	log "github.com/sirupsen/logrus"
	api_v1 "k8s.io/api/core/v1"
)

const (
	// The key of config file in ConfigMap
	ConfigFileKey = "workflow-controller.json"

	// Keys of images in config file
	GitResolverImage   = "git-resolver"
	ImageResolverImage = "image-resolver"
	KvResolverImage    = "kv-resolver"
	CoordinatorImage   = "coordinator"
	GCImage            = "gc"
)

var ResolverImageKeys = map[v1alpha1.ResourceType]string{
	v1alpha1.GitResourceType:   GitResolverImage,
	v1alpha1.ImageResourceType: ImageResolverImage,
	v1alpha1.KVResourceType:    KvResolverImage,
}

type ControllerConfig struct {
	// Images that used in controller, such as resource resolvers.
	Images map[string]string `json:"images"`
	// Logging configuration, such as log level.
	Logging LoggingConfig `json:"logging"`
	// GC configuration
	GC GCConfig `json:"gc"`
	// Limits of each resources should be retained
	Limits LimitsConfig `json:"limits"`
	// The PVC used to transfer artifacts in WorkflowRun
	PVC string `json:"pvc"`
}

type LoggingConfig struct {
	Level string `json:"level"`
}

type GCConfig struct {
	// Whether GC is enabled, it set to false, no GC would happen.
	Enabled bool `json:"enabled"`
	// After a WorkflowRun has been terminated, we won't clean it up immediately, but after a
	// delay time given by this configure item. When configured to 0, it equals to gc immediately.
	DelaySeconds time.Duration `json:"delay_seconds"`
	// How many times to retry when GC failed, 0 means no retry.
	RetryCount int `json:"retry"`
}

type LimitsConfig struct {
	// Maximum WorkflowRuns to be kept for each Workflow
	MaxWorkflowRuns int `json:"max_workflowruns"`
}

var Config ControllerConfig

func LoadConfig(cm *api_v1.ConfigMap) error {
	data, ok := cm.Data[ConfigFileKey]
	if !ok {
		fmt.Errorf("ConfigMap '%s' doesn't have data key '%s'", cm.Name, ConfigFileKey)
	}
	err := json.Unmarshal([]byte(data), &Config)
	if err != nil {
		log.WithField("data", data).Debug("Unmarshal config data error: ", err)
		return err
	}

	if !validate(&Config) {
		return fmt.Errorf("validate config failed")
	}

	InitLogger(&Config.Logging)
	return nil
}

// validate validates some required configurations.
func validate(config *ControllerConfig) bool {
	for _, k := range []string{GitResolverImage, ImageResolverImage, KvResolverImage, CoordinatorImage} {
		_, ok := config.Images[k]
		if !ok {
			return false
		}
	}

	return true
}
