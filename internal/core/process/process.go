// Package process defines the process domain value objects.
package process

import (
	"errors"
	"fmt"
)

// ErrNoPatterns is returned when no search patterns are provided.
var ErrNoPatterns = errors.New("at least one search pattern is required")

// MinPatternLength is the minimum allowed length for a process search pattern.
// Single- or double-character patterns match too broadly (e.g. "a" matches
// most system process names) and increase the risk of surfacing critical
// system processes as kill targets.
const MinPatternLength = 3

// MaxPatternLength is the maximum allowed length for a process search pattern.
const MaxPatternLength = 256

// Info is a snapshot of a single running process captured at discovery time.
type Info struct {
	Pid  int
	Name string
	Rss  int64
}

// Validate returns an error if pattern violates the length constraints.
// Length is measured in bytes; process names are expected to be ASCII.
func Validate(pattern string) error {
	if len(pattern) < MinPatternLength {
		return fmt.Errorf("search pattern %q must be at least %d characters", pattern, MinPatternLength)
	}
	if len(pattern) > MaxPatternLength {
		return fmt.Errorf("search pattern %q exceeds maximum length of %d characters", pattern, MaxPatternLength)
	}
	return nil
}
