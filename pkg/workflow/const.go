package workflow

const (
	// Default data path in resource resolver container.
	ResolverDefaultDataPath = "/workspace/data"
	// Name of the notify directory where coordinator would create ok file there.
	ResolverNotifyDir = "notify"
	// Notify directory path in resource resolver container.
	ResolverNotifyDirPath  = "/workspace/notify"
	ResourcePullCommand = "pull"
	ResourcePushCommand = "push"

	SidecarContainerPrefix   = "cyclone-sidecar-"
	CoordinatorContainerName = SidecarContainerPrefix + "coordinator"

	WorkflowLabelName         = "cyclone.io/workflow"
	PodLabelSelector          = "cyclone.io/workflow==true"
	WorkflowRunAnnotationName = "cyclone.io/workflowrun"
	StageAnnotationName       = "cyclone.io/stage"

	// Paths in coordinator container.
	DockerSockPath = "/var/run/docker.sock"
	ResolverPath   = "/workspace/resolvers"

	// Name of the default PV used by all workflow stages.
	DefaultPvVolumeName = "default-pv"
	// Name of the emptyDir volume shared between coordinator and sidecar containers, e.g.
	// image resolvers. Coordinator would notify resolvers that workload containers have
	// finished their work, so that resource resolvers can push resources.
	CoordinatorSidecarVolumeName = "coordinator-sidecar-volume"
	// Volume name to mount host /var/run/docker.sock to container, it's used by coordinator.
	DockerSockVolume = "docker-sock"
)
