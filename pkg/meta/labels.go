package meta

import (
	"fmt"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// LabelTenantName is the label key used to indicate the tenant which the resources belongs to
	LabelTenantName = "tenant.cyclone.io/name"

	// LabelProjectName is the label key used to indicate the project which the resources belongs to
	LabelProjectName = "project.cyclone.io/name"

	// LabelWorkflowName is the label key used to indicate the workflow which the resources belongs to
	LabelWorkflowName = "workflow.cyclone.io/name"

	// LabelWorkflowRunName is the label key used to indicate the workflowrun which the resources belongs to
	LabelWorkflowRunName = "workflowrun.cyclone.io/name"

	// LabelAcceleration is the label key used to indicate a workflowrun turned on acceleration
	LabelAcceleration = "workflowrun.cyclone.io/acceleration"

	// LabelStageTemplate is the label key used to represent a stage is a stage template
	LabelStageTemplate = "stage.cyclone.io/template"

	// LabelIntegrationType is the label key used to indicate type of integration
	LabelIntegrationType = "integration.cyclone.io/type"

	// LabelIntegrationClusterSchedulable is the label key used to indicate the cluster is schedulable for workflowruns in this tenant
	LabelIntegrationClusterSchedulable = "integration.cyclone.io/cluster-schedulable"

	// LabelBuiltin is the label key used to represent cyclone built in resources
	LabelBuiltin = "cyclone.io/builtin"

	// LabelScene is the label key used to indicate cyclone scenario
	LabelScene = "cyclone.io/scene"

	// TrueValue is the label value used to represent true
	TrueValue = "true"

	// FalseValue is the label value used to represent false
	FalseValue = "false"
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
	return fmt.Sprintf("%s=%s", LabelIntegrationClusterSchedulable, TrueValue)
}

// AddSchedulableClusterLabel adds schedulable label for integrated cluster to run workload.
func AddSchedulableClusterLabel(labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[LabelIntegrationClusterSchedulable] = TrueValue
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
