package process

import (
	"fmt"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// Service discovers and terminates OS processes.
type Service struct {
	process outbound.ProcessManager
}

// NewService constructs a Service with all required outbound ports injected.
func NewService(process outbound.ProcessManager) *Service {
	return &Service{process: process}
}

// FindProcesses discovers running processes matching patterns.
func (s *Service) FindProcesses(patterns []string) ([]process.Info, error) {
	if err := validateSearchPatterns(patterns); err != nil {
		return nil, err
	}
	processes, err := s.process.List()
	if err != nil {
		return nil, fmt.Errorf("process search failed: %w", err)
	}

	matches := process.Find(toProcessInfos(processes), patterns, s.process.OwnPid())

	if len(matches) == 0 {
		fmt.Printf("No processes found matching %v\n", patterns)
		return nil, nil
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
	pid := target.Pid
	if target.Info.IsProtected() {
		return false, false, fmt.Errorf("refusing to kill PID %d", pid)
	}
	name, err := s.process.LookupName(pid)
	if err != nil {
		return false, true, fmt.Errorf("could not verify PID %d: %w", pid, err)
	}
	if err := validateProcessName(target.Name, name); err != nil {
		return false, true, err
	}

	killed, err = s.process.Kill(pid)
	return killed, false, err
}
