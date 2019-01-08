package controller

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
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
	// Default resource requirements for containers in stage Pod
	ResourceRequirements corev1.ResourceRequirements `json:"default_resource_quota"`
	// The PVC used to transfer artifacts in WorkflowRun, and also to help share resources
	// among stages within WorkflowRun. If no PVC is given here, input resources won't be
	// shared among stages, but need to be pulled every time it's needed. And also if no
	// PVC given, artifacts are not supported.
	// TODO(ChenDe): Remove it when Cyclone can manage PVC for namespaces.
	PVC string `json:"pvc"`
	// Default secret used for Cyclone, auth of registry can be placed here. It's optional.
	// TODO(ChenDe): Remove it when Cyclone can manage secrets for namespaces.
	Secret string `json:secret`
	// Address of the Cyclone Server
	CycloneServerAddr string `json:"cyclone_server_addr"`
	// ServiceAccount used for coordinator to access k8s resources.
	ServiceAccount string `json:"service_account"`
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

func LoadConfig(cm *corev1.ConfigMap) error {
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
	if config.PVC == "" {
		log.Warn("PVC not configured, resources won't be shared among stages and artifacts unsupported.")
	}

	if config.Secret == "" {
		log.Warn("Secret not configured, no auth information would be available, e.g. docker registry auth.")
	}

	for _, k := range []string{GitResolverImage, ImageResolverImage, KvResolverImage, CoordinatorImage} {
		_, ok := config.Images[k]
		if !ok {
			return false
		}
	}

	return true
}
