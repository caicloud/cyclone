package common

import (
	"testing"

	"github.com/caicloud/cyclone/pkg/meta"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
)

func TestResolveWorkflowName(t *testing.T) {

	cases := map[string]struct {
		wfr    v1alpha1.WorkflowRun
		expect string
	}{
		"wfrRef": {
			wfr: v1alpha1.WorkflowRun{
				Spec: v1alpha1.WorkflowRunSpec{
					WorkflowRef: &corev1.ObjectReference{
						Kind: "Workflow",
						Name: "w1",
					},
				},
			},
			expect: "w1",
		},
		"wfrOwner": {
			wfr: v1alpha1.WorkflowRun{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{
							Kind: "Workflow",
							Name: "w1",
						}},
				},
			},
			expect: "w1",
		},
		"label": {
			wfr: v1alpha1.WorkflowRun{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						meta.LabelWorkflowName: "w1",
					},
				},
			},
			expect: "w1",
		},
		"none": {
			wfr:    v1alpha1.WorkflowRun{},
			expect: "",
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.expect, ResolveWorkflowName(c.wfr))
	}
}

func TestResolveProjectName(t *testing.T) {

	cases := map[string]struct {
		wfr    v1alpha1.WorkflowRun
		expect string
	}{
		"label": {
			wfr: v1alpha1.WorkflowRun{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						meta.LabelProjectName: "p1",
					},
				},
			},
			expect: "p1",
		},
		"none": {
			wfr:    v1alpha1.WorkflowRun{},
			expect: "",
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.expect, ResolveProjectName(c.wfr))
	}
}
