package osprocess

import (
	"errors"
	"os"
	"syscall"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// processHandle implements outbound.ProcessHandle around an *os.Process
// obtained by process.Pin. It performs no safety or name verification
// itself — callers must confirm via LookupName that the pid Pin was
// called with still refers to the intended, non-protected process before
// calling Kill.
type processHandle struct{ proc *os.Process }

// Kill sends SIGKILL to the process this handle refers to.
func (h *processHandle) Kill() (bool, error) {
	if err := h.proc.Signal(syscall.SIGKILL); err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			return false, outbound.NotFoundError{}
		}
		return false, err
	}
	return true, nil
}