package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

type stubService struct{ err error }

func (s *stubService) Play(_ inbound.GamePlayConfig) error { return s.err }

func TestRun_NoArgs_PrintsUsageAndReturnsNil(t *testing.T) {
	c := NewCLI(&stubService{})
	if err := c.Run([]string{}); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestRun_HelpFlag_PrintsUsageAndReturnsNil(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		t.Run(flag, func(t *testing.T) {
			c := NewCLI(&stubService{})
			if err := c.Run([]string{flag}); err != nil {
				t.Errorf("expected nil, got %v", err)
			}
		})
	}
}

func TestParseArgs_NoArgs_ReturnsErrUsage(t *testing.T) {
	_, err := parseArgs([]string{})
	if !errors.Is(err, errUsage) {
		t.Errorf("expected errUsage, got %v", err)
	}
}

func TestParseArgs_HelpFlag_ReturnsErrUsage(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		t.Run(flag, func(t *testing.T) {
			_, err := parseArgs([]string{flag})
			if !errors.Is(err, errUsage) {
				t.Errorf("expected errUsage, got %v", err)
			}
		})
	}
}

func TestParseArgs_BasicPattern(t *testing.T) {
	cfg, err := parseArgs([]string{"firefox"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Patterns) != 1 || cfg.Patterns[0] != "firefox" {
		t.Errorf("expected patterns=[firefox], got %v", cfg.Patterns)
	}
	if cfg.ConfirmMode {
		t.Error("expected ConfirmMode=false")
	}
	if cfg.Speed != 2.0 {
		t.Errorf("expected Speed=2.0, got %f", cfg.Speed)
	}
	if cfg.TimeLimit != 30 {
		t.Errorf("expected TimeLimit=30, got %d", cfg.TimeLimit)
	}
}

func TestParseArgs_MultiplePatterns(t *testing.T) {
	cfg, err := parseArgs([]string{"chrome", "firefox", "node"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Patterns) != 3 {
		t.Fatalf("expected 3 patterns, got %d", len(cfg.Patterns))
	}
	expected := []string{"chrome", "firefox", "node"}
	for i, p := range cfg.Patterns {
		if p != expected[i] {
			t.Errorf("pattern[%d]: expected %q, got %q", i, expected[i], p)
		}
	}
}

func TestParseArgs_ConfirmFlag(t *testing.T) {
	cfg, err := parseArgs([]string{"sleep", "--confirm"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.ConfirmMode {
		t.Error("expected ConfirmMode=true")
	}
}

func TestParseArgs_SpeedFlag(t *testing.T) {
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
			cfg, err := parseArgs([]string{"proc", tt.arg})
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if cfg.Speed != tt.want {
					t.Errorf("expected Speed=%f, got %f", tt.want, cfg.Speed)
				}
			}
		})
	}
}

func TestParseArgs_TimeFlag(t *testing.T) {
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
			cfg, err := parseArgs([]string{"proc", tt.arg})
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if cfg.TimeLimit != tt.want {
					t.Errorf("expected TimeLimit=%d, got %d", tt.want, cfg.TimeLimit)
				}
			}
		})
	}
}

func TestParseArgs_UnknownFlag(t *testing.T) {
	_, err := parseArgs([]string{"proc", "--unknown"})
	if err == nil {
		t.Error("expected error for unknown flag")
	}
}

func TestParseArgs_NoPatterns(t *testing.T) {
	_, err := parseArgs([]string{"--confirm"})
	if err == nil {
		t.Error("expected error when no patterns provided")
	}
}

func TestParseArgs_PatternTooShort(t *testing.T) {
	for _, p := range []string{"a", "ab"} {
		_, err := parseArgs([]string{p})
		if err == nil {
			t.Errorf("expected error for pattern %q shorter than MinPatternLength", p)
		}
	}
}

func TestParseArgs_PatternExactMinLength(t *testing.T) {
	min := strings.Repeat("a", process.MinPatternLength)
	_, err := parseArgs([]string{min})
	if err != nil {
		t.Errorf("expected no error for min-length pattern, got %v", err)
	}
}

func TestParseArgs_PatternTooLong(t *testing.T) {
	longPattern := make([]byte, process.MaxPatternLength+1)
	for i := range longPattern {
		longPattern[i] = 'a'
	}
	_, err := parseArgs([]string{string(longPattern)})
	if err == nil {
		t.Error("expected error for pattern exceeding max length")
	}
}
