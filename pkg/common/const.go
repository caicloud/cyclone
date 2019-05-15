package common

import (
	"os"
)

const (
	// CycloneLogo defines ascii art logo of Cyclone
	CycloneLogo = `
   ______           __
  / ____/_  _______/ /___  ____  ___ 
 / /   / / / / ___/ / __ \/ __ \/ _ \
/ /___/ /_/ / /__/ / /_/ / / / /  __/
\____/\__, /\___/_/\____/_/ /_/\___/ 
     /____/
`

	// ControlClusterName is name of the master cluster
	ControlClusterName = "cyclone-control-cluster"

	// ControllerInstanceEnvName is environment name for workflow controller instance
	ControllerInstanceEnvName = "CONTROLLER_INSTANCE_NAME"

	// EnvSystemNamespace is the evn key to indicate which namespace the cyclone system components installed in.
	// Cyclone built-in resources(such as stage templates) will be stored in the namespace too.
	EnvSystemNamespace = "SYSTEM_NAMESPACE"
)

// GetSystemNamespace ...
func GetSystemNamespace() string {
	envNamespace := os.Getenv(EnvSystemNamespace)
	if envNamespace != "" {
		return envNamespace
	}

	// If SystemNamespace environment is not configured, will return default value 'default'.
	return "default"
}
