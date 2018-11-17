package coordinator

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/caicloud/cyclone/pkg/common/constants"
)

func getPodName() string {
	return os.Getenv(constants.EnvStagePodName)
}

func getWorkflowrunName() string {
	return os.Getenv(constants.EnvWorkflowrunName)
}

func getStageName() string {
	return os.Getenv(constants.EnvStageName)
}

func getNamespace() string {
	n := os.Getenv(constants.EnvNamespace)
	if n == "" {
		return "default"
	}

	return n
}

func createDirectory(dirName string) bool {
	src, err := os.Stat(dirName)

	if os.IsNotExist(err) {
		errDir := os.MkdirAll(dirName, 0755)
		if errDir != nil {
			panic(err)
		}
		return true
	}

	if src.Mode().IsRegular() {
		log.Error(dirName, "already exist as a file!")
		return false
	}

	return false
}
