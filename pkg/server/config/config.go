package config

import (
	"log"
	"os"
	"strconv"
)

const (
	EnvCycloneServerPort = "ENV_CYCLONE_SERVER_PORT"
	EnvCycloneServerHost = "ENV_CYCLONE_SERVER_HOST"
	EnvKubeHost          = "ENV_KUBE_HOST"
	EnvKubeConfig        = "ENV_KUBE_CONFIG"
	EnvLogLevel          = "ENV_LOG_LEVEL"

	FlagCycloneServerPort = "cyclone-server-port"
	FlagCycloneServerHost = "cyclone-server-host"
	FlagKubeHost          = "kubehost"
	FlagKubeConfig        = "kubeconfig"
	FlagLogLevel          = "log-level"

	envVarCannotEmptyFormat = "The environment variable '%s' cannot be empty."

	DefaultCycloneServerPort int = 7099

	DefaultCycloneServerHost = "0.0.0.0"

	DefaultLogLevel = "info"
)

var (
	// listen port
	CycloneServerPort int
	// cyclone server
	CycloneServerHost string

	// k8s about
	KubeHost   string
	KubeConfig string

	// log
	LogLevel string
)

func init() {
	// listen port
	CycloneServerPort = LoadIntEnvVar(EnvCycloneServerPort, DefaultCycloneServerPort, true)
	// cyclone server
	CycloneServerHost = LoadEnvVar(EnvCycloneServerHost, DefaultCycloneServerHost, true)

	// k8s
	KubeHost = LoadEnvVar(EnvKubeHost, "", true)
	KubeConfig = LoadEnvVar(EnvKubeConfig, "", true)

	// log
	LogLevel = LoadEnvVar(EnvLogLevel, DefaultLogLevel, true)
}

// GetStringEnvWithDefault retrieves the value of the environment variable named
// by the key. If the variable is present in the environment the
// value is returned. Otherwise the default is returned.
func GetStringEnvWithDefault(name, defaultValue string) string {
	v, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	return v
}

// LoadEnvVar retrieves the value of 'name' frome environment.
// If canBeEmpty is false and the value of the 'name' is empty, os will exit.
func LoadEnvVar(name string, defaultValue string, canBeEmpty bool) string {
	value := GetStringEnvWithDefault(name, defaultValue)
	if len(value) == 0 && !canBeEmpty {
		log.Fatalf(envVarCannotEmptyFormat, name)
		os.Exit(-1)
	}
	return value
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
