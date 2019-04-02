package meta

const (
	// AnnotationAlias is the annotation key used to indicate the alias of resources
	AnnotationAlias = "cyclone.dev/alias"

	// AnnotationDescription is the annotation key used to describe resources
	AnnotationDescription = "cyclone.dev/description"

	// AnnotationOwner is the annotation key used to indicate the owner of resources.
	AnnotationOwner = "cyclone.dev/owner"

	// AnnotationStageName is annotation applied to pod to indicate which stage it related to
	AnnotationStageName = "stage.cyclone.dev/name"

	// AnnotationWorkflowRunName is annotation applied to pod to specify WorkflowRun the pod belongs to
	AnnotationWorkflowRunName = "workflowrun.cyclone.dev/name"

	// AnnotationWorkflowRunTrigger is the annotation key used to indicate the trigger of workflowruns.
	AnnotationWorkflowRunTrigger = "workflowrun.cyclone.dev/trigger"

	// AnnotationTenantInfo is the annotation key used for namespace to relate tenant information
	AnnotationTenantInfo = "tenant.cyclone.dev/info"

	// AnnotationTenantStorageUsage is annotation to store storage usuage information
	AnnotationTenantStorageUsage = "tenant.cyclone.dev/storage-usage"

	// AnnotationMetaNamespace is annotation applied to pod to specify the namespace where Workflow, WorkflowRun etc belong to.
	// TODO(robin) What is better?
	AnnotationMetaNamespace = "cyclone.dev/meta-namespace"

	// AnnotationGCPod is annotation applied to pod to indicate whether the pod is used for GC purpose
	AnnotationGCPod = "gc.cyclone.dev"
)
