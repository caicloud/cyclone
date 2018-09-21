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

func TestGetTargetURL(t *testing.T) {
	event := &api.Event{
		Project: &api.Project{
			Name:  "devops-dd-p1",
			Alias: "p1",
		},

		Pipeline: &api.Pipeline{
			Name:        "pipeline1",
			Annotations: map[string]string{"tenant": "devops"},
		},

		PipelineRecord: &api.PipelineRecord{
			ID: "123456",
		},
	}

	template := "http://192.168.19.96:30000/devops/pipeline/{{.Pipeline.Name}}/record/{{.PipelineRecord.ID}}?workspace={{.Project.Name}}&tenant={{index .Pipeline.Annotations \"tenant\"}}"
	url, err := getPipelineRecordWebURL(template, event)
	if err != nil {
		t.Fatalf("expect %v to nil", err)
	}
	expectURL := "http://192.168.19.96:30000/devops/pipeline/pipeline1/record/123456?workspace=devops-dd-p1&tenant=devops"
	if url != expectURL {
		t.Fatalf("expect %v to %v", url, expectURL)
	}

}
