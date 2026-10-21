package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

func TestKill_Run_ReturnsUnimplementedErrorAfterLoggingItOnce(t *testing.T) {
	// Given
	handler, logger := newHandlerFixture()

	// When
	err := NewKill(handler).Run(inbound.KillRequest{Patterns: []string{"proc"}})

	// Then
	assert.EqualError(t, err, "unimplemented")
	assert.Equal(t, []string{"unimplemented"}, logger.Errored)
}
