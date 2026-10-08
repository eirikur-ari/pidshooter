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

// process discovers and manages OS processes using the ps command.
type process struct {
	psPath string
	// timeout bounds how long a single ps invocation may run.
	timeout time.Duration
}

// NewProcess returns an outbound.ProcessManager backed by the OS ps command,
// failing if ps is unavailable.
func NewProcess() (outbound.ProcessManager, error) {
	path, err := exec.LookPath("ps")
	if err != nil {
		return nil, fmt.Errorf("ps not found: %w", err)
	}
	return &process{psPath: path, timeout: 5 * time.Second}, nil
}

func (p *process) Discover() ([]outbound.ProcessInfo, error) {
	processes, err := p.discover()
	if err != nil {
		return nil, fmt.Errorf("failed to collect processes: %w", err)
	}
	return processes, nil
}

func (p *process) OwnPID() int {
	return os.Getpid()
}

// OwnUID returns the effective UID of the calling process.
func (p *process) OwnUID() int {
	return os.Geteuid()
}

// LookupName returns the current command name of pid. A pid with no live
// process, because it no longer exists or has become a zombie, is reported as
// NotFoundError rather than a name.
func (p *process) LookupName(pid int) (string, error) {
	lookupFailed := func(err error) (string, error) {
		return "", fmt.Errorf("ps lookup for PID %d failed: %w", pid, err)
	}

	out, err := p.run("-p", strconv.Itoa(pid), "-o", processColumnNames())
	if err != nil {
		return lookupFailed(err)
	}
	processes, err := parseProcesses(out)
	if err != nil {
		return lookupFailed(err)
	}
	for _, proc := range processes {
		if proc.PID == pid && !isZombie(proc.State) {
			return proc.Name, nil
		}
	}
	return lookupFailed(outbound.NotFoundError{})
}

// Pin returns a ProcessHandle bound to pid, so the eventual Kill targets the
// process pinned here. A pid with no process is reported as NotFoundError.
func (p *process) Pin(pid int) (outbound.ProcessHandle, error) {
	// On Linux 5.3+, os.FindProcess opens a pidfd for pid, which the kernel
	// guarantees stays bound to that exact process for the handle's lifetime,
	// closing the PID-reuse window entirely. Platforms without a pidfd
	// equivalent (notably macOS) fall back to signaling by bare PID, leaving a
	// narrow window between the caller's LookupName re-verification and the
	// eventual Kill in which pid could theoretically be recycled; this residual
	// risk is accepted given how small the window is.
	proc, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
	}
	// FindProcess succeeds for a pid with no process, so signal 0, which only
	// checks existence, tells whether there is anything to pin.
	if err := proc.Signal(syscall.Signal(0)); errors.Is(err, os.ErrProcessDone) {
		_ = proc.Release()
		return nil, outbound.NotFoundError{}
	}
	return &processHandle{proc: proc}, nil
}

func (p *process) discover() ([]outbound.ProcessInfo, error) {
	output, err := p.run("-eo", processColumnNames())
	if err != nil {
		return nil, fmt.Errorf("ps command failed: %w", err)
	}
	return parseProcesses(output)
}

// run executes ps with args, failing once p.timeout elapses.
func (p *process) run(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, p.psPath, args...).Output()
	if err != nil {
		return nil, mapError(err)
	}
	return out, nil
}

// isZombie reports whether state denotes a zombie process.
func isZombie(state string) bool {
	return strings.HasPrefix(state, "Z")
}

// mapError maps ps's silent exit 1 (no process matched) to NotFoundError, and
// appends ps's stderr text to any other exit failure.
func mapError(err error) error {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return err
	}
	if len(exitErr.Stderr) == 0 && exitErr.ExitCode() == 1 {
		return outbound.NotFoundError{}
	}
	if len(exitErr.Stderr) > 0 {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
	}
	return err
}

// processColumnNames returns the comma-separated process attribute columns
// to query.
func processColumnNames() string {
	if runtime.GOOS == "darwin" {
		return "uid,pid,rss,stat,ucomm"
	}
	return "uid,pid,rss,stat,comm"
}

// parseProcesses parses ps output into ProcessInfo values, skipping any row it
// cannot parse. It fails if the header row is missing or malformed.
func parseProcesses(output []byte) ([]outbound.ProcessInfo, error) {
	lines := strings.Split(string(output), "\n")
	if !isProcessHeader(lines[0]) {
		return nil, fmt.Errorf("unexpected ps output: missing or malformed header line")
	}

	var processes []outbound.ProcessInfo
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
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
			continue
		}

		state := fields[3]
		name := strings.Join(fields[4:], " ")

		processes = append(processes, outbound.ProcessInfo{
			UID:   uid,
			PID:   pid,
			Rss:   rssKB * 1024,
			State: state,
			Name:  name,
		})
	}

	return processes, nil
}

// isProcessHeader reports whether line is the uid-led header row of ps output.
func isProcessHeader(line string) bool {
	fields := strings.Fields(line)
	return len(fields) > 0 && strings.EqualFold(fields[0], "uid")
}
