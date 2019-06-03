package accelerator

import (
	"fmt"
	"reflect"
	"testing"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/server/common"
)

func TestAccelerate(t *testing.T) {
	project := "project"
	testcase := map[string]struct {
		wfr    *v1alpha1.WorkflowRun
		expect v1alpha1.WorkflowRun
	}{
		"accelerate": {
			wfr: &v1alpha1.WorkflowRun{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:   "test1",
					Labels: map[string]string{meta.LabelWorkflowRunAcceleration: meta.LabelValueTrue},
				},
			},
			expect: v1alpha1.WorkflowRun{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:   "test1",
					Labels: map[string]string{meta.LabelWorkflowRunAcceleration: meta.LabelValueTrue},
				},
				Spec: v1alpha1.WorkflowRunSpec{
					PresetVolumes: []v1alpha1.PresetVolume{
						{
							Type:      v1alpha1.PresetVolumeTypePVC,
							Path:      fmt.Sprintf("%s/%s/m2", common.CachePrefixPath, project),
							MountPath: "/root/.m2",
						},
						{
							Type:      v1alpha1.PresetVolumeTypePVC,
							Path:      fmt.Sprintf("%s/%s/gradle", common.CachePrefixPath, project),
							MountPath: "/root/.gradle",
						},
						{
							Type:      v1alpha1.PresetVolumeTypePVC,
							Path:      fmt.Sprintf("%s/%s/npm", common.CachePrefixPath, project),
							MountPath: "/root/.npm",
						},
					},
				},
			},
		},
		"no-accelerate": {
			wfr: &v1alpha1.WorkflowRun{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:   "test1",
					Labels: map[string]string{meta.LabelWorkflowRunAcceleration: meta.LabelValueFalse},
				},
			},
			expect: v1alpha1.WorkflowRun{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:   "test1",
					Labels: map[string]string{meta.LabelWorkflowRunAcceleration: meta.LabelValueTrue},
				},
			},
		},
	}

	for k, tc := range testcase {
		NewAccelerator("test", project, tc.wfr).Accelerate()
		if !reflect.DeepEqual(tc.expect.Spec.PresetVolumes, tc.wfr.Spec.PresetVolumes) {
			t.Errorf("test %s failed, expect not equal to return", k)
		}
	}
}
