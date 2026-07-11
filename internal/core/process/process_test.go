package process

import (
	"errors"
	"strings"
	"testing"
)

func TestValidate_TooShort(t *testing.T) {
	for _, p := range []string{"", "a", "ab"} {
		if err := Validate(p); err == nil {
			t.Errorf("expected error for pattern %q shorter than MinPatternLength", p)
		}
	}
}

func TestValidate_ExactMinLength(t *testing.T) {
	p := strings.Repeat("a", MinPatternLength)
	if err := Validate(p); err != nil {
		t.Errorf("expected no error for min-length pattern, got %v", err)
	}
}

func TestValidate_Valid(t *testing.T) {
	if err := Validate("firefox"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_ExactMaxLength(t *testing.T) {
	p := strings.Repeat("a", MaxPatternLength)
	if err := Validate(p); err != nil {
		t.Errorf("expected no error for max-length pattern, got %v", err)
	}
}

func TestValidate_TooLong(t *testing.T) {
	p := strings.Repeat("a", MaxPatternLength+1)
	if err := Validate(p); err == nil {
		t.Error("expected error for pattern exceeding MaxPatternLength")
	}
}

func TestErrNoPatterns_IsSentinel(t *testing.T) {
	if !errors.Is(ErrNoPatterns, ErrNoPatterns) {
		t.Error("ErrNoPatterns must satisfy errors.Is against itself")
	}
}
