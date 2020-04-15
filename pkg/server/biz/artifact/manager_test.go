package artifact

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestScanAndClean(t *testing.T) {
	manager := NewManager("testdata")
	err := manager.scanAndClean(selectArtifact, time.Duration(0))
	assert.Nil(t, err)

	_, err = os.Stat("testdata/tenant1/project1/wf1/wfr1/artifacts/stage1/artifacts.tar")
	assert.NotNil(t, err)
	assert.True(t, os.IsNotExist(err))
}
