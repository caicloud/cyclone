package config

import (
	"log"
	"os"
	"strconv"

	"github.com/caicloud/storage-admin/pkg/constants"
)

const (
	EnvListenPort = "ENV_LISTEN_PORT"

	EnvClusterAdminHost    = "ENV_CLUSTER_ADMIN_HOST"
	EnvWatchIntervalSecond = "ENV_WATCH_INTERVAL_SECOND"

	EnvKubeHost   = "ENV_KUBE_HOST"
	EnvKubeConfig = "ENV_KUBE_CONFIG"

	EnvControllerMaxRetryTimes      = "ENV_CONTROLLER_MAX_RETRY_TIMES"
	EnvControllerResyncPeriodSecond = "ENV_CONTROLLER_RESYNC_PERIOD_SECOND"

	EnvCtrlClusterName = "ENV_CTRL_CLUSTER_NAME"

	EnvLogLevel     = "ENV_LOG_LEVEL"
	EnvLogVerbosity = "ENV_LOG_VERBOSITY"
)

var (
	// listen port
	ListenPort int
	// cluster admin
	ClusterAdminHost    string
	WatchIntervalSecond int
	// k8s about
	KubeHost   string
	KubeConfig string

	// controller
	ControllerMaxRetryTimes      int
	ControllerResyncPeriodSecond int

	// test
	CtrlClusterName string

	// log
	LogLevel     string
	LogVerbosity string
)

const (
	envVarCannotEmptyFormat = "The environment variable '%s' cannot be empty."
)

func init() {
	// listen port
	ListenPort = LoadIntEnvVar(EnvListenPort, constants.DefaultListenPort, true)
	// cluster admin
	ClusterAdminHost = LoadEnvVar(EnvClusterAdminHost, "", true)
	WatchIntervalSecond = LoadIntEnvVar(EnvWatchIntervalSecond, constants.DefaultWatchIntervalSecond, true)
	// k8s
	KubeHost = LoadEnvVar(EnvKubeHost, "", true)
	KubeConfig = LoadEnvVar(EnvKubeConfig, "", true)
	// controller
	ControllerMaxRetryTimes = LoadIntEnvVar(EnvControllerMaxRetryTimes, constants.DefaultControllerMaxRetryTimes, true)
	ControllerResyncPeriodSecond = LoadIntEnvVar(EnvControllerResyncPeriodSecond, constants.DefaultControllerResyncPeriodSecond, true)
	// test
	CtrlClusterName = LoadEnvVar(EnvCtrlClusterName, constants.ControlClusterName, true)
	// log
	LogLevel = LoadEnvVar(EnvLogLevel, constants.DefaultLogLevel, true)
	LogVerbosity = LoadEnvVar(EnvLogVerbosity, constants.DefaultLogVerbosity, true)
}

func GetStringEnvWithDefault(name, defaultValue string) string {
	ev, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	return ev
}

func LoadEnvVar(name string, defaultValue string, canBeEmpty bool) string {
	val := GetStringEnvWithDefault(name, defaultValue)
	if len(val) == 0 && !canBeEmpty {
		log.Fatalf(envVarCannotEmptyFormat, name)
		os.Exit(-1)
	}
	return val
}

func LoadFloatEnvVar(name string, defaultValue float64, canBeEmpty bool) float64 {
	val := LoadEnvVar(name, "", canBeEmpty)
	if len(val) == 0 {
		return defaultValue
	} else {
		floatVal, err := strconv.ParseFloat(val, 64)
		if err != nil {
			log.Fatalf("The value '%s' of environment variable '%s' is not a valid float64: %v\n", val, name, err)
			os.Exit(-1)
		}
		return floatVal
	}
}

func LoadIntEnvVar(name string, defaultValue int, canBeEmpty bool) int {
	val := LoadEnvVar(name, "", canBeEmpty)
	if len(val) == 0 {
		return defaultValue
	} else {
		intVal, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			log.Fatalf("The value '%s' of environment variable '%s' is not a valid integer: %v\n", val, name, err)
			os.Exit(-1)
		}
		return int(intVal)
	}
}
