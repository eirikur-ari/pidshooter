package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

func TestKill_Run_ReturnsUnimplementedError(t *testing.T) {
	// When
	err := NewKill().Run(inbound.KillRequest{Patterns: []string{"proc"}})

	// Then
	assert.EqualError(t, err, "unimplemented")
}
