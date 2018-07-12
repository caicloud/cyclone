package event

import (
	"testing"

	"github.com/caicloud/cyclone/cmd/worker/options"
	"github.com/caicloud/cyclone/pkg/api"
)

func TestGetWorkerQuota(t *testing.T) {

	workerQuota := api.WorkerQuota{
		LimitsCPU:      "2",
		LimitsMemory:   "4Gi",
		RequestsCPU:    "1",
		RequestsMemory: "2Gi",
	}
	quota := getWorkerQuota(options.DefaultQuota, &workerQuota)

	if quota[options.ResourceLimitsCPU].String() != "2" {
		t.Error("error:", options.ResourceLimitsCPU)
	}

	if quota[options.ResourceLimitsMemory].String() != "4Gi" {
		t.Error("error:", options.ResourceLimitsMemory)
	}

	if quota[options.ResourceRequestsCPU].String() != "1" {
		t.Error("error:", options.ResourceRequestsCPU)
	}

	if quota[options.ResourceRequestsMemory].String() != "2Gi" {
		t.Error("error:", options.ResourceRequestsMemory)
	}
}
