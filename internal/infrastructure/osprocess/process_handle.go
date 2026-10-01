package osprocess

import (
	"errors"
	"os"
	"syscall"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// processHandle wraps a pinned *os.Process. It performs no safety or name
// verification itself — callers must confirm via LookupName that the
// process it refers to is still the intended, non-protected one before
// calling Kill.
type processHandle struct{ proc *os.Process }

// Kill sends SIGKILL to the process this handle refers to.
func (h *processHandle) Kill() error {
	if err := h.proc.Signal(syscall.SIGKILL); err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			return outbound.NotFoundError{}
		}
		return err
	}
	return nil
}

// Release releases the OS resources FindProcess associated with this handle
// (a pidfd on Linux) rather than leaving them to be closed by the garbage
// collector's finalizer.
func (h *processHandle) Release() error {
	return h.proc.Release()
}
