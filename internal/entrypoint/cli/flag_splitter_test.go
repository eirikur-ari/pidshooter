package cli

import (
	"flag"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitAllPatterns(t *testing.T) {
	patterns, flagArgs, err := newFlagSplitter(newTestFlagSet()).split([]string{"chrome", "firefox"})

	require.NoError(t, err)
	assert.Equal(t, []string{"chrome", "firefox"}, patterns)
	assert.Empty(t, flagArgs)
}

func TestSplitPatternThenBoolFlag(t *testing.T) {
	patterns, flagArgs, err := newFlagSplitter(newTestFlagSet()).split([]string{"chrome", "--confirm"})

	require.NoError(t, err)
	assert.Equal(t, []string{"chrome"}, patterns)
	assert.Equal(t, []string{"--confirm"}, flagArgs)
}

func TestSplitBoolFlagThenPattern(t *testing.T) {
	patterns, flagArgs, err := newFlagSplitter(newTestFlagSet()).split([]string{"--confirm", "chrome"})

	require.NoError(t, err)
	assert.Equal(t, []string{"chrome"}, patterns)
	assert.Equal(t, []string{"--confirm"}, flagArgs)
}

func TestSplitValueFlagWithInlineValue(t *testing.T) {
	patterns, flagArgs, err := newFlagSplitter(newTestFlagSet()).split([]string{"--speed=3.5", "chrome"})

	require.NoError(t, err)
	assert.Equal(t, []string{"chrome"}, patterns)
	assert.Equal(t, []string{"--speed=3.5"}, flagArgs)
}

func TestSplitValueFlagWithSeparateValue(t *testing.T) {
	patterns, flagArgs, err := newFlagSplitter(newTestFlagSet()).split([]string{"--speed", "3.5", "chrome"})

	require.NoError(t, err)
	assert.Equal(t, []string{"chrome"}, patterns)
	assert.Equal(t, []string{"--speed", "3.5"}, flagArgs)
}

func TestSplitPatternsAndFlagsFullyInterspersed(t *testing.T) {
	args := []string{"chrome", "--speed", "3.5", "firefox", "--confirm", "node"}

	patterns, flagArgs, err := newFlagSplitter(newTestFlagSet()).split(args)

	require.NoError(t, err)
	assert.Equal(t, []string{"chrome", "firefox", "node"}, patterns)
	assert.Equal(t, []string{"--speed", "3.5", "--confirm"}, flagArgs)
}

func TestSplitUnknownFlag(t *testing.T) {
	_, _, err := newFlagSplitter(newTestFlagSet()).split([]string{"--bogus"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "flag provided but not defined: -bogus")
}

func TestSplitValueFlagMissingItsValue(t *testing.T) {
	_, _, err := newFlagSplitter(newTestFlagSet()).split([]string{"--speed"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "flag needs an argument: -speed")
}

func TestSplitLoneDashIsAPattern(t *testing.T) {
	patterns, flagArgs, err := newFlagSplitter(newTestFlagSet()).split([]string{"-"})

	require.NoError(t, err)
	assert.Equal(t, []string{"-"}, patterns)
	assert.Empty(t, flagArgs)
}

func TestSplitValueFlagFollowedByFlagLikeTokenReportsMissingValue(t *testing.T) {
	_, _, err := newFlagSplitter(newTestFlagSet()).split([]string{"--speed", "--confirm"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "flag needs an argument: -speed")
}

func TestSplitDoubleDashTreatsRemainingArgsAsPatterns(t *testing.T) {
	patterns, flagArgs, err := newFlagSplitter(newTestFlagSet()).split([]string{"chrome", "--", "--confirm", "-suspicious"})

	require.NoError(t, err)
	assert.Equal(t, []string{"chrome", "--confirm", "-suspicious"}, patterns)
	assert.Empty(t, flagArgs)
}

func TestSplitDoubleDashAsLastArg(t *testing.T) {
	patterns, flagArgs, err := newFlagSplitter(newTestFlagSet()).split([]string{"chrome", "--"})

	require.NoError(t, err)
	assert.Equal(t, []string{"chrome"}, patterns)
	assert.Empty(t, flagArgs)
}

func TestSplitEmptyArgs(t *testing.T) {
	patterns, flagArgs, err := newFlagSplitter(newTestFlagSet()).split(nil)

	require.NoError(t, err)
	assert.Empty(t, patterns)
	assert.Empty(t, flagArgs)
}

// newTestFlagSet returns a FlagSet with one bool flag and one non-bool
// flag registered, mirroring the shape play() builds in cli.go.
func newTestFlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Bool("confirm", false, "")
	fs.Float64("speed", 2.0, "")
	return fs
}
