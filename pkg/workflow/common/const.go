package common

// ContainerState represents container state.
type ContainerState string

const (
	// EnvStagePodName is an environment which represents pod name.
	EnvStagePodName = "POD_NAME"
	// EnvWorkflowrunName is an environment which represents workflowrun name.
	EnvWorkflowrunName = "WORKFLOWRUN_NAME"
	// EnvStagePodName is an environment which represents stage name.
	EnvStageName = "STAGE_NAME"
	// EnvWorkloadContainerName is an environment which represents the workload container name.
	EnvWorkloadContainerName = "WORKLOAD_CONTAINER_NAME"
	// EnvNamespace is an environment which represents namespace.
	EnvNamespace = "NAMESPACE"

	// Container name prefixes for sidecar. There are two kinds of sidecars in workflow:
	// - Those added automatically by Cyclone such as coordinator, resource resolvers.
	// - Those specified by users in stage spec as workload.
	CycloneSidecarPrefix = "cyclone-sidecar-"
	WorkloadSidecarPrefix = "workload-sidecar-"

	// Coordinator container name.
	CoordinatorSidecarName = CycloneSidecarPrefix + "coordinator"

	// Paths in resource resolver containers.
	// Default data path in resource resolver container.
	ResolverDefaultDataPath = "/workspace/data"
	// Name of the notify directory where coordinator would create ok file there.
	ResolverNotifyDir = "notify"
	// Notify directory path in resource resolver container.
	ResolverNotifyDirPath  = "/workspace/notify"

	ResourcePullCommand = "pull"
	ResourcePushCommand = "push"

	WorkflowLabelName         = "cyclone.io/workflow"
	PodLabelSelector          = "cyclone.io/workflow==true"
	WorkflowRunAnnotationName = "cyclone.io/workflowrun"
	StageAnnotationName       = "cyclone.io/stage"

	// Paths in coordinator container.
	CoordinatorResolverPath   = "/workspace/resolvers"
	CoordinatorResourcesPath = "/workspace/resolvers/resources"
	CoordinatorResolverNotifyPath = "/workspace/resolvers/notify"
	CoordinatorResolverNotifyOkPath = "/workspace/resolvers/notify/ok"
	CoordinatorLogsPath = "/workspace/logs"
	CoordinatorArtifactsPath = "/workspace/artifacts"

	// Name of the default PV used by all workflow stages.
	DefaultPvVolumeName = "default-pv"
	// Name of the emptyDir volume shared between coordinator and sidecar containers, e.g.
	// image resolvers. Coordinator would notify resolvers that workload containers have
	// finished their work, so that resource resolvers can push resources.
	CoordinatorSidecarVolumeName = "coordinator-sidecar-volume"
	// Volume name to mount host /var/run/docker.sock to container, it's used by coordinator.
	DockerSockVolume = "docker-sock"

	// Path of docker socket file
	DockerSockPath = "/var/run/docker.sock"

	// ContainerStateTerminated represents container is stopped.
	ContainerStateTerminated ContainerState = "Terminated"
	// ContainerStateInitialized represents container is Running or Stopped, not Init or Creating.
	ContainerStateInitialized ContainerState = "Initialized"
)
