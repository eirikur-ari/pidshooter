// Package process defines the process domain value objects.
package process

import (
	"errors"
	"fmt"
	"strings"
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
type Info interface {
	Pid() int
	Name() string
	Rss() int64
	// IsProtected reports whether this process must never be targeted —
	// any PID <= 1 (init, PID 0, or a negative PID), all of which are unsafe to kill.
	IsProtected() bool
}

type info struct {
	pid  int
	name string
	rss  int64
}

func (i info) Pid() int          { return i.pid }
func (i info) Name() string      { return i.name }
func (i info) Rss() int64        { return i.rss }
func (i info) IsProtected() bool { return i.pid <= 1 }

// NewInfo constructs an Info snapshot.
func NewInfo(pid int, name string, rss int64) Info {
	return info{pid: pid, name: name, rss: rss}
}

// Find returns the subset of processes whose name matches any pattern
// (case-insensitive substring match), excluding ownPid and protected PIDs.
func Find(processes []Info, patterns []string, ownPid int) []Info {
	var result []Info
	for _, pr := range processes {
		if pr.Pid() == ownPid || pr.IsProtected() {
			continue
		}
		for _, pattern := range patterns {
			if strings.Contains(strings.ToLower(pr.Name()), strings.ToLower(pattern)) {
				result = append(result, pr)
				break
			}
		}
	}
	return result
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
