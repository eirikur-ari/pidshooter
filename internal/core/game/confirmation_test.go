package game

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestConfirmation_Pending_FalseWhenEmpty(t *testing.T) {
	c := NewConfirmation(true)
	assert.False(t, c.Pending())
}

func TestConfirmation_Request_ConfirmMode_HoldsTarget(t *testing.T) {
	c := NewConfirmation(true)
	tgt := &Target{Info: process.Info{Pid: 1, Name: "x"}}
	result := c.Request(tgt)
	assert.Nil(t, result)
	assert.True(t, c.Pending())
}

func TestConfirmation_Request_PassthroughMode_ReturnsTarget(t *testing.T) {
	c := NewConfirmation(false)
	tgt := &Target{Info: process.Info{Pid: 1, Name: "x"}}
	result := c.Request(tgt)
	assert.Equal(t, tgt, result)
	assert.False(t, c.Pending())
}

func TestConfirmation_Accept_ReturnsAndClearsTarget(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "x"}}
	c := Confirmation{target: tgt}
	result := c.Accept()
	assert.Equal(t, tgt, result)
	assert.False(t, c.Pending())
}

func TestConfirmation_Cancel_ClearsPending(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "x"}}
	c := Confirmation{target: tgt}
	c.Cancel()
	assert.False(t, c.Pending())
}

func TestConfirmation_ViewState_NilWhenNoPending(t *testing.T) {
	c := NewConfirmation(true)
	assert.Nil(t, c.ViewState())
}

func TestConfirmation_ViewState_ReturnsPIDAndName(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 42, Name: "suspect"}}
	c := Confirmation{target: tgt}
	cs := c.ViewState()
	assert.NotNil(t, cs)
	assert.Equal(t, 42, cs.PID)
	assert.Equal(t, "suspect", cs.Name)
}
