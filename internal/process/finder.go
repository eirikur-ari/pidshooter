// Package process provides process discovery and filtering.
package process

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// MaxPatternLength is the maximum allowed length for a search pattern.
const MaxPatternLength = 256

// Finder retrieves and filters running processes.
type Finder interface {
	// List returns all processes currently visible to the OS.
	List() ([]Info, error)
	// Find returns processes whose name matches any of the given patterns.
	Find(patterns []string) ([]Info, error)
}

type finder struct{}

// New returns a Finder that uses ps to discover processes.
func NewFinder() Finder {
	return finder{}
}

// Find returns all running processes whose name contains any of the given
// patterns (case-insensitive substring match). It excludes the current
// process and PID 1.
func (f finder) Find(patterns []string) ([]Info, error) {
	if err := validate(patterns); err != nil {
		return nil, err
	}

	processes, err := f.List()
	if err != nil {
		return nil, fmt.Errorf("failed to collect processes: %w", err)
	}

	return filter(processes, patterns), nil
}

// List retrieves all running processes via the ps command.
func (finder) List() ([]Info, error) {
	flags := "-eo"
	if runtime.GOOS == "darwin" {
		flags = "-ceo"
	}
	//TODO: we might want to cover nushell requirements, as well review if we need to adjust ps command according to OS
	cmd := exec.Command("ps", flags, "pid,rss,comm")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ps command failed: %w", err)
	}

	var processes []Info
	lines := strings.Split(string(output), "\n")

	for _, line := range lines[1:] { // Skip header
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}

		rssKB, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			rssKB = 0
		}

		rssBytes := rssKB * 1024
		name := strings.Join(fields[2:], " ")

		processes = append(processes, &process{
			pid:  pid,
			name: name,
			rss:  rssBytes,
		})
	}

	return processes, nil
}

// validate checks that patterns are non-empty and within the allowed length.
func validate(patterns []string) error {
	if len(patterns) == 0 {
		return fmt.Errorf("at least one search patterns is required")
	}
	for _, pattern := range patterns {
		if pattern == "" {
			return fmt.Errorf("search patterns must not be empty")
		}
		if len(pattern) > MaxPatternLength {
			return fmt.Errorf("search patterns exceeds maximum length of %d characters", MaxPatternLength)
		}
	}
	return nil
}

// filter returns the subset of processes whose name contains any of the
// given patterns (case-insensitive). It excludes the current process and PID 1.
func filter(processes []Info, patterns []string) []Info {
	var result []Info
	myPID := os.Getpid()
	for _, proc := range processes {
		if proc.Pid() == myPID || proc.Pid() == 1 {
			continue
		}
		for _, pattern := range patterns {
			if strings.Contains(strings.ToLower(proc.Name()), strings.ToLower(pattern)) {
				result = append(result, proc)
				break
			}
		}
	}
	return result
}
