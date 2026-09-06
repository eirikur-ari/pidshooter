package process

import (
	"fmt"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// Service discovers and terminates OS processes.
type Service struct {
	processMgr outbound.ProcessManager
}

// NewService constructs a Service with all required outbound ports injected.
func NewService(processMgr outbound.ProcessManager) *Service {
	return &Service{processMgr: processMgr}
}

// FindProcesses discovers running processes matching patterns.
func (s *Service) FindProcesses(patterns []string) ([]process.Info, error) {
	processes, err := s.processMgr.List()
	if err != nil {
		return nil, apperror.NewError(apperror.CodeProcessDiscoveryFailed, apperror.SeverityFatal, "process discovery failed", err)
	}

	matches := process.Find(toProcessInfos(processes), patterns, s.processMgr.OwnPID())
	if err := process.ValidateProcesses(matches); err != nil {
		return nil, apperror.NewError(apperror.CodeNoProcessesFound, apperror.SeverityFatal, "", err)
	}

	fmt.Printf("Found %d process(es) matching %v. Starting game...\n", len(matches), patterns)

	return matches, nil
}

// Kill re-verifies target immediately before terminating it, since the PID may
// have been recycled by the OS to a different process — or exited entirely —
// in the time between discovery and the player confirming the kill. shouldReap
// reports whether the caller should reap the target instead of leaving it
// stuck as alive, because its backing process was already gone.
func (s *Service) Kill(target *game.Target) (killed, shouldReap bool, err error) {
	pid := target.PID
	if target.Info.IsProtected() {
		return false, false, fmt.Errorf("refusing to kill PID %d", pid)
	}
	name, err := s.processMgr.LookupName(pid)
	if err != nil {
		return false, true, fmt.Errorf("could not verify PID %d: %w", pid, err)
	}
	if err := process.ValidateName(target.Name, name); err != nil {
		return false, true, err
	}

	killed, err = s.processMgr.Kill(pid)
	return killed, false, err
}
