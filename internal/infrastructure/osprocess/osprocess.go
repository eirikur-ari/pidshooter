// Package osprocess implements process discovery and process killing via OS APIs.
package osprocess

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// Process implements outbound.ProcessManager using the ps command.
type Process struct{ psPath string }

// NewProcess returns an outbound.ProcessManager backed by the OS ps command.
// It resolves the absolute path to ps at construction time so the
// adapter does not depend on $PATH at runtime.
func NewProcess() (outbound.ProcessManager, error) {
	path, err := exec.LookPath("ps")
	if err != nil {
		return nil, fmt.Errorf("ps not found: %w", err)
	}
	return &Process{psPath: path}, nil
}

func (p *Process) List() ([]outbound.ProcessInfo, error) {
	processes, err := p.list()
	if err != nil {
		return nil, fmt.Errorf("failed to collect processes: %w", err)
	}
	return processes, nil
}

func (p *Process) OwnPID() int {
	return os.Getpid()
}

// LookupName returns the current comm name of pid from the OS, using the same
// ps flags and parsing as List so truncation and whitespace handling are
// identical between discovery and the kill-time safety recheck.
func (p *Process) LookupName(pid int) (string, error) {
	args := []string{"-p", strconv.Itoa(pid), "-o", "pid,rss,comm"}
	if runtime.GOOS == "darwin" {
		args = append([]string{"-c"}, args...)
	}
	out, err := exec.Command(p.psPath, args...).Output()
	if err != nil {
		return "", fmt.Errorf("ps lookup failed for PID %d: %w", pid, err)
	}
	processes, err := parseProcesses(out)
	if err != nil {
		return "", fmt.Errorf("ps lookup failed for PID %d: %w", pid, err)
	}
	for _, proc := range processes {
		if proc.PID == pid {
			return proc.Name, nil
		}
	}
	return "", fmt.Errorf("ps lookup failed for PID %d: process not found", pid)
}

// Kill sends SIGKILL to the process identified by pid. It performs no
// safety or name verification — callers must confirm via LookupName that
// pid still refers to the intended, non-protected process before calling Kill.
func (p *Process) Kill(pid int) (bool, error) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}
	if err := proc.Signal(syscall.SIGKILL); err != nil {
		return false, err
	}
	return true, nil
}

func (p *Process) list() ([]outbound.ProcessInfo, error) {
	flags := "-eo"
	if runtime.GOOS == "darwin" {
		flags = "-ceo"
	}
	cmd := exec.Command(p.psPath, flags, "pid,rss,comm")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ps command failed: %w", err)
	}
	return parseProcesses(output)
}

// parseProcesses parses `pid,rss,comm`-formatted ps output, shared by list()
// and LookupName so both apply identical truncation and whitespace handling
// to the same columns.
func parseProcesses(output []byte) ([]outbound.ProcessInfo, error) {
	var processes []outbound.ProcessInfo
	lines := strings.Split(string(output), "\n")

	for _, line := range lines[1:] { // skip header
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

		processes = append(processes, outbound.ProcessInfo{
			PID:  pid,
			Name: name,
			Rss:  rssKB * 1024,
		})
	}

	return processes, nil
}
