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
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// process implements outbound.Process using the ps command.
type process struct {
	psPath string
	// timeout bounds how long a single ps invocation may run. A wedged ps
	// (stalled /proc reader, hung container runtime) would otherwise hang
	// Discovery at startup with no feedback, or hang LookupName inside the game
	// loop's killOrReap goroutine, which never reaches its done-channel select.
	timeout time.Duration
}

// NewProcess returns an outbound.Process backed by the OS ps command.
// It resolves the absolute path to ps at construction time so the
// adapter does not depend on $PATH at runtime.
func NewProcess() (outbound.Process, error) {
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

// LookupName returns the current comm name of pid from the OS, using the same
// ps flags and parsing as Discover so truncation and whitespace handling are
// identical between discovery and the kill-time safety recheck. A pid with
// no live info — because it no longer exists, or because it still holds a
// PID slot but has become a zombie that can't be usefully signaled again —
// is reported as NotFoundError rather than as a name.
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

// Pin returns a ProcessHandle to pid, obtained now rather than at kill
// time, so the eventual Kill signals the exact process pinned here even if
// pid is later recycled to a different process.
//
// On Linux 5.3+, os.FindProcess opens a pidfd for pid, which the kernel
// guarantees stays bound to that exact process for the handle's lifetime,
// closing the PID-reuse window entirely. Platforms without a pidfd
// equivalent (notably macOS) fall back to signaling by bare PID, leaving a
// narrow window between the caller's LookupName re-verification and the
// eventual Kill in which pid could theoretically be recycled; this residual
// risk is accepted given how small the window is.
func (p *process) Pin(pid int) (outbound.ProcessHandle, error) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
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

// isZombie reports whether row's process state is a zombie (a PID slot that
// still exists but whose process has already exited and cannot usefully be
// signaled again). ps's STAT/state column always leads with the state
// letter and may carry additional single-letter flags after it (e.g. "ZN"),
// so this checks the leading character rather than an exact match.
func isZombie(state string) bool {
	return strings.HasPrefix(state, "Z")
}

// run executes ps with args, bounded by p.timeout so a wedged ps can't hang
// its caller indefinitely.
func (p *process) run(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, p.psPath, args...).Output()
	if err != nil {
		return nil, withStderr(err)
	}
	return out, nil
}

// withStderr appends ps's own stderr text to err when available, so a
// non-zero exit doesn't leave callers with just an opaque exit status.
func withStderr(err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
	}
	return err
}

// processColumnNames returns the comma-separated process attribute columns
// to query: uid, pid, rss, stat, and a command name. On darwin the command
// name column is ucomm — the kernel-owned name — rather than comm, which
// BSD ps instead derives from the process's own, unbounded, spoofable
// argv[0]; using ucomm keeps the kill-time name recheck honest about which
// binary is actually running. Elsewhere, comm is already kernel-owned, so
// no substitution is needed.
func processColumnNames() string {
	if runtime.GOOS == "darwin" {
		return "uid,pid,rss,stat,ucomm"
	}
	return "uid,pid,rss,stat,comm"
}

// parseProcesses parses uid/pid/rss/stat/name-formatted ps output, shared by
// discover() and LookupName so both apply identical truncation and whitespace
// handling to the same columns. A row with an unparseable uid, pid, or rss
// is skipped entirely.
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

// isProcessHeader reports whether line is the header row ps prints for the
// uid-led columns processColumnNames requests.
func isProcessHeader(line string) bool {
	fields := strings.Fields(line)
	return len(fields) > 0 && strings.EqualFold(fields[0], "uid")
}
