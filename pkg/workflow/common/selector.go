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

// AllContainers selects all containers, it returns true regardless of the container name.
func AllContainers(string) bool {
	return true
}

// OnlyCustomContainer judges whether a container is a custom container based on container name.
// Containers added by Cyclone would have CycloneSidecarPrefix prefix in container names.
func OnlyCustomContainer(name string) bool {
	return !strings.HasPrefix(name, CycloneSidecarPrefix)
}

// NonWorkloadSidecar selects all containers except workload sidecars.
func NonWorkloadSidecar(name string) bool {
	return !strings.HasPrefix(name, WorkloadSidecarPrefix)
}

// NonCoordinator selects all containers except coordinator.
func NonCoordinator(name string) bool {
	return name != CoordinatorSidecarName
}

// NonDockerInDocker selects all containers except docker:dind.
func NonDockerInDocker(name string) bool {
	return name != DockerInDockerSidecarName
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
