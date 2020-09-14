package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/caicloud/nirvana/log"
	core_v1 "k8s.io/api/core/v1"

	"github.com/caicloud/cyclone/pkg/server/common"
)

const (
	// ConfigFileKey is key of config file in ConfigMap
	ConfigFileKey = "cyclone-server.json"

	// EnvWebhookURLPrefix is the key of Environment variable to define webhook callback url prefix
	EnvWebhookURLPrefix = "WEBHOOK_URL_PREFIX"

	// EnvRecordWebURLTemplate is the key of Environment variable to define template of record url which used in
	// PR status 'Details' to associate PR with WorkflowRun website.
	EnvRecordWebURLTemplate = "RECORD_WEB_URL_TEMPLATE"
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

	// MongoConfig represents the config of mongodb config
	MongoConfig MongoConfig `json:"mongo_config"`

	// WorkerNamespaceQuota describes the resource quota of the namespace which will be used to run workflows,
	// eg map[core_v1.ResourceName]string{"cpu": "2", "memory": "4Gi"}
	WorkerNamespaceQuota map[core_v1.ResourceName]string `json:"worker_namespace_quota"`

	// WebhookURLPrefix represents the Cyclone server path to receive webhook requests.
	// If Cyclone server can be accessed by external systems, it would like be `https://{cyclone-server}/apis/v1alpha1`.
	WebhookURLPrefix string `json:"webhook_url_prefix"`

	// StorageUsageWatcher configures PVC storage usage watchers.
	StorageUsageWatcher StorageUsageWatcher `json:"storage_usage_watcher"`

	// CreateBuiltinTemplates configures whether to create builtin stage templates while cyclone server start up.
	CreateBuiltinTemplates bool `json:"create_builtin_templates"`

	// InitDefaultTenant configures whether to create cyclone default tenant while cyclone server start up.
	InitDefaultTenant bool `json:"init_default_tenant"`

	// OpenControlCluster indicates whether to open control cluster for workflow execution when tenant created
	OpenControlCluster bool `json:"open_control_cluster"`

	// Images that used in cyclone, such as GC image.
	Images map[string]string `json:"images"`

	// Notifications represents the config to send notifications after workflowruns finish.
	Notifications []NotificationEndpoint `json:"notifications"`

	// RecordWebURLTemplate represents the URL template to generate web URLs for workflowruns.
	RecordWebURLTemplate string `json:"record_web_url_template"`

	// ClientSet holds the common attributes that can be passed to a Kubernetes client on cyclone server handlers
	// initialization.
	ClientSet ClientSetConfig `json:"client_set"`
}

// ClientSetConfig defines rate limit config for a Kubernetes client
type ClientSetConfig struct {
	// QPS indicates the maximum QPS to the master from this client.
	// If it's zero, the created RESTClient will use DefaultQPS: 5
	QPS float32 `json:"qps"`

	// Maximum burst for throttle.
	// If it's zero, the created RESTClient will use DefaultBurst: 10.
	Burst int `json:"burst"`
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

// StorageUsageWatcher configures PVC storage usage watchers.
type StorageUsageWatcher struct {
	// Image is image for the storage usage watcher, for example 'busybox:1.30.0'
	Image string `json:"image"`
	// ReportURL is url where to report the usage
	ReportURL string `json:"report_url"`
	// IntervalSeconds is intervals to report storage usage
	IntervalSeconds string `json:"interval_seconds"`
	// ResourceRequirements specifies resource requirements of the watcher container.
	ResourceRequirements map[core_v1.ResourceName]string `json:"resource_requirements"`
}

// NotificationEndpoint represents the config of notification endpoint.
// Server will send notifications about finished workflowruns
// if notification endpoints are configured.
type NotificationEndpoint struct {
	// Name represents the name of notification endpoint.
	Name string `json:"name"`

	// URL represents the URL to send the notification.
	URL string `json:"url"`
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

	modifier(&Config)

	log.Infof("cyclone server config: %v", Config)
	return nil
}

// validate validates some required configurations.
func validate(config *CycloneServerConfig) bool {
	return validateNotification(config.Notifications)
}

// validateNotification validates notification configurations.
// The names of notification endpoints must be unique.
func validateNotification(nes []NotificationEndpoint) bool {
	names := map[string]struct{}{}
	for _, ne := range nes {
		if _, ok := names[ne.Name]; ok {
			log.Errorf("There are multiple notification endpoints with same name: %s", ne.Name)
			return false
		}

		names[ne.Name] = struct{}{}
	}

	return true
}

// modifier modifies the config, give the config some default value if they are not set.
func modifier(config *CycloneServerConfig) {
	if config.CycloneServerHost == "" {
		log.Warning("CycloneServerHost not configured, will use default value '0.0.0.0'")
		config.CycloneServerHost = "0.0.0.0"
	}

	if config.CycloneServerPort == 0 {
		log.Warning("CycloneServerPort not configured, will use default value '7099'")
		config.CycloneServerPort = 7099
	}

	if config.DefaultPVCConfig.Size == "" {
		log.Warning("DefaultPVCConfig.Size not configured, will use default value '5Gi'")
		config.DefaultPVCConfig.Size = common.DefaultPVCSize
	}

	if config.WorkerNamespaceQuota == nil {
		log.Warning("WorkerNamespaceQuota not configured, will use default quota")
		config.WorkerNamespaceQuota = map[core_v1.ResourceName]string{
			core_v1.ResourceLimitsCPU:      common.QuotaCPULimit,
			core_v1.ResourceLimitsMemory:   common.QuotaMemoryLimit,
			core_v1.ResourceRequestsCPU:    common.QuotaCPURequest,
			core_v1.ResourceRequestsMemory: common.QuotaMemoryRequest,
		}
	}
}

// GetWebhookURLPrefix returns webhook callback url prefix. It tries to get the url from "WEBHOOK_URL_PREFIX"
// environment variable, if the value is empty, then get it from configmap.
func GetWebhookURLPrefix() string {
	urlPrefix := os.Getenv(EnvWebhookURLPrefix)
	if urlPrefix != "" {
		return urlPrefix
	}

	return Config.WebhookURLPrefix
}

// GetRecordWebURLTemplate returns record web URL template. It tries to get the url from "RECORD_WEB_URL_TEMPLATE"
// environment variable, if the value is empty, then get it from configmap.
func GetRecordWebURLTemplate() string {
	template := os.Getenv(EnvRecordWebURLTemplate)
	if template != "" {
		return template
	}
	return Config.RecordWebURLTemplate
}

// Mongo Safe struct in mgo.v2
type Safe struct {
	W        int    `json:"w"`        // Min # of servers to ack before success
	WMode    string `json:"wmode"`    // Write mode for MongoDB 2.0+ (e.g. "majority")
	WTimeout int    `json:"wtimeout"` // Milliseconds to wait for W before timing out
	FSync    bool   `json:"fsync"`    // Sync via the journal if present, or via data files sync otherwise
	J        bool   `json:"j"`        // Sync via the journal if present
}

type MongoConfig struct {
	Addrs string `json:"addrs"`
	DB    string `json:"db"`
	Mode  string `json:"mode"`
	Safe  *Safe  `json:"safe"`
}
