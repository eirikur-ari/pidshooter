package process

import (
	"strings"
	"testing"
)

func TestFind_EmptyPatterns(t *testing.T) {
	_, err := Find([]string{})
	if err == nil {
		t.Error("expected error for empty patterns")
	}
}

func TestFind_EmptyString(t *testing.T) {
	_, err := Find([]string{""})
	if err == nil {
		t.Error("expected error for empty string pattern")
	}
}

func TestFind_PatternTooLong(t *testing.T) {
	long := strings.Repeat("a", MaxPatternLength+1)
	_, err := Find([]string{long})
	if err == nil {
		t.Error("expected error for oversized pattern")
	}
}

func TestFind_ValidPattern(t *testing.T) {
	results, err := Find([]string{"process.test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, p := range results {
		if p.PID == 1 {
			t.Error("should not include PID 1")
		}
	}
}

func TestFind_ExcludesOwnPID(t *testing.T) {
	results, err := Find([]string{"go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, p := range results {
		if p.PID == 1 {
			t.Error("should not include PID 1")
		}
		if p.PID <= 0 {
			t.Errorf("invalid PID: %d", p.PID)
		}
	}
}
