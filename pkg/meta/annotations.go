package meta

const (
	// AnnotationValueFalse is boolean value false for annotation
	AnnotationValueFalse = "false"

	// AnnotationAlias is the annotation key used to indicate the alias of resources
	AnnotationAlias = "cyclone.dev/alias"

	// AnnotationDescription is the annotation key used to describe resources
	AnnotationDescription = "cyclone.dev/description"

	// AnnotationStageName is annotation applied to pod to indicate which stage it related to
	AnnotationStageName = "stage.cyclone.dev/name"

	// AnnotationWorkflowRunName is annotation applied to pod to specify WorkflowRun the pod belongs to
	AnnotationWorkflowRunName = "workflowrun.cyclone.dev/name"

	// AnnotationWorkflowRunTrigger is the annotation key used to indicate the trigger of workflowruns.
	AnnotationWorkflowRunTrigger = "workflowrun.cyclone.dev/trigger"

	// AnnotationWorkflowRunSCMEvent is the annotation key used to indicate the SCM event data to trigger workflowruns.
	AnnotationWorkflowRunSCMEvent = "workflowrun.cyclone.dev/scm-event"

	// AnnotationWorkflowRunPRUpdatedAt is the annotation key used to indicate the time that SCM event gets triggered.
	AnnotationWorkflowRunPRUpdatedAt = "workflowrun.cyclone.dev/scm-pr-updated-at"

	// AnnotationTenantInfo is the annotation key used for namespace to relate tenant information
	AnnotationTenantInfo = "tenant.cyclone.dev/info"

	// AnnotationTenantStorageUsage is annotation to store storage usuage information
	AnnotationTenantStorageUsage = "tenant.cyclone.dev/storage-usage"

	// AnnotationMetaNamespace is annotation applied to pod to specify the namespace where Workflow, WorkflowRun etc belong to.
	AnnotationMetaNamespace = "cyclone.dev/meta-namespace"

	// AnnotationStageResult is annotation to hold execution results (JSON format) of a stage.
	AnnotationStageResult = "stage.cyclone.dev/execution-results"

	// AnnotationIstioInject is annotation to decide whether to inject istio sidecar
	AnnotationIstioInject = "sidecar.istio.io/inject"
)
