//go:build integration

package osprocess_test

import (
	"os"
	"strings"
	"testing"

	"github.com/eirikur-ari/pidshooter/internal/adapter/driven/osprocess"
)

func TestIntegration_List_ReturnsResults(t *testing.T) {
	processes, err := osprocess.NewFinder().List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(processes) == 0 {
		t.Error("expected at least one process")
	}
}

func TestIntegration_List_ValidFields(t *testing.T) {
	processes, err := osprocess.NewFinder().List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range processes {
		if p.Pid() <= 0 {
			t.Errorf("invalid PID: %d", p.Pid())
		}
		if p.Name() == "" {
			t.Errorf("empty name for PID %d", p.Pid())
		}
		if p.Rss() < 0 {
			t.Errorf("negative RSS for PID %d: %d", p.Pid(), p.Rss())
		}
	}
}

func TestIntegration_List_ShortProcessNames(t *testing.T) {
	processes, err := osprocess.NewFinder().List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range processes {
		if strings.Contains(p.Name(), "/") {
			t.Errorf("PID %d has a full path in name: %q", p.Pid(), p.Name())
		}
	}
}

func TestIntegration_Find_ExcludesOwnPID(t *testing.T) {
	ownPID := os.Getpid()
	results, err := osprocess.NewFinder().Find([]string{"go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range results {
		if p.Pid() == ownPID {
			t.Errorf("own PID %d should be excluded", ownPID)
		}
		if p.Pid() == 1 {
			t.Error("PID 1 should be excluded")
		}
	}
}
