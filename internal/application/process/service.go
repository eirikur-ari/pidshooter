package process

import (
	"errors"
	"fmt"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// Service discovers and kills processes.
type Service struct {
	manager  outbound.ProcessManager
	reporter outbound.ProcessReporter
	patterns []string
}

// NewService returns a Service that finds processes matching patterns.
func NewService(manager outbound.ProcessManager, reporter outbound.ProcessReporter, patterns []string) *Service {
	return &Service{manager: manager, reporter: reporter, patterns: patterns}
}

// FindRequest holds the root-handling options for finding processes.
type FindRequest struct {
	// IncludeRoot is true to also match processes owned by root.
	IncludeRoot bool
	// AllowRoot is true to permit running as root.
	AllowRoot bool
}

// FindProcesses returns the processes matching the Service's patterns and reports them.
// It returns an error if running as root without request.AllowRoot, if the patterns are invalid,
// if process discovery fails, or if no process matches.
func (s *Service) FindProcesses(request FindRequest) ([]process.Info, error) {
	if err := process.ValidateRoot(s.manager.OwnUID(), request.AllowRoot); err != nil {
		return nil, apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "", err)
	}
	if err := process.ValidatePatterns(s.patterns); err != nil {
		return nil, apperror.NewError(apperror.CodeInvalidConfig, apperror.SeverityFatal, "invalid configuration", err)
	}

	processes, err := s.manager.Discover()
	if err != nil {
		return nil, apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityFatal, "process discovery failed", err)
	}

	matches := process.Find(toInfos(processes), s.patterns, s.manager.OwnPID(), s.manager.OwnUID(), request.IncludeRoot)
	if err := process.ValidateProcesses(matches); err != nil {
		return nil, apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityFatal, "", err)
	}

	s.reporter.Report(toProcessInfos(matches), s.patterns)

	return matches, nil
}

// Kill kills the process with the given pid if it is still named name.
// It returns an error if protected is true or if the process cannot be verified or killed.
// The error has code apperror.CodeProcessNotFound if and only if the process is gone or no longer has that name.
func (s *Service) Kill(pid int, name string, protected bool) error {
	if protected {
		return apperror.NewError(apperror.CodeKillFailed, apperror.SeverityWarning, fmt.Sprintf("refusing to kill PID %d", pid), nil)
	}
	handle, err := s.manager.Pin(pid)
	if errors.As(err, &outbound.NotFoundError{}) {
		return apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityWarning, fmt.Sprintf("PID %d already exited", pid), err)
	}
	if err != nil {
		return apperror.NewError(apperror.CodeKillFailed, apperror.SeverityWarning, fmt.Sprintf("could not pin PID %d", pid), err)
	}
	defer func() { _ = handle.Release() }()

	currentName, err := s.manager.LookupName(pid)
	if err != nil {
		if errors.As(err, &outbound.NotFoundError{}) {
			return apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityWarning, fmt.Sprintf("PID %d already exited", pid), err)
		}
		return apperror.NewError(apperror.CodeKillFailed, apperror.SeverityWarning, fmt.Sprintf("could not verify PID %d", pid), err)
	}
	if err := process.ValidateName(name, currentName); err != nil {
		return apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityWarning, fmt.Sprintf("PID %d now belongs to another process", pid), err)
	}

	err = handle.Kill()
	if errors.As(err, &outbound.NotFoundError{}) {
		return apperror.NewError(apperror.CodeProcessNotFound, apperror.SeverityWarning, fmt.Sprintf("PID %d already exited", pid), err)
	}
	if err != nil {
		return apperror.NewError(apperror.CodeKillFailed, apperror.SeverityWarning, fmt.Sprintf("failed to kill PID %d", pid), err)
	}
	return nil
}
