package v1alpha1

import (
	"fmt"
	"os"
	"strings"
)

const (
	// cycloneHome is the home folder for Cyclone.
	cycloneHome = "/var/lib/cyclone"

	// logsFolderName is the folder name for logs files.
	logsFolderName = "logs"
)

func getLogFilePath(workflowrun, stage, container, namespace string) (string, error) {
	if workflowrun == "" || stage == "" || container == "" {
		return "", fmt.Errorf("workflowrun/stage/container/namespace can not be empty")
	}

	rf, _ := getLogFolder(workflowrun, stage, namespace)
	return strings.Join([]string{rf, container}, string(os.PathSeparator)), nil
}

func getLogFolder(workflowrun, stage, namespace string) (string, error) {
	if workflowrun == "" || stage == "" || namespace == "" {
		return "", fmt.Errorf("workflowrun/stage/namespace can not be empty")
	}
	return strings.Join([]string{cycloneHome, namespace, workflowrun, stage, logsFolderName}, string(os.PathSeparator)), nil
}
