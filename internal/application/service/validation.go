package service

import (
	"fmt"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func validateSearchPatterns(patterns []string) error {
	if len(patterns) == 0 {
		return process.ErrNoPatterns
	}
	for _, p := range patterns {
		if err := process.Validate(p); err != nil {
			return err
		}
	}
	return nil
}

func validateProcessName(expected, actual string) error {
	if actual != expected {
		return fmt.Errorf("PID name mismatch: expected %q, got %q", expected, actual)
	}
	return nil
}
