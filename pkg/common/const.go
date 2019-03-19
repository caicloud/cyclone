package common

import (
	"time"
)

const (
	// CycloneLogo defines ascii art logo of Cyclone
	CycloneLogo = `
	______           __
   / ____/_  _______/ /___  ____  ___ 
  / /   / / / / ___/ / __ \/ __ \/ _ \
 / /___/ /_/ / /__/ / /_/ / / / /  __/
 \____/\__, /\___/_/\____/_/ /_/\___/ 
	  /____/
 `

	// AdminTenant is name of the system admin tenant, it's a default tenant created when Cyclone
	// start, and resources shared among all tenants would be placed in this tenant, such as stage
	// templates.
	AdminTenant = "admin"

	// TenantNamespacePrefix is the prefix of namespace which related to a specific tenant
	TenantNamespacePrefix = "cyclone-"

	// TenantPVCPrefix is the prefix of pvc which related to a specific tenant
	TenantPVCPrefix = "cyclone-pvc-"

	// DefaultPVCSize is the default size of pvc
	DefaultPVCSize = "5Gi"

	// AnnotationTenant is the annotation key used for namespace to relate tenant information
	AnnotationTenant = "cyclone.io/tenant-info"

	// AnnotationAlias is the annotation key used to indicate the alias of resources
	AnnotationAlias = "cyclone.io/alias"

	// AnnotationDescription is the annotation key used to describe resources
	AnnotationDescription = "cyclone.io/description"

	// WorkflowRunAnnotationName is annotation applied to pod to specify WorkflowRun the pod belongs to
	WorkflowRunAnnotationName = "workflowrun.cyclone.io"

	// MetaNamespaceAnnotationName is annotation applied to pod to specify the namespace where Workflow, WorkflowRun etc belong to.
	MetaNamespaceAnnotationName = "cyclone.io/meta-namespace"

	// GCAnnotationName is annotation applied to pod to indicate whether the pod is used for GC purpose
	GCAnnotationName = "cyclone.io/gc"

	// StageAnnotationName is annotation applied to pod to indicate which stage it related to
	StageAnnotationName = "stage.cyclone.io"

	// LabelProjectName is the label key used to indicate the project which the resources belongs to
	LabelProjectName = "project.cyclone.io/name"

	// LabelWorkflowName is the label key used to indicate the workflow which the resources belongs to
	LabelWorkflowName = "workflow.cyclone.io/name"

	// LabelIntegrationType is the label key used to indicate type of integration
	LabelIntegrationType = "integration.cyclone.io/type"

	// LabelClusterOn is the label key used to indicate the cluster is a worker for the tenant
	LabelClusterOn = "cyclone.io/cluster-worker"

	// LabelStageTemplate is the label key used to represent a stage is a stage template
	LabelStageTemplate = "stage.cyclone.io/template"

	// WorkflowLabelName is label to indicate resources created by Cyclone workflow engine
	WorkflowLabelName = "workflow.cyclone.io"
	// WorkflowNameLabelName is label applied to WorkflowRun to specify Workflow
	// Deprecated, use LabelWorkflowName instead.
	WorkflowNameLabelName = "cyclone.io/workflow-name"
	// PodLabelSelector is selector used to select pod created by Cyclone stages
	// Deprecated.
	PodLabelSelector = "cyclone.io/workflow==true"
	// StageTemplateLabelSelector is label selector to select stage templates
	// Deprecated.
	StageTemplateLabelSelector = "cyclone.io/stage-template=true"

	// LabelTrueValue is the label value used to represent true
	LabelTrueValue = "true"

	// LabelFalseValue is the label value used to represent false
	LabelFalseValue = "false"

	// LabelOwner is the label key used to indicate namespaces created by cyclone
	LabelOwner = "cyclone.io/owner"

	// OwnerCyclone is the label value used to indicate namespaces created by cyclone
	OwnerCyclone = "cyclone"

	// LabelBuiltin is the label key used to represent cyclone built in resources
	LabelBuiltin = "cyclone.io/builtin"

	// LabelScene is the label key used to indicate cyclone scenario
	LabelScene = "cyclone.io/scene"

	// SceneCICD is the label value used to indicate cyclone CI/CD scenario
	SceneCICD = "cicd"

	// SceneAI is the label value used to indicate cyclone AI scenario
	SceneAI = "ai"

	// QuotaCPULimit represents default value of 'limits.cpu'
	QuotaCPULimit = "2"
	// QuotaCPURequest represents default value of 'requests.cpu'
	QuotaCPURequest = "1"
	// QuotaMemoryLimit represents default value of 'limits.memory'
	QuotaMemoryLimit = "4Gi"
	// QuotaMemoryRequest represents default value of 'requests.memory'
	QuotaMemoryRequest = "2Gi"

	// SecretKeyIntegration is the key of the secret dada to indicate its value is about integration information.
	SecretKeyIntegration = "integration"

	// ControlClusterName is the name of control cluster
	ControlClusterName = "control-cluster"

	// ResyncPeriod defines resync period for controllers
	ResyncPeriod = time.Minute * 5

	// EnvStagePodName is an environment which represents pod name.
	EnvStagePodName = "POD_NAME"
	// EnvStageInfo is an environment which represents stage information.
	EnvStageInfo = "STAGE_INFO"
	// EnvWorkflowRunInfo is an environment which represents workflowrun information.
	EnvWorkflowRunInfo = "WORKFLOWRUN_INFO"
	// EnvOutputResourcesInfo is an environment which represents output resources information.
	EnvOutputResourcesInfo = "OUTPUT_RESOURCES_INFO"
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
	DefaultCycloneServerAddr = "cyclone-server"

	// CycloneSidecarPrefix defines container name prefixes for sidecar. There are two kinds of
	// sidecars in workflow:
	// - Those added automatically by Cyclone such as coordinator, resource resolvers.
	// - Those specified by users in stage spec as workload.
	CycloneSidecarPrefix = "csc-"

	// WorkloadSidecarPrefix defines workload sidecar container name prefix.
	WorkloadSidecarPrefix = "wsc-"

	// CoordinatorSidecarName defines name of coordinator container.
	CoordinatorSidecarName = CycloneSidecarPrefix + "co"

	// DockerInDockerSidecarName defines name of docker in docker container.
	DockerInDockerSidecarName = CycloneSidecarPrefix + "dind"

	// ResolverDefaultWorkspacePath is workspace path in resource resolver containers.
	// Following files or directories will be in this workspace.
	// - ${WORKFLOWRUN_NAME}-pulling.lock File lock determine which stage to pull the resource
	// - notify Directory contains notify file indicating readiness of output resource data.
	// - data Directory contains the data of the resource. For example, source code.
	ResolverDefaultWorkspacePath = "/workspace"
	// ResolverDefaultDataPath is data path in resource resolver containers.
	ResolverDefaultDataPath = "/workspace/data"
	// ResolverNotifyDir is name of the notify directory where coordinator would create ok file there.
	ResolverNotifyDir = "notify"
	// ResolverNotifyDirPath is notify directory path in resource resolver container.
	ResolverNotifyDirPath = "/workspace/notify"

	// ResourcePullCommand indicates pull resource
	ResourcePullCommand = "pull"
	// ResourcePushCommand indicates push resource
	ResourcePushCommand = "push"

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
	// DockerInDockerSockVolume is volume used for docker-in-docker to share it's sock file with other containers.
	DockerInDockerSockVolume = "docker-dind-sock"
	// DockerConfigJSONVolume is volume for config.json in secret.
	DockerConfigJSONVolume = "cyclone-docker-secret-volume"

	// DockerSockPath is path of docker socket file in container
	DockerSockPath = "/var/run"

	// DockerConfigPath is path of docker config
	DockerConfigPath = "/root/.docker"
	// DockerConfigJSONFile is name of docker config file
	DockerConfigJSONFile = "config.json"

	// GCContainerName is name of GC container
	GCContainerName = "gc"
	// GCDataPath is parent folder holding data to be cleaned by GC pod
	GCDataPath = "/workspace"

	// StageMountPath is path that we will mount PV on in container.
	StageMountPath = "/__cyclone__workspace"

	// CoordinatorWorkspacePath is path of artifacts in coordinator container
	CoordinatorWorkspacePath = "/workspace/"
)

// ContainerState represents container state.
type ContainerState string

const (
	// ContainerStateTerminated represents container is stopped.
	ContainerStateTerminated ContainerState = "Terminated"
	// ContainerStateInitialized represents container is Running or Stopped, not Init or Creating.
	ContainerStateInitialized ContainerState = "Initialized"
)
