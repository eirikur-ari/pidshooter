package cli

import (
	"flag"
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/config"
)

func TestFlagMapperToConfigOptionsLeavesUnpassedFlagsNil(t *testing.T) {
	mapper, flagSet := newTestFlagMapper()
	require.NoError(t, flagSet.Parse(nil))

	opts := mapper.toConfigOptions()

	assert.Nil(t, opts.Game.ConfirmMode)
	assert.Nil(t, opts.Game.Speed)
	assert.Nil(t, opts.Game.TimeLimit)
	assert.Nil(t, opts.Process.IncludeRoot)
	assert.Nil(t, opts.Process.AllowRoot)
}

func TestFlagMapperToConfigOptionsSetsAllowRootWhenPassed(t *testing.T) {
	mapper, flagSet := newTestFlagMapper()
	require.NoError(t, flagSet.Parse([]string{"--i-am-root"}))

	opts := mapper.toConfigOptions()

	require.NotNil(t, opts.Process.AllowRoot)
	assert.True(t, *opts.Process.AllowRoot)
}

func TestFlagMapperToConfigOptionsSetsConfirmWhenPassed(t *testing.T) {
	mapper, flagSet := newTestFlagMapper()
	require.NoError(t, flagSet.Parse([]string{"--confirm"}))

	opts := mapper.toConfigOptions()

	require.NotNil(t, opts.Game.ConfirmMode)
	assert.True(t, *opts.Game.ConfirmMode)
}

func TestFlagMapperToConfigOptionsSetsSpeedWhenPassed(t *testing.T) {
	mapper, flagSet := newTestFlagMapper()
	require.NoError(t, flagSet.Parse([]string{"--speed=3.5"}))

	opts := mapper.toConfigOptions()

	require.NotNil(t, opts.Game.Speed)
	assert.Equal(t, 3.5, *opts.Game.Speed)
}

func TestFlagMapperToConfigOptionsSetsTimeLimitWhenPassed(t *testing.T) {
	mapper, flagSet := newTestFlagMapper()
	require.NoError(t, flagSet.Parse([]string{"--time=60"}))

	opts := mapper.toConfigOptions()

	require.NotNil(t, opts.Game.TimeLimit)
	assert.Equal(t, 60, *opts.Game.TimeLimit)
}

func TestFlagMapperToConfigOptionsSetsIncludeRootWhenPassed(t *testing.T) {
	mapper, flagSet := newTestFlagMapper()
	require.NoError(t, flagSet.Parse([]string{"--include-root"}))

	opts := mapper.toConfigOptions()

	require.NotNil(t, opts.Process.IncludeRoot)
	assert.True(t, *opts.Process.IncludeRoot)
}

func TestFlagMapperToConfigOptionsSetsAllFlagsWhenAllPassed(t *testing.T) {
	mapper, flagSet := newTestFlagMapper()
	require.NoError(t, flagSet.Parse([]string{"--confirm", "--speed=4.0", "--time=90", "--include-root", "--i-am-root"}))

	opts := mapper.toConfigOptions()

	require.NotNil(t, opts.Game.ConfirmMode)
	assert.True(t, *opts.Game.ConfirmMode)
	require.NotNil(t, opts.Game.Speed)
	assert.Equal(t, 4.0, *opts.Game.Speed)
	require.NotNil(t, opts.Game.TimeLimit)
	assert.Equal(t, 90, *opts.Game.TimeLimit)
	require.NotNil(t, opts.Process.IncludeRoot)
	assert.True(t, *opts.Process.IncludeRoot)
	require.NotNil(t, opts.Process.AllowRoot)
	assert.True(t, *opts.Process.AllowRoot)
}

func TestFlagMapperToConfigOptionsClonesFlagSetPointers(t *testing.T) {
	mapper, flagSet := newTestFlagMapper()
	require.NoError(t, flagSet.Parse([]string{"--speed=3.5"}))

	opts := mapper.toConfigOptions()

	require.NotNil(t, opts.Game.Speed)
	assert.NotSame(t, mapper.speed, opts.Game.Speed, "the options must not alias the FlagSet's own destination pointer")
}

func TestFlagMapperToConfigOptionsConcurrentCallsDoNotRace(t *testing.T) {
	mapper, flagSet := newTestFlagMapper()
	require.NoError(t, flagSet.Parse([]string{"--speed=3.5"}))

	const goroutines = 20
	results := make([]config.Options, goroutines)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func() {
			defer wg.Done()
			results[i] = mapper.toConfigOptions()
		}()
	}
	wg.Wait()

	for _, opts := range results {
		require.NotNil(t, opts.Game.Speed)
		assert.Equal(t, 3.5, *opts.Game.Speed)
	}
}

// newTestFlagMapper returns a flagMapper backed by a fresh flag.FlagSet, for
// tests to Parse args into before calling toConfigOptions.
func newTestFlagMapper() (*flagMapper, *flag.FlagSet) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	return newFlagMapper(flagSet), flagSet
}
