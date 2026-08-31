// Package process defines the process domain value objects.
package process

import (
	"fmt"
	"strings"
)

// MinPatternLength is the minimum allowed length for a process search pattern.
// Single- or double-character patterns match too broadly (e.g. "a" matches
// most system process names) and increase the risk of surfacing critical
// system processes as kill targets.
const MinPatternLength = 3

// MaxPatternLength is the maximum allowed length for a process search pattern.
const MaxPatternLength = 256

// TODO: perhaps rename the source file to info.go, since it only contains the Info struct and related functions
// Info is a snapshot of a single running process captured at discovery time.
type Info struct {
	Pid  int
	Name string
	Rss  int64
}

// NewInfo constructs an Info snapshot.
func NewInfo(pid int, name string, rss int64) Info {
	return Info{Pid: pid, Name: name, Rss: rss}
}

// IsProtected reports whether this process must never be targeted —
// any PID <= 1 (init, PID 0, or a negative PID), all of which are unsafe to kill.
func (i Info) IsProtected() bool { return i.Pid <= 1 }

// Find returns the subset of processes whose name matches any pattern
// (case-insensitive substring match), excluding ownPid and protected PIDs.
func Find(processes []Info, patterns []string, ownPid int) []Info {
	var result []Info
	for _, pr := range processes {
		if pr.Pid == ownPid || pr.IsProtected() {
			continue
		}
		for _, pattern := range patterns {
			if strings.Contains(strings.ToLower(pr.Name), strings.ToLower(pattern)) {
				result = append(result, pr)
				break
			}
		}
	}
	return result
}

// ValidateProcesses returns an error if processes is empty.
func ValidateProcesses(processes []Info) error {
	if len(processes) == 0 {
		return fmt.Errorf("no processes found")
	}
	return nil
}

// ValidateName returns an error if actual does not match expected — used to
// detect PID recycling between discovery and a later re-verification.
func ValidateName(expected, actual string) error {
	if actual != expected {
		return fmt.Errorf("pid name mismatch: expected %q, got %q", expected, actual)
	}
	return nil
}

// ValidatePatterns returns an error if patterns is empty or any pattern violates the length constraints.
func ValidatePatterns(patterns []string) error {
	if err := validatePatterns(patterns); err != nil {
		return err
	}
	return validatePatternLength(patterns)
}

func validatePatterns(patterns []string) error {
	if len(patterns) == 0 {
		return fmt.Errorf("at least one search pattern is required")
	}
	return nil
}

// validatePatternLength checks each pattern against the length constraints.
// Length is measured in bytes; process names are expected to be ASCII.
func validatePatternLength(patterns []string) error {
	for _, p := range patterns {
		if len(p) < MinPatternLength {
			return fmt.Errorf("search pattern %q must be at least %d characters", p, MinPatternLength)
		}
		if len(p) > MaxPatternLength {
			return fmt.Errorf("search pattern %q exceeds maximum length of %d characters", p, MaxPatternLength)
		}
	}
	return nil
}
