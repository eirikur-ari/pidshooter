// Package process provides process discovery and filtering.
package process

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ProcessInfo holds information about a running process.
type ProcessInfo struct {
	Pid  int
	Name string
	RSS  int64 // Resident Set Size in bytes
}

// MaxPatternLength is the maximum allowed length for a search pattern.
const MaxPatternLength = 256

// FindProcesses returns all running processes whose name contains any of the
// given patterns (case-insensitive substring match). It filters out the
// current process and PID 1.
func FindProcesses(patterns []string) ([]ProcessInfo, error) {
	err := validate(patterns)
	if err != nil {
		return nil, err
	}

	processes, err := listProcesses()
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %w", err)
	}

	return collectProcesses(processes, patterns), nil
}

// validates search pattern input
func validate(patterns []string) error {
	if len(patterns) == 0 {
		return fmt.Errorf("at least one search patterns is required")
	}
	for _, keyword := range patterns {
		if keyword == "" {
			return fmt.Errorf("search patterns must not be empty")
		}
		if len(keyword) > MaxPatternLength {
			return fmt.Errorf("search patterns exceeds maximum length of %d characters", MaxPatternLength)
		}
	}
	return nil
}

// listProcesses retrieves all running processes using the ps command.
// This approach works on any POSIX-compatible system.
func listProcesses() ([]ProcessInfo, error) {
	// Use ps with POSIX-compatible flags, include RSS (in KB)
	cmd := exec.Command("ps", "-ceo", "pid,rss,comm")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ps command failed: %w", err)
	}

	var processes []ProcessInfo
	lines := strings.Split(string(output), "\n")

	for _, line := range lines[1:] { // Skip header
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split into PID, RSS, and command name
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

		// Convert KB to bytes
		rssBytes := rssKB * 1024

		// Command name is everything after PID and RSS
		name := strings.Join(fields[2:], " ")

		processes = append(processes, ProcessInfo{
			Pid:  pid,
			Name: name,
			RSS:  rssBytes,
		})
	}

	return processes, nil
}

func collectProcesses(processes []ProcessInfo, patterns []string) []ProcessInfo {
	var result []ProcessInfo
	myPID := os.Getpid()
	for _, proc := range processes {
		// Skip own process and PID 1 (init/systemd)
		if proc.Pid == myPID || proc.Pid == 1 {
			continue
		}
		for _, pattern := range patterns {
			if strings.Contains(strings.ToLower(proc.Name), strings.ToLower(pattern)) {
				result = append(result, proc)
				break
			}
		}
	}
	return result
}
