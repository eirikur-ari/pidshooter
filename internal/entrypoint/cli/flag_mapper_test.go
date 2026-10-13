package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/config"
)

func TestFlagMapper_toConfigOptions_MapsOnlyThePassedFlags(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedOptions config.Options
	}{
		{"no flags", nil, config.Options{}},
		{"confirm", []string{"--confirm"}, config.Options{Game: config.GameOptions{ConfirmMode: new(true)}}},
		{"confirm explicitly false", []string{"--confirm=false"}, config.Options{Game: config.GameOptions{ConfirmMode: new(false)}}},
		{"speed", []string{"--speed=3.5"}, config.Options{Game: config.GameOptions{Speed: new(3.5)}}},
		{"time", []string{"--time=60"}, config.Options{Game: config.GameOptions{TimeLimit: new(60)}}},
		{"include-root", []string{"--include-root"}, config.Options{Process: config.ProcessOptions{IncludeRoot: new(true)}}},
		{"include-root explicitly false", []string{"--include-root=false"}, config.Options{Process: config.ProcessOptions{IncludeRoot: new(false)}}},
		{"i-am-root", []string{"--i-am-root"}, config.Options{Process: config.ProcessOptions{AllowRoot: new(true)}}},
		{
			"all flags",
			[]string{"--confirm", "--speed=4.0", "--time=90", "--include-root", "--i-am-root"},
			config.Options{
				Game:    config.GameOptions{ConfirmMode: new(true), Speed: new(4.0), TimeLimit: new(90)},
				Process: config.ProcessOptions{IncludeRoot: new(true), AllowRoot: new(true)},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			mapper, flagSet := newFlagMapperFixture()
			require.NoError(t, flagSet.Parse(test.args))

			// When
			options := mapper.toConfigOptions()

			// Then
			assert.Equal(t, test.expectedOptions, options)
		})
	}
}

func TestFlagMapper_toConfigOptions_ReturnsIndependentCopyOfConfigValues(t *testing.T) {
	// Given
	mapper, flagSet := newFlagMapperFixture()
	require.NoError(t, flagSet.Parse([]string{"--confirm", "--speed=3.5", "--time=60", "--include-root", "--i-am-root"}))

	// When
	options := mapper.toConfigOptions()

	// Then
	assert.NotSame(t, mapper.confirm, options.Game.ConfirmMode, "the options must hold their own copy of the value")
	assert.NotSame(t, mapper.speed, options.Game.Speed, "the options must hold their own copy of the value")
	assert.NotSame(t, mapper.timeLimit, options.Game.TimeLimit, "the options must hold their own copy of the value")
	assert.NotSame(t, mapper.includeRoot, options.Process.IncludeRoot, "the options must hold their own copy of the value")
	assert.NotSame(t, mapper.allowRoot, options.Process.AllowRoot, "the options must hold their own copy of the value")
}
