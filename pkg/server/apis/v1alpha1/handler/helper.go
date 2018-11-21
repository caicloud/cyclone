package handler

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

func getLogFilePath(workflowrun, stage, container string) (string, error) {
	if workflowrun == "" || stage == "" || container == "" {
		return "", fmt.Errorf("workflowrun or stage or container can not be empty")
	}

	rf, _ := getLogFolder(workflowrun, stage)
	return strings.Join([]string{rf, container}, string(os.PathSeparator)), nil
}

func getLogFolder(workflowrun, stage string) (string, error) {
	if workflowrun == "" || stage == "" {
		return "", fmt.Errorf("workflowrun or stage can not be empty")
	}
	return strings.Join([]string{cycloneHome, workflowrun, stage, logsFolderName}, string(os.PathSeparator)), nil
}
