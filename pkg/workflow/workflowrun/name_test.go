package workflowrun

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPodName(t *testing.T) {
	name := PodName("wf", "st")
	if !strings.HasPrefix(name, "wf-st-") {
		t.Errorf("pod name expected to have prefix 'wf-st', but got: %s", name)
	}
	assert.Equal(t, 5, len(strings.TrimPrefix(name, "wf-st-")))
}

func TestGCPodName(t *testing.T) {
	assert.Equal(t, GCPodName("wfr"), "wfrgc--wfr")
}

func TestInputContainerName(t *testing.T) {
	assert.Equal(t, "i1", InputContainerName(1))
}

func TestOutputContainerName(t *testing.T) {
	assert.Equal(t, "csc-o1", OutputContainerName(1))
}
