// Package osprocess implements process discovery and process killing via OS APIs.
package osprocess

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	gamedriven "github.com/eirikur-ari/pidshooter/internal/domain/game/ports/driven"
	"github.com/eirikur-ari/pidshooter/internal/domain/process/ports/driven"
)

// proc holds the raw data for a single running process.
type proc struct {
	pid  int
	name string
	rss  int64
}

func (p *proc) Pid() int     { return p.pid }
func (p *proc) Name() string { return p.name }
func (p *proc) Rss() int64   { return p.rss }

// Finder implements driven.Finder using the ps command.
type Finder struct{ psPath string }

// NewFinder returns a driven.Finder backed by the OS ps command.
// It resolves the absolute path to ps at construction time so the
// adapter does not depend on $PATH at runtime.
func NewFinder() (driven.Finder, error) {
	path, err := exec.LookPath("ps")
	if err != nil {
		return nil, fmt.Errorf("ps not found: %w", err)
	}
	return &Finder{psPath: path}, nil
}

func (f *Finder) Find(patterns []string) ([]driven.Info, error) {
	if err := validate(patterns); err != nil {
		return nil, err
	}
	processes, err := f.List()
	if err != nil {
		return nil, fmt.Errorf("failed to collect processes: %w", err)
	}
	return filter(processes, patterns), nil
}

func (f *Finder) List() ([]driven.Info, error) {
	flags := "-eo"
	if runtime.GOOS == "darwin" {
		flags = "-ceo"
	}
	//TODO: we might want to cover nushell requirements, as well review if we need to adjust ps command according to OS
	cmd := exec.Command(f.psPath, flags, "pid,rss,comm")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ps command failed: %w", err)
	}

	var processes []driven.Info
	lines := strings.Split(string(output), "\n")

	for _, line := range lines[1:] { // skip header
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}

		rssKB, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			rssKB = 0
		}

		name := strings.Join(fields[2:], " ")

		processes = append(processes, &proc{
			pid:  pid,
			name: name,
			rss:  rssKB * 1024,
		})
	}

	return processes, nil
}

func validate(patterns []string) error {
	if len(patterns) == 0 {
		return fmt.Errorf("at least one search pattern is required")
	}
	for _, pattern := range patterns {
		if pattern == "" {
			return fmt.Errorf("search patterns must not be empty")
		}
		if len(pattern) > driven.MaxPatternLength {
			return fmt.Errorf("search patterns exceeds maximum length of %d characters", driven.MaxPatternLength)
		}
	}
	return nil
}

func filter(processes []driven.Info, patterns []string) []driven.Info {
	var result []driven.Info
	myPID := os.Getpid()
	for _, p := range processes {
		if p.Pid() == myPID || p.Pid() == 1 {
			continue
		}
		for _, pattern := range patterns {
			if strings.Contains(strings.ToLower(p.Name()), strings.ToLower(pattern)) {
				result = append(result, p)
				break
			}
		}
	}
	return result
}

// Killer implements gamedriven.ProcessKiller by sending SIGKILL via the OS.
type Killer struct{}

// NewKiller returns a gamedriven.ProcessKiller that sends SIGKILL to the target PID.
func NewKiller() gamedriven.ProcessKiller {
	return &Killer{}
}

func (k *Killer) Kill(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Signal(syscall.SIGKILL)
}
