package pscommand

import (
	"errors"
	"os"
	"syscall"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// osProcess is a process of the operating system.
type osProcess interface {
	// Signal sends the given signal to the process.
	Signal(signal os.Signal) error
	// Release frees the resources held for the process.
	Release() error
}

// handle refers to one pinned process.
type handle struct{ process osProcess }

// Kill sends SIGKILL to the process. A process that no longer exists is
// reported as outbound.NotFoundError.
func (h *handle) Kill() error {
	if err := h.process.Signal(syscall.SIGKILL); err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			return outbound.NotFoundError{}
		}
		return err
	}
	return nil
}

// Release frees the resources held by the handle.
func (h *handle) Release() error {
	return h.process.Release()
}

// verify returns outbound.NotFoundError, after releasing the process, when it
// no longer exists.
func (h *handle) verify() error {
	if err := h.process.Signal(syscall.Signal(0)); errors.Is(err, os.ErrProcessDone) {
		_ = h.process.Release()
		return outbound.NotFoundError{}
	}
	return nil
}
