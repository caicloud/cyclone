package common

import "time"

// ContainerState represents container state.
type ContainerState string

const (
	// ResyncPeriod defines resync period for controllers
	ResyncPeriod = time.Minute * 5

	// EnvStagePodName is an environment which represents pod name.
	EnvStagePodName = "POD_NAME"
	// EnvWorkflowrunName is an environment which represents workflowrun name.
	EnvWorkflowrunName = "WORKFLOWRUN_NAME"
	// EnvStageName is an environment which represents stage name.
	EnvStageName = "STAGE_NAME"
	// EnvWorkloadContainerName is an environment which represents the workload container name.
	EnvWorkloadContainerName = "WORKLOAD_CONTAINER_NAME"
	// EnvNamespace is an environment which represents namespace.
	EnvNamespace = "NAMESPACE"
	// EnvCycloneServerAddr is an environment which represents cyclone server address.
	EnvCycloneServerAddr = "CYCLONE_SERVER_ADDR"

	// DefaultCycloneServerAddr defines default Cyclone Server address
	DefaultCycloneServerAddr = "native-cyclone-server"

	// CycloneSidecarPrefix defines container name prefixes for sidecar. There are two kinds of
	// sidecars in workflow:
	// - Those added automatically by Cyclone such as coordinator, resource resolvers.
	// - Those specified by users in stage spec as workload.
	CycloneSidecarPrefix = "cyclone-sidecar-"

	// WorkloadSidecarPrefix defines workload sidecar container name prefix.
	WorkloadSidecarPrefix = "workload-sidecar-"

	// CoordinatorSidecarName defines name of coordinator container.
	CoordinatorSidecarName = CycloneSidecarPrefix + "coordinator"

	// ResolverDefaultDataPath is paths in resource resolver containers.
	// Default data path in resource resolver container.
	ResolverDefaultDataPath = "/workspace/data"
	// ResolverNotifyDir is name of the notify directory where coordinator would create ok file there.
	ResolverNotifyDir = "notify"
	// ResolverNotifyDirPath is notify directory path in resource resolver container.
	ResolverNotifyDirPath = "/workspace/notify"

	// ResourcePullCommand indicates pull resource
	ResourcePullCommand = "pull"
	// ResourcePushCommand indicates push resource
	ResourcePushCommand = "push"

	// WorkflowLabelName is label to indicate resources created by Cyclone workflow engine
	WorkflowLabelName = "cyclone.io/workflow"
	// WorkflowRunLabelName is label applied to WorkflowRun to specify Workflow
	WorkflowRunLabelName = "cyclone.io/workflow-name"
	// PodLabelSelector is selector used to select pod created by Cyclone stages
	PodLabelSelector = "cyclone.io/workflow==true"
	// WorkflowRunAnnotationName is annotation applied to pod to specify WorkflowRun the pod belongs to
	WorkflowRunAnnotationName = "cyclone.io/workflowrun"
	// GCAnnotationName is annotation applied to pod to indicate whether the pod is used for GC purpose
	GCAnnotationName = "cyclone.io/gc"
	// StageAnnotationName is annotation applied to pod to indicate which stage it related to
	StageAnnotationName = "cyclone.io/stage"
	// StageTemplateLabelName indicates whether a stage is used as stage template
	StageTemplateLabelName = "cyclone.io/stage-template"
	// StageTemplateLabelSelector is label selector to select stage templates
	StageTemplateLabelSelector = "cyclone.io/stage-template=true"

	// CoordinatorResolverPath ...
	CoordinatorResolverPath = "/workspace/resolvers"
	// CoordinatorResourcesPath ...
	CoordinatorResourcesPath = "/workspace/resolvers/resources"
	// CoordinatorResolverNotifyPath ...
	CoordinatorResolverNotifyPath = "/workspace/resolvers/notify"
	// CoordinatorResolverNotifyOkPath ...
	CoordinatorResolverNotifyOkPath = "/workspace/resolvers/notify/ok"
	// CoordinatorArtifactsPath ...
	CoordinatorArtifactsPath = "/workspace/artifacts"

	// DefaultPvVolumeName is name of the default PV used by all workflow stages.
	DefaultPvVolumeName = "default-pv"
	// CoordinatorSidecarVolumeName is name of the emptyDir volume shared between coordinator and
	// sidecar containers, e.g. image resolvers. Coordinator would notify resolvers that workload
	// containers have finished their work, so that resource resolvers can push resources.
	CoordinatorSidecarVolumeName = "coordinator-sidecar-volume"
	// DockerSockVolume is volume name to mount host /var/run/docker.sock to container, it's used by coordinator.
	DockerSockVolume = "docker-sock"
	// DockerConfigJSONVolume is volume for config.json in secret.
	DockerConfigJSONVolume = "cyclone-docker-secret-volume"

	// DockerSockPath is path of docker socket file
	DockerSockPath = "/var/run/docker.sock"

	// DockerConfigPath is path of docker config
	DockerConfigPath = "/root/.docker"
	// DockerConfigJSONFile is name of docker config file
	DockerConfigJSONFile = "config.json"

	// ContainerStateTerminated represents container is stopped.
	ContainerStateTerminated ContainerState = "Terminated"
	// ContainerStateInitialized represents container is Running or Stopped, not Init or Creating.
	ContainerStateInitialized ContainerState = "Initialized"
)

const (
	// GCContainerName is name of GC container
	GCContainerName = "gc"
	// GCDataPath is parent folder holding data to be cleaned by GC pod
	GCDataPath = "/workspace"
)

// InputResourceVolumeName ...
func InputResourceVolumeName(name string) string {
	return "input-" + name
}

// OutputResourceVolumeName ...
func OutputResourceVolumeName(name string) string {
	return "output-" + name
}
