package process

import (
	"errors"
	"fmt"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// Service discovers and terminates OS processes.
type Service struct {
	proc outbound.Process
}

// NewService constructs a Service with all required outbound ports injected.
func NewService(proc outbound.Process) *Service {
	return &Service{proc: proc}
}

// FindProcesses discovers running processes matching patterns. includeRoot
// additionally permits root-owned processes as matches, regardless of the
// caller's own effective UID.
func (s *Service) FindProcesses(patterns []string, includeRoot bool) ([]process.Info, error) {
	processes, err := s.proc.Discover()
	if err != nil {
		return nil, apperror.NewError(apperror.CodeProcessDiscoveryFailed, apperror.SeverityFatal, "process discovery failed", err)
	}

	matches := process.Find(toProcessInfos(processes), patterns, s.proc.OwnPID(), s.proc.OwnUID(), includeRoot)
	if err := process.ValidateProcesses(matches); err != nil {
		return nil, apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityFatal, "", err)
	}

	fmt.Printf("Found %d process(es) matching %v. Starting game...\n", len(matches), patterns)

	return matches, nil
}

// Kill pins pid, re-verifies its name, then kills it through that same
// pinned reference — so a PID recycled by the OS to a different process
// between verification and kill cannot be silently signaled in the
// original's place. shouldReap reports whether the caller should treat the
// process as already gone rather than as a failed kill.
func (s *Service) Kill(pid int, procName string, protected bool) (shouldReap bool, err error) {
	if protected {
		return false, apperror.NewError(apperror.CodeProcessDiscoveryFailed, apperror.SeverityWarning, fmt.Sprintf("refusing to kill PID %d", pid), nil)
	}
	handle, err := s.proc.Pin(pid)
	if err != nil {
		return true, apperror.NewError(apperror.CodeProcessDiscoveryFailed, apperror.SeverityWarning, fmt.Sprintf("could not pin PID %d", pid), err)
	}
	defer func() { _ = handle.Release() }()

	currentName, err := s.proc.LookupName(pid)
	if err != nil {
		if errors.As(err, &outbound.NotFoundError{}) {
			return true, apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityWarning, fmt.Sprintf("PID %d already exited", pid), err)
		}
		return false, apperror.NewError(apperror.CodeProcessDiscoveryFailed, apperror.SeverityWarning, fmt.Sprintf("could not verify PID %d", pid), err)
	}
	if err := process.ValidateName(procName, currentName); err != nil {
		return true, apperror.NewError(apperror.CodeProcessDiscoveryFailed, apperror.SeverityWarning, "", err)
	}

	err = handle.Kill()
	if errors.As(err, &outbound.NotFoundError{}) {
		return true, apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityWarning, fmt.Sprintf("PID %d already exited", pid), err)
	}
	if err != nil {
		return false, apperror.NewError(apperror.CodeKillFailed, apperror.SeverityWarning, fmt.Sprintf("failed to kill PID %d", pid), err)
	}
	return false, nil
}
