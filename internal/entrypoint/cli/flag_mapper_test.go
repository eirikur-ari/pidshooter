package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
)

func TestFlagMapper_toPlayRequest_MapsOnlyThePassedFlags(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedRequest inbound.PlayRequest
	}{
		{"no flags", nil, inbound.PlayRequest{}},
		{"confirm", []string{"--confirm"}, inbound.PlayRequest{Game: inbound.GameOptions{ConfirmMode: new(true)}}},
		{"confirm explicitly false", []string{"--confirm=false"}, inbound.PlayRequest{Game: inbound.GameOptions{ConfirmMode: new(false)}}},
		{"speed", []string{"--speed=3.5"}, inbound.PlayRequest{Game: inbound.GameOptions{Speed: new(3.5)}}},
		{"time", []string{"--time=60"}, inbound.PlayRequest{Game: inbound.GameOptions{TimeLimit: new(60)}}},
		{"include-root", []string{"--include-root"}, inbound.PlayRequest{Process: inbound.ProcessOptions{IncludeRoot: new(true)}}},
		{"include-root explicitly false", []string{"--include-root=false"}, inbound.PlayRequest{Process: inbound.ProcessOptions{IncludeRoot: new(false)}}},
		{"i-am-root", []string{"--i-am-root"}, inbound.PlayRequest{Process: inbound.ProcessOptions{AllowRoot: new(true)}}},
		{
			"all flags",
			[]string{"--confirm", "--speed=4.0", "--time=90", "--include-root", "--i-am-root"},
			inbound.PlayRequest{
				Game:    inbound.GameOptions{ConfirmMode: new(true), Speed: new(4.0), TimeLimit: new(90)},
				Process: inbound.ProcessOptions{IncludeRoot: new(true), AllowRoot: new(true)},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			mapper, flagSet := newFlagMapperFixture()
			require.NoError(t, flagSet.Parse(test.args))

			// When
			request := mapper.toPlayRequest(nil)

			// Then
			assert.Equal(t, test.expectedRequest, request)
		})
	}
}

func TestFlagMapper_toPlayRequest_ReturnsIndependentCopyOfConfigValues(t *testing.T) {
	// Given
	mapper, flagSet := newFlagMapperFixture()
	require.NoError(t, flagSet.Parse([]string{"--confirm", "--speed=3.5", "--time=60", "--include-root", "--i-am-root"}))

	// When
	request := mapper.toPlayRequest(nil)

	// Then
	assert.NotSame(t, mapper.confirm, request.Game.ConfirmMode, "the request must hold their own copy of the value")
	assert.NotSame(t, mapper.speed, request.Game.Speed, "the request must hold their own copy of the value")
	assert.NotSame(t, mapper.timeLimit, request.Game.TimeLimit, "the request must hold their own copy of the value")
	assert.NotSame(t, mapper.includeRoot, request.Process.IncludeRoot, "the request must hold their own copy of the value")
	assert.NotSame(t, mapper.allowRoot, request.Process.AllowRoot, "the request must hold their own copy of the value")
}
