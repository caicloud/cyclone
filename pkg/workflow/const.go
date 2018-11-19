package workflow

const (
	ResourcePullCommand      = "pull"
	ResourcePushCommand      = "push"
	ResolverDataPath         = "/workspace/data"
	SidecarContainerPrefix   = "cyclone-sidecar-"
	CoordinatorContainerName = SidecarContainerPrefix + "coordinator"

	WorkflowLabelName         = "cyclone.io/workflow"
	PodLabelSelector          = "cyclone.io/workflow==true"
	WorkflowRunAnnotationName = "cyclone.io/workflowrun"
	StageAnnotationName       = "cyclone.io/stage"

	DockerSockPath = "/var/run/docker.sock"

	// Name of the default PV used by all workflow stages.
	DefaultPvVolumeName = "default-pv"
	// Name of the emptyDir volume shared between workload containers and coordinator.
	// It's mainly used by coordinator to copy artifacts from workload containers.
	CoordinatorVolumeName = "coordinator-volume"
	// Volume name to mount host /var/run/docker.sock to container, it's used by coordinator.
	DockerSockVolume = "docker-sock"
)
