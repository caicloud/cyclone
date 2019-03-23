package accelerator

import (
	"fmt"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	api "github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/common"
)

// Accelerator ...
type Accelerator struct {
	// wfr represents a workflowrun
	wfr *api.WorkflowRun
	// project the wfr belongs to
	project string
}

// NewAccelerator new an accelerator
func NewAccelerator(project string, wfr *api.WorkflowRun) *Accelerator {
	return &Accelerator{
		wfr:     wfr,
		project: project,
	}
}

// Accelerate will check if the workflowrun has label 'workflowrun.cyclone.io/acceleration=true',
// True will mount some volumes into all stages under the related workflow to cache building dependencies.
// volumes including:
// - '/root/.m2'  maven dependency path
// - '/root/.gradle'  gradle dependency path
// - '/root/.npm'  npm dependency path
func (a *Accelerator) Accelerate() {
	if a.wfr.Labels != nil && a.wfr.Labels[common.LabelAcceleration] == common.LabelTrueValue {
		a.wfr.Spec.PresetVolumes = []v1alpha1.PresetVolume{
			{
				Type:      v1alpha1.PresetVolumeTypePV,
				Path:      fmt.Sprintf("%s/%s/m2", common.CachePrefixPath, a.project),
				MountPath: "/root/.m2",
			},
			{
				Type:      v1alpha1.PresetVolumeTypePV,
				Path:      fmt.Sprintf("%s/%s/gradle", common.CachePrefixPath, a.project),
				MountPath: "/root/.gradle",
			},
			{
				Type:      v1alpha1.PresetVolumeTypePV,
				Path:      fmt.Sprintf("%s/%s/npm", common.CachePrefixPath, a.project),
				MountPath: "/root/.npm",
			},
		}
	}
}
