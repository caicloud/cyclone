package hook

import (
	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// Manager is an interface described hook management functions
type Manager interface {
	// Register hooks
	Register(tenant string, wft v1alpha1.WorkflowTrigger) error
	// Unregister hooks
	Unregister(tenant string, wft v1alpha1.WorkflowTrigger) error
}

// GetManager ...
func GetManager(typ v1alpha1.TriggerType) (Manager, error) {
	switch typ {
	case v1alpha1.TriggerTypeSCM:
		return getScmManager(), nil
	}

	return nil, cerr.ErrorUnsupported.Error("trigger type", typ)
}
