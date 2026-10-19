package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

func TestList_Run_ReturnsUnimplementedError(t *testing.T) {
	// When
	err := NewList().Run(inbound.ListRequest{Patterns: []string{"proc"}})

	// Then
	assert.EqualError(t, err, "unimplemented")
}
