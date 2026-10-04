package process

import (
	"fmt"
	"strings"
)

// minPatternLength is the minimum allowed length for a process search pattern.
// Shorter patterns match too broadly (e.g. "a" matches most system process
// names) and increase the risk of surfacing critical system processes as kill
// targets.
const minPatternLength = 3

// maxPatternLength is the maximum allowed length for a process search pattern.
const maxPatternLength = 256

// Info holds the basic details of a single process.
type Info struct {
	PID  int
	Name string
	Rss  int64
	UID  int
}

// NewInfo constructs an Info.
func NewInfo(pid int, name string, rss int64, uid int) Info {
	return Info{PID: pid, Name: name, Rss: rss, UID: uid}
}

// IsProtected reports whether this process must never be targeted because it
// is unsafe to kill.
func (i Info) IsProtected() bool { return i.PID <= 1 }

// IsKillableBy reports whether a caller with effective UID ownUID may target
// this process. Root may target any process, everyone else only their own,
// plus root-owned ones when includeRoot is set.
func (i Info) IsKillableBy(ownUID int, includeRoot bool) bool {
	return ownUID == 0 || i.UID == ownUID || (includeRoot && i.UID == 0)
}

// Find returns the processes whose name contains any pattern (case-insensitive),
// excluding ownPID (the caller's own process ID), protected processes, and any
// process the caller, running as ownUID (its effective user ID), is not
// permitted to kill.
func Find(processes []Info, patterns []string, ownPID, ownUID int, includeRoot bool) []Info {
	var result []Info
	for _, process := range processes {
		if process.PID == ownPID || process.IsProtected() || !process.IsKillableBy(ownUID, includeRoot) {
			continue
		}
		for _, pattern := range patterns {
			if strings.Contains(strings.ToLower(process.Name), strings.ToLower(pattern)) {
				result = append(result, process)
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

// ValidateName returns an error if actual does not match expected.
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

// validatePatterns returns an error if patterns is empty.
func validatePatterns(patterns []string) error {
	if len(patterns) == 0 {
		return fmt.Errorf("at least one search pattern is required")
	}
	return nil
}

// validatePatternLength checks each pattern against the length constraints.
func validatePatternLength(patterns []string) error {
	for _, p := range patterns {
		if len(p) < minPatternLength {
			return fmt.Errorf("search pattern %q must be at least %d characters", p, minPatternLength)
		}
		if len(p) > maxPatternLength {
			return fmt.Errorf("search pattern %q exceeds maximum length of %d characters", p, maxPatternLength)
		}
	}
	return nil
}

// ValidateRoot returns an error when ownUID is root and allowRoot is false.
func ValidateRoot(ownUID int, allowRoot bool) error {
	if ownUID == 0 && !allowRoot {
		return fmt.Errorf("refusing to run as root without an explicit override")
	}
	return nil
}
