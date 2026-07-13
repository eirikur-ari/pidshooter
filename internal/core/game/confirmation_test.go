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

func TestConfirmation_Set_ConfirmMode_HoldsTarget(t *testing.T) {
	c := NewConfirmation(true)
	tgt := &Target{Info: process.Info{Pid: 1, Name: "x"}}
	result := c.Set(tgt)
	assert.Nil(t, result)
	assert.True(t, c.Pending())
}

func TestConfirmation_Set_PassthroughMode_ReturnsTarget(t *testing.T) {
	c := NewConfirmation(false)
	tgt := &Target{Info: process.Info{Pid: 1, Name: "x"}}
	result := c.Set(tgt)
	assert.Equal(t, tgt, result)
	assert.False(t, c.Pending())
}

func TestConfirmation_Kill_ReturnsAndClearsTarget(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "x"}}
	c := Confirmation{target: tgt}
	result := c.Kill()
	assert.Equal(t, tgt, result)
	assert.False(t, c.Pending())
}

func TestConfirmation_Clear_CancelsPending(t *testing.T) {
	tgt := &Target{Info: process.Info{Pid: 1, Name: "x"}}
	c := Confirmation{target: tgt}
	c.Clear()
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
