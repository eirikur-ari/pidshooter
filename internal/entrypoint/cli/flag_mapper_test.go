package cli

import (
	"flag"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlagMapperToRunRequestIncludesPatterns(t *testing.T) {
	mapper, flagSet := newTestFlagMapper()
	require.NoError(t, flagSet.Parse(nil))

	req := mapper.toRunRequest([]string{"chrome", "firefox"})

	assert.Equal(t, []string{"chrome", "firefox"}, req.Patterns)
}

func TestFlagMapperToRunRequestLeavesUnpassedFlagsNil(t *testing.T) {
	mapper, flagSet := newTestFlagMapper()
	require.NoError(t, flagSet.Parse(nil))

	req := mapper.toRunRequest(nil)

	assert.Nil(t, req.Config.Game.ConfirmMode)
	assert.Nil(t, req.Config.Game.Speed)
	assert.Nil(t, req.Config.Game.TimeLimit)
	assert.Nil(t, req.Config.Process.IncludeRoot)
}

func TestFlagMapperToRunRequestSetsConfirmWhenPassed(t *testing.T) {
	mapper, flagSet := newTestFlagMapper()
	require.NoError(t, flagSet.Parse([]string{"--confirm"}))

	req := mapper.toRunRequest(nil)

	require.NotNil(t, req.Config.Game.ConfirmMode)
	assert.True(t, *req.Config.Game.ConfirmMode)
}

func TestFlagMapperToRunRequestSetsSpeedWhenPassed(t *testing.T) {
	mapper, flagSet := newTestFlagMapper()
	require.NoError(t, flagSet.Parse([]string{"--speed=3.5"}))

	req := mapper.toRunRequest(nil)

	require.NotNil(t, req.Config.Game.Speed)
	assert.Equal(t, 3.5, *req.Config.Game.Speed)
}

func TestFlagMapperToRunRequestSetsTimeLimitWhenPassed(t *testing.T) {
	mapper, flagSet := newTestFlagMapper()
	require.NoError(t, flagSet.Parse([]string{"--time=60"}))

	req := mapper.toRunRequest(nil)

	require.NotNil(t, req.Config.Game.TimeLimit)
	assert.Equal(t, 60, *req.Config.Game.TimeLimit)
}

func TestFlagMapperToRunRequestSetsIncludeRootWhenPassed(t *testing.T) {
	mapper, flagSet := newTestFlagMapper()
	require.NoError(t, flagSet.Parse([]string{"--include-root"}))

	req := mapper.toRunRequest(nil)

	require.NotNil(t, req.Config.Process.IncludeRoot)
	assert.True(t, *req.Config.Process.IncludeRoot)
}

func TestFlagMapperToRunRequestSetsAllFlagsWhenAllPassed(t *testing.T) {
	mapper, flagSet := newTestFlagMapper()
	require.NoError(t, flagSet.Parse([]string{"--confirm", "--speed=4.0", "--time=90", "--include-root"}))

	req := mapper.toRunRequest(nil)

	require.NotNil(t, req.Config.Game.ConfirmMode)
	assert.True(t, *req.Config.Game.ConfirmMode)
	require.NotNil(t, req.Config.Game.Speed)
	assert.Equal(t, 4.0, *req.Config.Game.Speed)
	require.NotNil(t, req.Config.Game.TimeLimit)
	assert.Equal(t, 90, *req.Config.Game.TimeLimit)
	require.NotNil(t, req.Config.Process.IncludeRoot)
	assert.True(t, *req.Config.Process.IncludeRoot)
}

// newTestFlagMapper returns a flagMapper backed by a fresh flag.FlagSet, for
// tests to Parse args into before calling toRunRequest.
func newTestFlagMapper() (*flagMapper, *flag.FlagSet) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	return newFlagMapper(flagSet), flagSet
}
