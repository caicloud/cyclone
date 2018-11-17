package constants

const (
	// ContainerCoordinatorName is the container name of the workload sidecar
	// which is designed to collect logs and involve the output resolver sidecar
	// to start working.
	ContainerCoordinatorName = "cyclone-coordinator"

	// ContainerOutputResolverName is the container name of the workload sidecar
	// which is designed to process outputs.
	ContainerOutputResolverName = "cyclone-output-resolver"

	// EnvStagePodName is an environment which represents pod name.
	EnvStagePodName = "POD_NAME"

	// EnvWorkflowrunName is an environment which represents workflowrun name.
	EnvWorkflowrunName = "WORKFLOWRUN_NAME"

	// EnvStagePodName is an environment which represents stage name.
	EnvStageName = "STAGE_NAME"

	// EnvNamespace is an environment which represents namespace.
	EnvNamespace = "NAMESPACE"

	// FmtContainerLogPath is the format of container log path.
	FmtContainerLogPath = ContainerLogDir + "/%s.log"

	// ContainerLogDir is the containers log directory.
	ContainerLogDir = "/tmp/coordinator/logs"

	// OutputResolverStartFlagPath is the path of the file which is watched
	// by output resolver container, once the file exists, resolver starts to work.
	// And the file will created by coordinator container after all
	// customized containers completion.
	OutputResolverStartFlagPath = "/tmp/cyclone-flags/resolver.start"
)
