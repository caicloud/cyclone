package pod

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetResourceVolumeName(t *testing.T) {
	assert.Equal(t, "rsc-git", GetResourceVolumeName("git"))
}
