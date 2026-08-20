package game

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestConfirmationPendingFalseWhenEmpty(t *testing.T) {
	c := newConfirmation(true)
	assert.False(t, c.Pending())
}

func TestConfirmationRequestConfirmModeHoldsTarget(t *testing.T) {
	c := newConfirmation(true)
	tgt := &Target{Info: process.NewInfo(1, "x", 0)}
	result := c.Request(tgt)
	assert.Nil(t, result)
	assert.True(t, c.Pending())
}

func TestConfirmationRequestPassthroughModeReturnsTarget(t *testing.T) {
	c := newConfirmation(false)
	tgt := &Target{Info: process.NewInfo(1, "x", 0)}
	result := c.Request(tgt)
	assert.Equal(t, tgt, result)
	assert.False(t, c.Pending())
}

func TestConfirmationAcceptReturnsAndClearsTarget(t *testing.T) {
	tgt := &Target{Info: process.NewInfo(1, "x", 0)}
	c := confirmation{target: tgt}
	result := c.Accept()
	assert.Equal(t, tgt, result)
	assert.False(t, c.Pending())
}

func TestConfirmationCancelClearsPending(t *testing.T) {
	tgt := &Target{Info: process.NewInfo(1, "x", 0)}
	c := confirmation{target: tgt}
	c.Cancel()
	assert.False(t, c.Pending())
}
