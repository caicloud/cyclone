package coordinator

import (
	"os"
	"strings"

	"github.com/caicloud/cyclone/pkg/common"
)

func getPodName() string {
	return os.Getenv(common.EnvStagePodName)
}

func getWorkloadContainer() string {
	return os.Getenv(common.EnvWorkloadContainerName)
}

func getCycloneServerAddr() string {
	addr := os.Getenv(common.EnvCycloneServerAddr)
	if addr == "" {
		addr = common.DefaultCycloneServerAddr
	}
	return addr
}

func getNamespace() string {
	n := os.Getenv(common.EnvNamespace)
	if n == "" {
		return "default"
	}

	return n
}

// refineContainerID strips the 'docker://' prefix from k8s ContainerID string
func refineContainerID(id string) string {
	schemeIndex := strings.Index(id, "://")
	if schemeIndex == -1 {
		return id
	}
	return id[schemeIndex+3:]
}
