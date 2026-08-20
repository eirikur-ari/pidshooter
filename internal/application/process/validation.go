package process

import (
	"fmt"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func validateSearchPatterns(patterns []string) error {
	return process.Validate(patterns)
}

func validateProcessName(expected, actual string) error {
	if actual != expected {
		return fmt.Errorf("PID name mismatch: expected %q, got %q", expected, actual)
	}
	return nil
}
