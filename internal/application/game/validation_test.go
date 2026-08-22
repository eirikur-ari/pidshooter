package game

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestValidateProcessesEmptySlice(t *testing.T) {
	assert.Error(t, validateProcesses(nil))
}

func TestValidateProcessesNonEmpty(t *testing.T) {
	assert.NoError(t, validateProcesses([]process.Info{process.NewInfo(100, "target", 4096)}))
}