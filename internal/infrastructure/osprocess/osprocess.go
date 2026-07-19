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

// Process implements outbound.Process using the ps command.
type Process struct{ psPath string }

// NewProcess returns an outbound.Process backed by the OS ps command.
// It resolves the absolute path to ps at construction time so the
// adapter does not depend on $PATH at runtime.
func NewProcess() (outbound.Process, error) {
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

func (p *Process) OwnPid() int {
	return os.Getpid()
}

// LookupName returns the current comm name of pid from the OS, using the same
// ps flags as List so truncation and format are consistent.
func (p *Process) LookupName(pid int) (string, error) {
	args := []string{"-p", strconv.Itoa(pid), "-o", "comm="}
	if runtime.GOOS == "darwin" {
		args = []string{"-c", "-p", strconv.Itoa(pid), "-o", "comm="}
	}
	out, err := exec.Command(p.psPath, args...).Output()
	if err != nil {
		return "", fmt.Errorf("ps lookup failed for PID %d: %w", pid, err)
	}
	return strings.TrimSpace(string(out)), nil
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
	//TODO: we might want to cover nushell requirements, as well review if we need to adjust ps command according to OS
	cmd := exec.Command(p.psPath, flags, "pid,rss,comm")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ps command failed: %w", err)
	}

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
			Pid:  pid,
			Name: name,
			Rss:  rssKB * 1024,
		})
	}

	return processes, nil
}
