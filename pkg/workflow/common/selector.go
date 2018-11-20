package common

import (
	"strings"
)

type ContainerSelector func(name string) bool

// WorkloadContainerSelector selector workload containers.
func WorkloadContainersSelector(name string) bool {
	if strings.HasPrefix(name, CycloneSidecarPrefix) {
		return false
	}

	if strings.HasPrefix(name, WorkloadSidecarPrefix) {
		return false
	}

	return true
}

// Pass check whether the given container name passes the given selectors.
func Pass(name string, selectors []ContainerSelector) bool {
	for _, s := range selectors {
		if !s(name) {
			return false
		}
	}
	return true
}