//go:build integration

package process_test

import (
	"os"
	"strings"
	"testing"

	"github.com/eirikur-ari/pidshooter/internal/process"
)

func TestIntegration_ListProcesses_ReturnsResults(t *testing.T) {
	collector := process.NewDefaultCollector()
	processes, err := collector.Collect()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(processes) == 0 {
		t.Error("expected at least one process")
	}
}

func TestIntegration_ListProcesses_ValidFields(t *testing.T) {
	collector := process.NewDefaultCollector()
	processes, err := collector.Collect()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range processes {
		if p.Pid <= 0 {
			t.Errorf("invalid PID: %d", p.Pid)
		}
		if p.Name == "" {
			t.Errorf("empty name for PID %d", p.Pid)
		}
		if p.RSS < 0 {
			t.Errorf("negative RSS for PID %d: %d", p.Pid, p.RSS)
		}
	}
}

func TestIntegration_ListProcesses_ShortProcessNames(t *testing.T) {
	collector := process.NewDefaultCollector()
	processes, err := collector.Collect()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range processes {
		if strings.Contains(p.Name, "/") {
			t.Errorf("PID %d has a full path in name: %q", p.Pid, p.Name)
		}
	}
}

func TestIntegration_FindProcesses_ExcludesOwnPID(t *testing.T) {
	ownPID := os.Getpid()
	results, err := process.FindProcesses([]string{"go"}, process.NewDefaultCollector())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range results {
		if p.Pid == ownPID {
			t.Errorf("own PID %d should be excluded", ownPID)
		}
		if p.Pid == 1 {
			t.Error("PID 1 should be excluded")
		}
	}
}
