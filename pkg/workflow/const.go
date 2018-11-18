package workflow

const (
	SidecarContainerPrefix   = "sidecar.cyclone.io/"
	ResolverDataPath         = "/workspace/data"
	CoordinatorContainerName = "coordinator"

	WorkflowLabelName         = "cyclone.io/workflow"
	PodLabelSelector          = "cyclone.io/workflow==true"
	WorkflowrunAnnotationName = "cyclone.io/workflowrun"
	StageAnnotationName       = "cyclone.io/stage"
)
