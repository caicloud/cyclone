package common

import (
	"strings"
)

// ContainerSelector is a function to select containers
type ContainerSelector func(name string) bool

// OnlyWorkload selects only workload containers.
func OnlyWorkload(name string) bool {
	if strings.HasPrefix(name, CycloneSidecarPrefix) {
		return false
	}

	if strings.HasPrefix(name, WorkloadSidecarPrefix) {
		return false
	}

	return true
}

// NonWorkloadSidecar selects all containers except workload sidecars.
func NonWorkloadSidecar(name string) bool {
	if strings.HasPrefix(name, WorkloadSidecarPrefix) {
		return false
	}

	return true
}

// NonCoordinator selects all containers except coordinator.
func NonCoordinator(name string) bool {
	return name != CoordinatorSidecarName
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
