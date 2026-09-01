package cli

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

type stubService struct{ err error }

func (s *stubService) Run(_ inbound.Config) error { return s.err }

type captureService struct {
	err error
	cfg inbound.Config
}

func (s *captureService) Run(cfg inbound.Config) error {
	s.cfg = cfg
	return s.err
}

func newSilentCLI(svc inbound.Runner) *CLI {
	c := NewCLI(svc, &fake.Logger{})
	c.out = io.Discard
	c.errOut = io.Discard
	return c
}

func TestRunNoArgsPrintsUsageAndReturnsNil(t *testing.T) {
	assert.NoError(t, newSilentCLI(&stubService{}).Run([]string{}))
}

func TestRunHelpFlagPrintsUsageAndReturnsNil(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		t.Run(flag, func(t *testing.T) {
			assert.NoError(t, newSilentCLI(&stubService{}).Run([]string{flag}))
		})
	}
}

func TestRunBasicPattern(t *testing.T) {
	svc := &captureService{}
	require.NoError(t, newSilentCLI(svc).Run([]string{"firefox"}))
	require.Len(t, svc.cfg.Patterns, 1)
	assert.Equal(t, "firefox", svc.cfg.Patterns[0])
	assert.False(t, svc.cfg.ConfirmMode)
	assert.Equal(t, 2.0, svc.cfg.Speed)
	assert.Equal(t, 30, svc.cfg.TimeLimit)
}

func TestRunMultiplePatterns(t *testing.T) {
	svc := &captureService{}
	require.NoError(t, newSilentCLI(svc).Run([]string{"chrome", "firefox", "node"}))
	require.Len(t, svc.cfg.Patterns, 3)
	for i, want := range []string{"chrome", "firefox", "node"} {
		assert.Equal(t, want, svc.cfg.Patterns[i])
	}
}

func TestRunConfirmFlag(t *testing.T) {
	svc := &captureService{}
	require.NoError(t, newSilentCLI(svc).Run([]string{"sleep", "--confirm"}))
	assert.True(t, svc.cfg.ConfirmMode)
}

func TestRunSpeedFlag(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    float64
		wantErr bool
	}{
		{"valid speed", "--speed=3.5", 3.5, false},
		{"min speed", "--speed=0.5", 0.5, false},
		{"max speed", "--speed=5.0", 5.0, false},
		{"invalid", "--speed=abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &captureService{}
			err := newSilentCLI(svc).Run([]string{"proc", tt.arg})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, svc.cfg.Speed)
			}
		})
	}
}

func TestRunTimeFlag(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    int
		wantErr bool
	}{
		{"valid time", "--time=60", 60, false},
		{"no limit", "--time=0", 0, false},
		{"invalid", "--time=abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &captureService{}
			err := newSilentCLI(svc).Run([]string{"proc", tt.arg})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, svc.cfg.TimeLimit)
			}
		})
	}
}

func TestRunUnknownFlag(t *testing.T) {
	assert.Error(t, newSilentCLI(&stubService{}).Run([]string{"proc", "--unknown"}))
}

func TestRunNoArgsWithFlagsForwardsToService(t *testing.T) {
	svc := &captureService{}
	require.NoError(t, newSilentCLI(svc).Run([]string{"--confirm"}))
	assert.Empty(t, svc.cfg.Patterns)
	assert.True(t, svc.cfg.ConfirmMode)
}

func TestRunLogsUnrecognizedError(t *testing.T) {
	svc := &stubService{err: errors.New("boom")}
	logger := &fake.Logger{}
	c := newSilentCLI(svc)
	c.logger = logger

	err := c.Run([]string{"proc"})

	require.Error(t, err)
	require.Len(t, logger.Errors, 1)
	assert.Contains(t, logger.Errors[0], "boom")
}

func TestRunDoesNotLogAppError(t *testing.T) {
	fatal := apperror.NewError(apperror.CodeGameFailed, apperror.SeverityFatal, "game session failed", errors.New("boom"))
	svc := &stubService{err: fatal}
	logger := &fake.Logger{}
	c := newSilentCLI(svc)
	c.logger = logger

	err := c.Run([]string{"proc"})

	require.Error(t, err)
	assert.Empty(t, logger.Errors, "an *apperror.Error was already logged by the application layer")
}

func TestRunReturnsErrorOnFatal(t *testing.T) {
	fatal := apperror.NewError(apperror.CodeGameFailed, apperror.SeverityFatal, "game session failed", errors.New("boom"))
	svc := &stubService{err: fatal}

	err := newSilentCLI(svc).Run([]string{"proc"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "game session failed")
}
