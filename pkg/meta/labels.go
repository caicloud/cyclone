package meta

import (
	"fmt"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// LabelTenantName is the label key used to indicate the tenant which the resources belongs to
	LabelTenantName = "tenant.cyclone.dev/name"

	// LabelProjectName is the label key used to indicate the project which the resources belongs to
	LabelProjectName = "project.cyclone.dev/name"

	// LabelWorkflowName is the label key used to indicate the workflow which the resources belongs to
	LabelWorkflowName = "workflow.cyclone.dev/name"

	// LabelWorkflowRunName is the label key used to indicate the workflowrun which the resources belongs to
	LabelWorkflowRunName = "workflowrun.cyclone.dev/name"

	// LabelWorkflowRunAcceleration is the label key used to indicate a workflowrun turned on acceleration
	LabelWorkflowRunAcceleration = "workflowrun.cyclone.dev/acceleration"

	// LabelStageTemplate is the label key used to represent a stage is a stage template
	LabelStageTemplate = "stage.cyclone.dev/template"

	// LabelIntegrationType is the label key used to indicate type of integration
	LabelIntegrationType = "integration.cyclone.dev/type"

	// LabelIntegrationSchedulableCluster is the label key used to indicate the cluster is schedulable for workflowruns in this tenant
	LabelIntegrationSchedulableCluster = "integration.cyclone.dev/schedulable-cluster"

	// LabelPodKind is the label key applied to pod to indicate whether the pod is used for GC purpose.
	LabelPodKind = "pod.kubernetes.io/kind"

	// LabelPodCreatedBy is the label key applied to pod to indicate who the pod is created by.
	LabelPodCreatedBy = "pod.kubernetes.io/created-by"

	// LabelBuiltin is the label key used to represent cyclone built in resources
	LabelBuiltin = "cyclone.dev/builtin"

	// LabelScene is the label key used to indicate cyclone scenario
	LabelScene = "cyclone.dev/scene"

	// TrueValue is the label value used to represent true
	TrueValue = "true"

	// FalseValue is the label value used to represent false
	FalseValue = "false"

	// CycloneCreator is the label value used to represent the resources created by Cyclone.
	CycloneCreator = "cyclone"
)

// PodKind represents the type of pods created by Cyclone.
type PodKind string

func (pk PodKind) String() string {
	return string(pk)
}

const (
	// PodKindGC represents the pod is used for GC purpose.
	PodKindGC PodKind = "gc"
)

// ProjectSelector is a selector for cyclone CRD resources which have corresponding project label
func ProjectSelector(project string) string {
	return LabelProjectName + "=" + project
}

// WorkflowSelector is a selector for cyclone CRD resources which have corresponding workflow label
func WorkflowSelector(workflow string) string {
	return LabelWorkflowName + "=" + workflow
}

// SchedulableClusterSelector is a selector for clusters which are use to perform workload
func SchedulableClusterSelector() string {
	return fmt.Sprintf("%s=%s", LabelIntegrationSchedulableCluster, TrueValue)
}

// AddSchedulableClusterLabel adds schedulable label for integrated cluster to run workload.
func AddSchedulableClusterLabel(labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[LabelIntegrationSchedulableCluster] = TrueValue
	return labels
}

// AddStageTemplateLabel adds template label for stages.
func AddStageTemplateLabel(labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[LabelStageTemplate] = TrueValue
	return labels
}

// StageTemplateSelector returns a label selector to query stage templates.
func StageTemplateSelector() string {
	return fmt.Sprintf("%s=%s", LabelStageTemplate, TrueValue)
}

// BuiltinLabelSelector returns a label selector to query cyclone built-in resources.
func BuiltinLabelSelector() string {
	return fmt.Sprintf("%s=%s", LabelBuiltin, TrueValue)
}

// CyclonePodSelector selects pods that are created by Cyclone, for example, stage execution pods, GC pods.
func CyclonePodSelector() string {
	return fmt.Sprintf("%s=%s", LabelPodCreatedBy, CycloneCreator)
}

// LabelExistsSelector returns a label selector to query resources with label key exists.
func LabelExistsSelector(key string) string {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      key,
				Operator: metav1.LabelSelectorOpExists,
			},
		},
	})

	if err != nil {
		log.Errorf("Fail to new label exists selector")
	}
	return selector.String()
}
