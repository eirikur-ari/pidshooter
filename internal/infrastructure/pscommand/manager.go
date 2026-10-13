package pscommand

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// host is the machine the program runs on.
type host interface {
	// run returns the output of ps run with the given arguments. It returns an
	// outbound.NotFoundError when ps matched no process, and otherwise an error
	// that includes ps's own error text.
	run(args ...string) ([]byte, error)
	// find returns the process with the given pid.
	find(pid int) (osProcess, error)
	// pid returns the PID of the calling process.
	pid() int
	// uid returns the effective user ID of the calling process.
	uid() int
	// goos returns the operating system the program runs on.
	goos() string
}

// manager gives access to the processes running on the local operating
// system, seeing them through the ps command.
type manager struct {
	host host
}

// NewManager returns an outbound.ProcessManager for the operating
// system's processes. It fails if the ps command is unavailable.
func NewManager() (outbound.ProcessManager, error) {
	path, err := exec.LookPath("ps")
	if err != nil {
		return nil, fmt.Errorf("ps not found: %w", err)
	}
	return &manager{host: &runner{path: path, timeout: 5 * time.Second}}, nil
}

// Discover returns the processes currently running.
func (p *manager) Discover() ([]outbound.ProcessInfo, error) {
	collectFailed := func(err error) ([]outbound.ProcessInfo, error) {
		return nil, fmt.Errorf("failed to collect processes: %w", err)
	}

	output, err := p.host.run("-eo", processColumnNames(p.host.goos()))
	if err != nil {
		return collectFailed(err)
	}
	processes, err := parseProcesses(output)
	if err != nil {
		return collectFailed(err)
	}
	return processes, nil
}

// OwnPID returns the PID of the calling process.
func (p *manager) OwnPID() int {
	return p.host.pid()
}

// OwnUID returns the effective UID of the calling process.
func (p *manager) OwnUID() int {
	return p.host.uid()
}

// LookupName returns the current name of the process with the given pid. A
// process that no longer exists, or has become a zombie, is reported as
// outbound.NotFoundError.
func (p *manager) LookupName(pid int) (string, error) {
	lookupFailed := func(err error) (string, error) {
		return "", fmt.Errorf("ps lookup for PID %d failed: %w", pid, err)
	}

	out, err := p.host.run("-p", strconv.Itoa(pid), "-o", processColumnNames(p.host.goos()))
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

// Pin returns an outbound.ProcessHandle bound to the process with the given
// pid. A pid with no process is reported as outbound.NotFoundError.
func (p *manager) Pin(pid int) (outbound.ProcessHandle, error) {
	process, err := p.host.find(pid)
	if err != nil {
		return nil, err
	}
	processHandle := &handle{process: process}
	// FindProcess succeeds even for a pid with no process, so verify checks that one exists.
	if err := processHandle.verify(); err != nil {
		return nil, err
	}
	return processHandle, nil
}

// isZombie reports whether state denotes a zombie process.
func isZombie(state string) bool {
	return strings.HasPrefix(state, "Z")
}

// processColumnNames returns the comma-separated process attribute columns to
// query on the given operating system.
func processColumnNames(goos string) string {
	if goos == "darwin" {
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

// isProcessHeader reports whether a line is the header row of ps output.
func isProcessHeader(line string) bool {
	fields := strings.Fields(line)
	return len(fields) > 0 && strings.EqualFold(fields[0], "uid")
}
