// Package process provides process discovery and filtering.
package process

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Info holds information about a running process.
type Info struct {
	PID  int
	Name string
	RSS  int64 // Resident Set Size in bytes
}

// MaxPatternLength is the maximum allowed length for a search pattern.
const MaxPatternLength = 256

// Find returns all running processes whose name contains any of the
// given patterns (case-insensitive substring match). It filters out the
// current process and PID 1.
func Find(patterns []string) ([]Info, error) {
	err := validateSearchPatterns(patterns)
	if err != nil {
		return nil, err
	}

	myPID := os.Getpid()

	lowerPatterns := ToLower(patterns)

	processes, err := list()
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %w", err)
	}

	matches := Matches(processes, myPID, lowerPatterns)

	return matches, nil
}

func Matches(processes []Info, myPID int, lowerPatterns []string) []Info {
	var matches []Info
	for _, p := range processes {
		if p.PID == myPID || p.PID == 1 {
			continue
		}
		lowerName := strings.ToLower(p.Name)
		for _, pat := range lowerPatterns {
			if strings.Contains(lowerName, pat) {
				matches = append(matches, p)
				break
			}
		}
	}
	return matches
}

func ToLower(patterns []string) []string {
	lowerPatterns := make([]string, len(patterns))
	for i, p := range patterns {
		lowerPatterns[i] = strings.ToLower(p)
	}
	return lowerPatterns
}

func validateSearchPatterns(patterns []string) error {
	if len(patterns) == 0 {
		return fmt.Errorf("at least one search pattern is required")
	}
	for _, p := range patterns {
		if p == "" {
			return fmt.Errorf("search pattern must not be empty")
		}
		if len(p) > MaxPatternLength {
			return fmt.Errorf("search pattern exceeds maximum length of %d characters", MaxPatternLength)
		}
	}
	return nil
}

// list retrieves all running processes using the ps command.
func list() ([]Info, error) {
	cmd := exec.Command("ps", "-eo", "pid,rss,comm")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ps command failed: %w", err)
	}

	var processes []Info
	lines := strings.Split(string(output), "\n")

	for _, line := range lines[1:] {
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

		name := strings.Join(fields[2:], " ")

		processes = append(processes, Info{
			PID:  pid,
			Name: name,
			RSS:  rssKB * 1024,
		})
	}

	return processes, nil
}
