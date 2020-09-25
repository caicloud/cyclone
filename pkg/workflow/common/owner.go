package common

import (
	"reflect"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/meta"
)

// ResolveWorkflowName finds workflow name in the workflowRun's filed in order
// - WorkflowRef
// - OwnerReference
// - Label workflow
// if found returns the value, otherwise returns empty.
func ResolveWorkflowName(wfr v1alpha1.WorkflowRun) string {
	workflowKind := reflect.TypeOf(v1alpha1.Workflow{}).Name()
	if wfr.Spec.WorkflowRef != nil && wfr.Spec.WorkflowRef.Kind == workflowKind {
		return wfr.Spec.WorkflowRef.Name
	}

	for _, or := range wfr.OwnerReferences {
		if or.Kind == workflowKind {
			return or.Name
		}
	}

	if wfr.Labels != nil {
		if n, ok := wfr.Labels[meta.LabelWorkflowName]; ok {
			return n
		}
	}
	return ""
}

// ResolveProjectName finds project name in the workflowRun's label,
// if found returns the value, otherwise returns empty.
func ResolveProjectName(wfr v1alpha1.WorkflowRun) string {
	if wfr.Labels != nil {
		if n, ok := wfr.Labels[meta.LabelProjectName]; ok {
			return n
		}
	}
	return ""
}
