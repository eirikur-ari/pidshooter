package application

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func newTestRunner() *Runner {
	return NewRunner(&fake.Process{}, &fake.Store{}, &fake.Renderer{}, fake.NewInputSource())
}

func assertFatal(t *testing.T, err error) {
	t.Helper()
	var inErr *inbound.Error
	require.ErrorAs(t, err, &inErr)
	assert.Equal(t, inbound.ErrorSeverityFatal, inErr.Severity)
}

func TestRunnerRunSpeedTooLow(t *testing.T) {
	err := newTestRunner().Run(inbound.Config{Patterns: []string{"proc"}, Speed: movement.MinSpeed - 0.1})
	assertFatal(t, err)
}

func TestRunnerRunSpeedTooHigh(t *testing.T) {
	err := newTestRunner().Run(inbound.Config{Patterns: []string{"proc"}, Speed: movement.MaxSpeed + 0.1})
	assertFatal(t, err)
}

func TestRunnerRunNegativeTimeLimit(t *testing.T) {
	err := newTestRunner().Run(inbound.Config{Patterns: []string{"proc"}, Speed: 2.0, TimeLimit: -1})
	assertFatal(t, err)
}

func TestRunnerRunNoPatterns(t *testing.T) {
	err := newTestRunner().Run(inbound.Config{Speed: 2.0})
	assertFatal(t, err)
	assert.ErrorIs(t, err, process.ErrNoPatterns)
}

func TestRunnerRunPatternTooShort(t *testing.T) {
	for _, p := range []string{"a", "ab"} {
		err := newTestRunner().Run(inbound.Config{Patterns: []string{p}, Speed: 2.0})
		assertFatal(t, err)
	}
}

func TestRunnerRunPatternExactMinLength(t *testing.T) {
	min := strings.Repeat("a", process.MinPatternLength)
	err := newTestRunner().Run(inbound.Config{Patterns: []string{min}, Speed: 2.0})

	var inErr *inbound.Error
	require.ErrorAs(t, err, &inErr)
	assert.Equal(t, inbound.ErrorSeverityWarning, inErr.Severity)
	assert.Contains(t, inErr.Error(), "no processes found")
}

func TestRunnerRunPatternTooLong(t *testing.T) {
	long := strings.Repeat("a", process.MaxPatternLength+1)
	err := newTestRunner().Run(inbound.Config{Patterns: []string{long}, Speed: 2.0})
	assertFatal(t, err)
}
