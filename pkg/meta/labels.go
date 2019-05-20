package meta

import (
	"fmt"
	"os"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/common"
)

const (
	// LabelControllerInstance is instance name of the workflow controller
	LabelControllerInstance = "controller.cyclone.dev/instance"

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

	// LabelWorkflowRunNotificationSent is the label key used to indicate a workflowrun has been sent as notification
	LabelWorkflowRunNotificationSent = "workflowrun.cyclone.dev/notification-sent"

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

	// LabelValueTrue is the label value used to represent true
	LabelValueTrue = "true"

	// LabelValueFalse is the label value used to represent false
	LabelValueFalse = "false"

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
	return fmt.Sprintf("%s=%s", LabelIntegrationSchedulableCluster, LabelValueTrue)
}

// AddSchedulableClusterLabel adds schedulable label for integrated cluster to run workload.
func AddSchedulableClusterLabel(labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[LabelIntegrationSchedulableCluster] = LabelValueTrue
	return labels
}

// AddNotificationSentLabel adds notification sent label for workflowruns.
func AddNotificationSentLabel(labels map[string]string, sent bool) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}

	if sent {
		labels[LabelWorkflowRunNotificationSent] = LabelValueTrue
	} else {
		labels[LabelWorkflowRunNotificationSent] = LabelValueFalse
	}

	return labels
}

// AddStageTemplateLabel adds template label for stages.
func AddStageTemplateLabel(labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[LabelStageTemplate] = LabelValueTrue
	return labels
}

// StageTemplateSelector returns a label selector to query stage templates.
func StageTemplateSelector() string {
	return fmt.Sprintf("%s=%s", LabelStageTemplate, LabelValueTrue)
}

// BuiltinLabelSelector returns a label selector to query cyclone built-in resources.
func BuiltinLabelSelector() string {
	return fmt.Sprintf("%s=%s", LabelBuiltin, LabelValueTrue)
}

// CyclonePodSelector selects pods that are created by Cyclone (for example, stage execution pods, GC pods)
// and manged by current workflow controller.
func CyclonePodSelector() string {
	instance := os.Getenv(common.ControllerInstanceEnvName)
	if len(instance) == 0 {
		return fmt.Sprintf("%s=%s,!%s", LabelPodCreatedBy, CycloneCreator, LabelControllerInstance)
	}

	return fmt.Sprintf("%s=%s,%s=%s", LabelPodCreatedBy, CycloneCreator, LabelControllerInstance, instance)
}

// WorkflowTriggerSelector selects workflow triggers managed by current controller instance
func WorkflowTriggerSelector() string {
	instance := os.Getenv(common.ControllerInstanceEnvName)
	if len(instance) == 0 {
		return fmt.Sprintf("!%s", LabelControllerInstance)
	}

	return fmt.Sprintf("%s=%s", LabelControllerInstance, instance)
}

// WorkflowRunSelector selects WorkflowRun that managed by current controller instance.
func WorkflowRunSelector() string {
	instance := os.Getenv(common.ControllerInstanceEnvName)
	if len(instance) == 0 {
		return fmt.Sprintf("!%s", LabelControllerInstance)
	}

	return fmt.Sprintf("%s=%s", LabelControllerInstance, instance)
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

// LabelExists checks the existence of expected label, return true if exists, otherwise return false.
func LabelExists(labels map[string]string, expectedLabel string) bool {
	if labels == nil {
		return false
	}

	_, ok := labels[expectedLabel]
	return ok
}
