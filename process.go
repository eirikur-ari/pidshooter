package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ProcessInfo holds information about a running process.
type ProcessInfo struct {
	PID  int
	Name string
	RSS  int64 // Resident Set Size in bytes
}

// maxPatternLength is the maximum allowed length for a search pattern.
const maxPatternLength = 256

// FindProcesses returns all running processes whose name contains any of the
// given patterns (case-insensitive substring match). It filters out the current
// process and PID 1.
func FindProcesses(patterns []string) ([]ProcessInfo, error) {
	if len(patterns) == 0 {
		return nil, fmt.Errorf("at least one search pattern is required")
	}
	for _, p := range patterns {
		if p == "" {
			return nil, fmt.Errorf("search pattern must not be empty")
		}
		if len(p) > maxPatternLength {
			return nil, fmt.Errorf("search pattern exceeds maximum length of %d characters", maxPatternLength)
		}
	}

	myPID := os.Getpid()

	// Lowercase all patterns once
	lowerPatterns := make([]string, len(patterns))
	for i, p := range patterns {
		lowerPatterns[i] = strings.ToLower(p)
	}

	processes, err := listProcesses()
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %w", err)
	}

	var matches []ProcessInfo
	for _, p := range processes {
		// Skip own process and PID 1 (init/systemd)
		if p.PID == myPID || p.PID == 1 {
			continue
		}
		lowerName := strings.ToLower(p.Name)
		for _, pat := range lowerPatterns {
			if strings.Contains(lowerName, pat) {
				matches = append(matches, p)
				break // Don't add same process twice
			}
		}
	}

	return matches, nil
}

// listProcesses retrieves all running processes using the ps command.
// This approach works on any POSIX-compatible system.
func listProcesses() ([]ProcessInfo, error) {
	// Use ps with POSIX-compatible flags, include RSS (in KB)
	cmd := exec.Command("ps", "-eo", "pid,rss,comm")
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

		// Command name is everything after PID and RSS
		name := strings.Join(fields[2:], " ")

		processes = append(processes, ProcessInfo{
			PID:  pid,
			Name: name,
			RSS:  rssKB * 1024, // Convert KB to bytes
		})
	}

	return processes, nil
}
