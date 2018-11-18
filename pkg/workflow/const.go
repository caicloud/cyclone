package workflow

const (
	SidecarContainerPrefix   = "sidecar.cyclone.io/"
	ResolverDataPath         = "/workspace/data"
	CoordinatorContainerName = "coordinator"

	WorkflowLabelName         = "cyclone.io/workflow"
	PodLabelSelector          = "cyclone.io/workflow==true"
	WorkflowrunAnnotationName = "cyclone.io/workflowrun"
	StageAnnotationName       = "cyclone.io/stage"

	// Name of the default PV used by all workflow stages.
	DefaultPvVolumeName = "default-pv"
	// Name of the emptyDir volume shared between workload containers and coordinator.
	// It's mainly used by coordinator to copy artifacts from workload containers.
	CoordinatorVolumeName = "coordinator-volume"
)
