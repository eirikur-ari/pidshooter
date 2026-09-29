package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/testutil/helper"
)

func TestRequestValidateSpeedTooLow(t *testing.T) {
	err := Request{Game: GameRequest{Speed: helper.Ptr(movement.MinSpeed - 0.1), TimeLimit: helper.Ptr(30)}}.validate()
	assert.Error(t, err)
}

func TestRequestValidateSpeedTooHigh(t *testing.T) {
	err := Request{Game: GameRequest{Speed: helper.Ptr(movement.MaxSpeed + 0.1), TimeLimit: helper.Ptr(30)}}.validate()
	assert.Error(t, err)
}

func TestRequestValidateSpeedMinBoundary(t *testing.T) {
	err := Request{Game: GameRequest{Speed: helper.Ptr(movement.MinSpeed), TimeLimit: helper.Ptr(30)}}.validate()
	assert.NoError(t, err)
}

func TestRequestValidateSpeedMaxBoundary(t *testing.T) {
	err := Request{Game: GameRequest{Speed: helper.Ptr(movement.MaxSpeed), TimeLimit: helper.Ptr(30)}}.validate()
	assert.NoError(t, err)
}

func TestRequestValidateNegativeTimeLimit(t *testing.T) {
	err := Request{Game: GameRequest{Speed: helper.Ptr(2.0), TimeLimit: helper.Ptr(-1)}}.validate()
	assert.Error(t, err)
}

func TestRequestValidateZeroTimeLimitIsUnlimited(t *testing.T) {
	err := Request{Game: GameRequest{Speed: helper.Ptr(2.0), TimeLimit: helper.Ptr(0)}}.validate()
	assert.NoError(t, err)
}

func TestRequestValidateUnsetSpeedAndTimeLimitAreNotValidated(t *testing.T) {
	err := Request{}.validate()
	assert.NoError(t, err)
}

func TestRequestValidateValid(t *testing.T) {
	req := Request{Game: GameRequest{Speed: helper.Ptr(2.0), TimeLimit: helper.Ptr(30)}}
	assert.NoError(t, req.validate())
}
