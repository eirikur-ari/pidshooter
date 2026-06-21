package main

import (
	"testing"

	"github.com/eirikur-ari/pidshooter/internal/process"
)

func TestParseArgs_BasicPattern(t *testing.T) {
	patterns, confirm, speed, timeLimit, err := parseArgs([]string{"firefox"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(patterns) != 1 || patterns[0] != "firefox" {
		t.Errorf("expected patterns=[firefox], got %v", patterns)
	}
	if confirm {
		t.Error("expected confirm=false")
	}
	if speed != 2.0 {
		t.Errorf("expected speed=2.0, got %f", speed)
	}
	if timeLimit != 30 {
		t.Errorf("expected timeLimit=30, got %d", timeLimit)
	}
}

func TestParseArgs_MultiplePatterns(t *testing.T) {
	patterns, _, _, _, err := parseArgs([]string{"chrome", "firefox", "node"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(patterns) != 3 {
		t.Fatalf("expected 3 patterns, got %d", len(patterns))
	}
	expected := []string{"chrome", "firefox", "node"}
	for i, p := range patterns {
		if p != expected[i] {
			t.Errorf("pattern[%d]: expected %q, got %q", i, expected[i], p)
		}
	}
}

func TestParseArgs_ConfirmFlag(t *testing.T) {
	_, confirm, _, _, err := parseArgs([]string{"sleep", "--confirm"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirm {
		t.Error("expected confirm=true")
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
			_, _, speed, _, err := parseArgs([]string{"proc", tt.arg})
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if speed != tt.want {
					t.Errorf("expected speed=%f, got %f", tt.want, speed)
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
			_, _, _, timeLimit, err := parseArgs([]string{"proc", tt.arg})
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if timeLimit != tt.want {
					t.Errorf("expected time=%d, got %d", tt.want, timeLimit)
				}
			}
		})
	}
}

func TestParseArgs_UnknownFlag(t *testing.T) {
	_, _, _, _, err := parseArgs([]string{"proc", "--unknown"})
	if err == nil {
		t.Error("expected error for unknown flag")
	}
}

func TestParseArgs_NoPatterns(t *testing.T) {
	_, _, _, _, err := parseArgs([]string{"--confirm"})
	if err == nil {
		t.Error("expected error when no patterns provided")
	}
}

func TestParseArgs_PatternTooLong(t *testing.T) {
	longPattern := make([]byte, process.MaxPatternLength+1)
	for i := range longPattern {
		longPattern[i] = 'a'
	}
	_, _, _, _, err := parseArgs([]string{string(longPattern)})
	if err == nil {
		t.Error("expected error for pattern exceeding max length")
	}
}
