package controller

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

const (
	// DevModeEnvName determines whether workflow controller is in development mode.
	// In development mode, resource resolver containers, coordinator containers will
	// have image pull policy being 'Always', otherwise it's 'IfNotPresent'.
	DevModeEnvName = "DEVELOP_MODE"

	// ConfigFileKey is key of config file in ConfigMap
	ConfigFileKey = "workflow-controller.json"

	// CoordinatorImage is key of coordinator image in config file
	CoordinatorImage = "coordinator"
	// GCImage is key of the GC image in config file
	GCImage = "gc"
	// DindImage is key of the docker-in-docker image in config file
	DindImage = "dind"
)

// WorkflowControllerConfig configures Workflow Controller
type WorkflowControllerConfig struct {
	// Images that used in controller, such as gc image.
	Images map[string]string `json:"images"`
	// Logging configuration, such as log level.
	Logging LoggingConfig `json:"logging"`
	// GC configuration
	GC GCConfig `json:"gc"`
	// Limits of each resources should be retained
	Limits LimitsConfig `json:"limits"`
	// ResourceRequirements is default resource requirements for containers in stage Pod
	ResourceRequirements corev1.ResourceRequirements `json:"default_resource_quota"`
	// ExecutionContext defines default namespace and pvc used to run workflow.
	ExecutionContext ExecutionContext `json:"execution_context"`
	// CycloneServerAddr is address of the Cyclone Server
	CycloneServerAddr string `json:"cyclone_server_addr"`
	// NotificationURL represents the config to send notifications after workflowruns finish.
	// It can be configured as Cyclone server notification URL to take advantage of its scenarized functions.
	NotificationURL string `json:"notification_url"`
	// DindSettings is settings for Docker in Docker
	DindSettings DindSettings `json:"dind"`
}

// LoggingConfig configures logging
type LoggingConfig struct {
	Level string `json:"level"`
}

// ExecutionContext defines default namespace and pvc used to run workflow.
type ExecutionContext struct {
	// Namespace is namespace where to run workflow.
	Namespace string `json:"namespace"`
	// PVC is pvc used to run workflow. It's used to transfer artifacts in WorkflowRun, and
	// also to help share resources among stages within WorkflowRun. If no PVC is given here,
	// input resources won't be shared among stages, but need to be pulled every time it's needed.
	// And also if no PVC given, artifacts are not supported.
	PVC string `json:"pvc"`
	// ServiceAccount is the service account applied to the pod runed
	ServiceAccount string `json:"service_account"`
}

// GCConfig configures GC
type GCConfig struct {
	// Enabled controllers whether GC is enabled, if set to false, no GC would happen.
	Enabled bool `json:"enabled"`
	// DelaySeconds defines the time after a WorkflowRun terminated to perform GC. When configured to 0.
	// it equals to gc immediately.
	DelaySeconds time.Duration `json:"delay_seconds"`
	// RetryCount defines how many times to retry when GC failed, 0 means no retry.
	RetryCount int `json:"retry"`
}

// LimitsConfig configures maximum WorkflowRun to keep for each Workflow
type LimitsConfig struct {
	// Maximum WorkflowRuns to be kept for each Workflow
	MaxWorkflowRuns int `json:"max_workflowruns"`
}

// DindSettings is settings for Docker in Docker.
type DindSettings struct {
	// InsecureRegistries is list of insecure registries, for docker registries with
	// self-signed certs, it's useful to bypass the cert check.
	InsecureRegistries []string `json:"insecure_registries"`
	// Bip specifies IP subnet used for docker0 bridge
	Bip string `json:"bip"`
}

// Config is Workflow Controller config instance
var Config WorkflowControllerConfig

// LoadConfig loads configuration from ConfigMap
func LoadConfig(cm *corev1.ConfigMap) error {
	data, ok := cm.Data[ConfigFileKey]
	if !ok {
		return fmt.Errorf("ConfigMap '%s' doesn't have data key '%s'", cm.Name, ConfigFileKey)
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
func validate(config *WorkflowControllerConfig) bool {
	if config.ExecutionContext.PVC == "" {
		log.Warn("PVC not configured, resources won't be shared among stages and artifacts unsupported.")
	}

	return true
}

// ImagePullPolicy determines image pull policy based on environment variable DEVELOP_MODE
// This pull policy will be used in image resolver containers and coordinator containers.
func ImagePullPolicy() corev1.PullPolicy {
	if os.Getenv(DevModeEnvName) == "true" {
		return corev1.PullAlways
	}
	return corev1.PullIfNotPresent
}
