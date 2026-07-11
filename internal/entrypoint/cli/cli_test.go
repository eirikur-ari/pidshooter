package cli

import (
	"errors"
	"io"
	"strings"
	"testing"

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
	if err := newSilentCLI(&stubService{}).Run([]string{}); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestRun_HelpFlag_PrintsUsageAndReturnsNil(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		t.Run(flag, func(t *testing.T) {
			if err := newSilentCLI(&stubService{}).Run([]string{flag}); err != nil {
				t.Errorf("expected nil, got %v", err)
			}
		})
	}
}

func TestRun_BasicPattern(t *testing.T) {
	svc := &captureService{}
	if err := newSilentCLI(svc).Run([]string{"firefox"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(svc.cfg.Patterns) != 1 || svc.cfg.Patterns[0] != "firefox" {
		t.Errorf("expected patterns=[firefox], got %v", svc.cfg.Patterns)
	}
	if svc.cfg.ConfirmMode {
		t.Error("expected ConfirmMode=false")
	}
	if svc.cfg.Speed != 2.0 {
		t.Errorf("expected Speed=2.0, got %f", svc.cfg.Speed)
	}
	if svc.cfg.TimeLimit != 30 {
		t.Errorf("expected TimeLimit=30, got %d", svc.cfg.TimeLimit)
	}
}

func TestRun_MultiplePatterns(t *testing.T) {
	svc := &captureService{}
	if err := newSilentCLI(svc).Run([]string{"chrome", "firefox", "node"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(svc.cfg.Patterns) != 3 {
		t.Fatalf("expected 3 patterns, got %d", len(svc.cfg.Patterns))
	}
	for i, want := range []string{"chrome", "firefox", "node"} {
		if svc.cfg.Patterns[i] != want {
			t.Errorf("pattern[%d]: expected %q, got %q", i, want, svc.cfg.Patterns[i])
		}
	}
}

func TestRun_ConfirmFlag(t *testing.T) {
	svc := &captureService{}
	if err := newSilentCLI(svc).Run([]string{"sleep", "--confirm"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !svc.cfg.ConfirmMode {
		t.Error("expected ConfirmMode=true")
	}
}

func TestRun_SpeedFlag(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    float64
		wantErr bool
	}{
		{"valid speed", "--speed=3.5", 3.5, false},
		{"min speed", "--speed=0.1", 0.1, false},
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
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if svc.cfg.Speed != tt.want {
					t.Errorf("expected Speed=%f, got %f", tt.want, svc.cfg.Speed)
				}
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
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if svc.cfg.TimeLimit != tt.want {
					t.Errorf("expected TimeLimit=%d, got %d", tt.want, svc.cfg.TimeLimit)
				}
			}
		})
	}
}

func TestRun_UnknownFlag(t *testing.T) {
	if err := newSilentCLI(&stubService{}).Run([]string{"proc", "--unknown"}); err == nil {
		t.Error("expected error for unknown flag")
	}
}

func TestRun_NoPatterns(t *testing.T) {
	err := newSilentCLI(&stubService{}).Run([]string{"--confirm"})
	if err == nil {
		t.Fatal("expected error when no patterns provided")
	}
	if !errors.Is(err, process.ErrNoPatterns) {
		t.Errorf("expected process.ErrNoPatterns, got %v", err)
	}
}

func TestRun_PatternTooShort(t *testing.T) {
	for _, p := range []string{"a", "ab"} {
		if err := newSilentCLI(&stubService{}).Run([]string{p}); err == nil {
			t.Errorf("expected error for pattern %q shorter than MinPatternLength", p)
		}
	}
}

func TestRun_PatternExactMinLength(t *testing.T) {
	min := strings.Repeat("a", process.MinPatternLength)
	if err := newSilentCLI(&captureService{}).Run([]string{min}); err != nil {
		t.Errorf("expected no error for min-length pattern, got %v", err)
	}
}

func TestRun_PatternTooLong(t *testing.T) {
	longPattern := strings.Repeat("a", process.MaxPatternLength+1)
	if err := newSilentCLI(&stubService{}).Run([]string{longPattern}); err == nil {
		t.Error("expected error for pattern exceeding max length")
	}
}
