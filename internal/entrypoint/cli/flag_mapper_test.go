package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/config"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestFlagMapper_toConfigOptions_MapsOnlyThePassedFlags(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedOptions config.Options
	}{
		{"no flags", nil, config.Options{}},
		{"confirm", []string{"--confirm"}, config.Options{Game: config.GameOptions{ConfirmMode: testutil.Pointer(true)}}},
		{"confirm explicitly false", []string{"--confirm=false"}, config.Options{Game: config.GameOptions{ConfirmMode: testutil.Pointer(false)}}},
		{"speed", []string{"--speed=3.5"}, config.Options{Game: config.GameOptions{Speed: testutil.Pointer(3.5)}}},
		{"time", []string{"--time=60"}, config.Options{Game: config.GameOptions{TimeLimit: testutil.Pointer(60)}}},
		{"include-root", []string{"--include-root"}, config.Options{Process: config.ProcessOptions{IncludeRoot: testutil.Pointer(true)}}},
		{"include-root explicitly false", []string{"--include-root=false"}, config.Options{Process: config.ProcessOptions{IncludeRoot: testutil.Pointer(false)}}},
		{"i-am-root", []string{"--i-am-root"}, config.Options{Process: config.ProcessOptions{AllowRoot: testutil.Pointer(true)}}},
		{
			"all flags",
			[]string{"--confirm", "--speed=4.0", "--time=90", "--include-root", "--i-am-root"},
			config.Options{
				Game:    config.GameOptions{ConfirmMode: testutil.Pointer(true), Speed: testutil.Pointer(4.0), TimeLimit: testutil.Pointer(90)},
				Process: config.ProcessOptions{IncludeRoot: testutil.Pointer(true), AllowRoot: testutil.Pointer(true)},
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
