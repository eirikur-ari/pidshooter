package cli

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

type stubService struct{ err error }

func (s *stubService) Play(_ inbound.GamePlayConfig) error { return s.err }

type captureService struct {
	err error
	cfg inbound.GamePlayConfig
}

func (s *captureService) Play(cfg inbound.GamePlayConfig) error {
	s.cfg = cfg
	return s.err
}

func newSilentCLI(svc inbound.GamePlay) *CLI {
	c := NewCLI(svc)
	c.out = io.Discard
	c.errOut = io.Discard
	return c
}

func TestRun_NoArgs_PrintsUsageAndReturnsNil(t *testing.T) {
	assert.NoError(t, newSilentCLI(&stubService{}).Run([]string{}))
}

func TestRun_HelpFlag_PrintsUsageAndReturnsNil(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		t.Run(flag, func(t *testing.T) {
			assert.NoError(t, newSilentCLI(&stubService{}).Run([]string{flag}))
		})
	}
}

func TestRun_BasicPattern(t *testing.T) {
	svc := &captureService{}
	require.NoError(t, newSilentCLI(svc).Run([]string{"firefox"}))
	require.Len(t, svc.cfg.Patterns, 1)
	assert.Equal(t, "firefox", svc.cfg.Patterns[0])
	assert.False(t, svc.cfg.ConfirmMode)
	assert.Equal(t, 2.0, svc.cfg.Speed)
	assert.Equal(t, 30, svc.cfg.TimeLimit)
}

func TestRun_MultiplePatterns(t *testing.T) {
	svc := &captureService{}
	require.NoError(t, newSilentCLI(svc).Run([]string{"chrome", "firefox", "node"}))
	require.Len(t, svc.cfg.Patterns, 3)
	for i, want := range []string{"chrome", "firefox", "node"} {
		assert.Equal(t, want, svc.cfg.Patterns[i])
	}
}

func TestRun_ConfirmFlag(t *testing.T) {
	svc := &captureService{}
	require.NoError(t, newSilentCLI(svc).Run([]string{"sleep", "--confirm"}))
	assert.True(t, svc.cfg.ConfirmMode)
}

func TestRun_SpeedFlag(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    float64
		wantErr bool
	}{
		{"valid speed", "--speed=3.5", 3.5, false},
		{"min speed", "--speed=0.5", 0.5, false},
		{"max speed", "--speed=5.0", 5.0, false},
		{"too low", "--speed=0.0", 0, true},
		{"too high", "--speed=6.0", 0, true},
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

func TestRun_TimeFlag(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    int
		wantErr bool
	}{
		{"valid time", "--time=60", 60, false},
		{"no limit", "--time=0", 0, false},
		{"negative", "--time=-1", 0, true},
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

func TestRun_UnknownFlag(t *testing.T) {
	assert.Error(t, newSilentCLI(&stubService{}).Run([]string{"proc", "--unknown"}))
}

func TestRun_NoPatterns(t *testing.T) {
	err := newSilentCLI(&stubService{}).Run([]string{"--confirm"})
	require.Error(t, err)
	assert.ErrorIs(t, err, process.ErrNoPatterns)
}

func TestRun_PatternTooShort(t *testing.T) {
	for _, p := range []string{"a", "ab"} {
		assert.Error(t, newSilentCLI(&stubService{}).Run([]string{p}), "expected error for pattern %q shorter than MinPatternLength", p)
	}
}

func TestRun_PatternExactMinLength(t *testing.T) {
	min := strings.Repeat("a", process.MinPatternLength)
	assert.NoError(t, newSilentCLI(&captureService{}).Run([]string{min}))
}

func TestRun_PatternTooLong(t *testing.T) {
	longPattern := strings.Repeat("a", process.MaxPatternLength+1)
	assert.Error(t, newSilentCLI(&stubService{}).Run([]string{longPattern}))
}
