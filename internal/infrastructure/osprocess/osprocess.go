package osprocess

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// Process implements outbound.ProcessManager using the ps command.
type Process struct {
	psPath string
	// timeout bounds how long a single ps invocation may run. A wedged ps
	// (stalled /proc reader, hung container runtime) would otherwise hang
	// List at startup with no feedback, or hang LookupName inside the game
	// loop's killOrReap goroutine, which never reaches its done-channel select.
	timeout time.Duration
}

// NewProcess returns an outbound.ProcessManager backed by the OS ps command.
// It resolves the absolute path to ps at construction time so the
// adapter does not depend on $PATH at runtime.
func NewProcess() (outbound.ProcessManager, error) {
	path, err := exec.LookPath("ps")
	if err != nil {
		return nil, fmt.Errorf("ps not found: %w", err)
	}
	return &Process{psPath: path, timeout: 5 * time.Second}, nil
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

// OwnUID returns the effective UID of the calling process.
func (p *Process) OwnUID() int {
	return os.Geteuid()
}

// LookupName returns the current comm name of pid from the OS, using the same
// ps flags and parsing as List so truncation and whitespace handling are
// identical between discovery and the kill-time safety recheck.
func (p *Process) LookupName(pid int) (string, error) {
	out, err := p.run("-p", strconv.Itoa(pid), "-o", processColumnNames())
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
		if errors.Is(err, os.ErrProcessDone) {
			return false, outbound.NotFoundError{}
		}
		return false, err
	}
	return true, nil
}

func (p *Process) list() ([]outbound.ProcessInfo, error) {
	output, err := p.run("-eo", processColumnNames())
	if err != nil {
		return nil, fmt.Errorf("ps command failed: %w", err)
	}
	return parseProcesses(output)
}

// run executes ps with args, bounded by p.timeout so a wedged ps can't hang
// its caller indefinitely.
func (p *Process) run(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	return exec.CommandContext(ctx, p.psPath, args...).Output()
}

// processColumnNames returns the comma-separated process attribute columns
// to query: uid, pid, rss, and a command name. On darwin the command name
// column is ucomm — the kernel-owned name — rather than comm, which BSD ps
// instead derives from the process's own, unbounded, spoofable argv[0];
// using ucomm keeps the kill-time name recheck honest about which binary is
// actually running. Elsewhere, comm is already kernel-owned, so no
// substitution is needed.
func processColumnNames() string {
	if runtime.GOOS == "darwin" {
		return "uid,pid,rss,ucomm"
	}
	return "uid,pid,rss,comm"
}

// parseProcesses parses uid/pid/rss/name-formatted ps output, shared by
// list() and LookupName so both apply identical truncation and whitespace
// handling to the same columns.
func parseProcesses(output []byte) ([]outbound.ProcessInfo, error) {
	var processes []outbound.ProcessInfo
	lines := strings.Split(string(output), "\n")

	for _, line := range lines[1:] { // skip header
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		uid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}

		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}

		rssKB, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			rssKB = 0
		}

		name := strings.Join(fields[3:], " ")

		processes = append(processes, outbound.ProcessInfo{
			PID:  pid,
			Name: name,
			Rss:  rssKB * 1024,
			UID:  uid,
		})
	}

	return processes, nil
}
