package accelerator

import (
	"fmt"

	"github.com/caicloud/nirvana/log"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/server/biz/usage"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
)

// CacheSizeLimit is cache size limit, it's percentage of the total PVC size.
const CacheSizeLimit = 0.8

// Accelerator ...
type Accelerator struct {
	// tenant name
	tenant string
	// project the wfr belongs to
	project string
	// wfr represents a workflowrun
	wfr *v1alpha1.WorkflowRun
	// reporter reports PVC usage used for workflow in the tenant
	reporter usage.PVCReporter
}

// NewAccelerator new an accelerator
func NewAccelerator(tenant, project string, wfr *v1alpha1.WorkflowRun) *Accelerator {
	reporter, err := usage.NewPVCReporter(handler.K8sClient, tenant)
	if err != nil {
		log.Warningf("Create pvc reporter for tenant %s error: %v", tenant, err)
	}

	return &Accelerator{
		tenant:   tenant,
		wfr:      wfr,
		project:  project,
		reporter: reporter,
	}
}

// Accelerate will check if the workflowrun has label 'workflowrun.cyclone.dev/acceleration=true',
// True will mount some volumes into all stages under the related workflow to cache building dependencies.
// volumes including:
// - '/root/.m2'  maven dependency path
// - '/root/.gradle'  gradle dependency path
// - '/root/.npm'  npm dependency path
// - '/root/.ivy2' sbt default dependency cache path
// - '/root/.cache' coursier is a dependency resolver, this is coursier default dependency path
func (a *Accelerator) Accelerate() {
	if !a.allowed() {
		return
	}

	if a.wfr.Labels != nil && a.wfr.Labels[meta.LabelWorkflowRunAcceleration] == meta.LabelValueTrue {
		a.wfr.Spec.PresetVolumes = append(a.wfr.Spec.PresetVolumes, []v1alpha1.PresetVolume{
			{
				Type:      v1alpha1.PresetVolumeTypePVC,
				Path:      fmt.Sprintf("%s/%s/m2", common.CachePrefixPath, a.project),
				MountPath: "/root/.m2",
			},
			{
				Type:      v1alpha1.PresetVolumeTypePVC,
				Path:      fmt.Sprintf("%s/%s/gradle", common.CachePrefixPath, a.project),
				MountPath: "/root/.gradle",
			},
			{
				Type:      v1alpha1.PresetVolumeTypePVC,
				Path:      fmt.Sprintf("%s/%s/npm", common.CachePrefixPath, a.project),
				MountPath: "/root/.npm",
			},
			{
				Type:      v1alpha1.PresetVolumeTypePVC,
				Path:      fmt.Sprintf("%s/%s/sbt/cache", common.CachePrefixPath, a.project),
				MountPath: "/root/.cache",
			},
			{
				Type:      v1alpha1.PresetVolumeTypePVC,
				Path:      fmt.Sprintf("%s/%s/sbt/ivy2", common.CachePrefixPath, a.project),
				MountPath: "/root/.ivy2",
			},
		}...)
	}
}

// allowed determines whether it's allowed to open acceleration for the workflow execution. For the moment,
// only PVC storage constraint is enforced.
func (a *Accelerator) allowed() bool {
	if a.reporter == nil {
		return true
	}

	used, err := a.reporter.UsedPercentage("caches")
	if err != nil {
		log.Warningf("Get caches usage error: %v", err)
		return true
	}

	log.Infof("caches used %.2f PVC storage in tenant %s", used, a.tenant)
	if used >= float64(CacheSizeLimit) {
		log.Warningf("caches used %.2f PVC storage, exceeds limit %.2f, will stop acceleration, tenant: %s", used, CacheSizeLimit, a.tenant)
		return false
	}

	return true
}
