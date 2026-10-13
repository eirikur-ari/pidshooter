package pscommand

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// runner is the machine the program runs on.
type runner struct {
	// path is the location of the ps executable.
	path string
	// timeout bounds how long a single ps invocation may run.
	timeout time.Duration
}

// run returns the output of ps run with args, failing once the timeout elapses.
// ps's silent exit status 1, meaning no process matched, is reported as
// outbound.NotFoundError, and ps's own error text is added to any other failure.
func (r *runner) run(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, r.path, args...).Output()
	if err == nil {
		return out, nil
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return nil, err
	}
	return nil, mapError(exitErr.ExitCode(), exitErr.Stderr, err)
}

// find returns the process with the given pid.
func (*runner) find(pid int) (osProcess, error) {
	// On Linux 5.4+, os.FindProcess opens a pidfd for pid, which the kernel
	// guarantees stays bound to that exact process for the handle's lifetime,
	// closing the PID-reuse window entirely. Platforms without a pidfd
	// equivalent (notably macOS) fall back to signaling by bare PID, leaving a
	// narrow window between the caller's LookupName re-verification and the
	// eventual Kill in which pid could theoretically be recycled; this residual
	// risk is accepted given how small the window is.
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
	}
	return process, nil
}

// pid returns the PID of the calling process.
func (*runner) pid() int {
	return os.Getpid()
}

// uid returns the effective user ID of the calling process.
func (*runner) uid() int {
	return os.Geteuid()
}

// goos returns the operating system the program runs on.
func (*runner) goos() string {
	return runtime.GOOS
}

// mapError maps ps's silent exit 1 (no process matched) to NotFoundError, and
// appends ps's stderr text to any other exit failure.
func mapError(exitCode int, stderr []byte, cause error) error {
	if len(stderr) == 0 && exitCode == 1 {
		return outbound.NotFoundError{}
	}
	if len(stderr) > 0 {
		return fmt.Errorf("%w: %s", cause, strings.TrimSpace(string(stderr)))
	}
	return cause
}
