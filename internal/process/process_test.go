package process

import (
	"os"
	"strings"
	"testing"
)

// validate tests

func TestValidate(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		if err := validate([]string{}); err == nil {
			t.Error("expected error for empty slice")
		}
	})

	t.Run("empty term", func(t *testing.T) {
		if err := validate([]string{""}); err == nil {
			t.Error("expected error for empty term")
		}
	})

	t.Run("term too long", func(t *testing.T) {
		long := strings.Repeat("a", MaxPatternLength+1)
		if err := validate([]string{long}); err == nil {
			t.Error("expected error for term exceeding max length")
		}
	})

	t.Run("valid terms", func(t *testing.T) {
		if err := validate([]string{"myapp", "worker"}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// collectProcesses tests

func TestCollectProcesses_MatchesByName(t *testing.T) {
	processes := []ProcessInfo{
		{Pid: 100, Name: "myapp", RSS: 1024},
		{Pid: 200, Name: "worker", RSS: 2048},
	}
	result := collectProcesses(processes, []string{"myapp"})
	if len(result) != 1 || result[0].Pid != 100 {
		t.Errorf("expected to match 'myapp', got %v", result)
	}
}

func TestCollectProcesses_SubstringMatch(t *testing.T) {
	processes := []ProcessInfo{
		{Pid: 100, Name: "myapp-worker", RSS: 1024},
	}
	result := collectProcesses(processes, []string{"app"})
	if len(result) != 1 {
		t.Error("expected substring match")
	}
}

func TestCollectProcesses_CaseInsensitive(t *testing.T) {
	processes := []ProcessInfo{
		{Pid: 100, Name: "MyApp", RSS: 1024},
	}
	result := collectProcesses(processes, []string{"myapp"})
	if len(result) != 1 {
		t.Error("expected case-insensitive match")
	}
}

func TestCollectProcesses_MultipleTerms(t *testing.T) {
	processes := []ProcessInfo{
		{Pid: 100, Name: "myapp", RSS: 1024},
		{Pid: 200, Name: "worker", RSS: 2048},
		{Pid: 300, Name: "other", RSS: 512},
	}
	result := collectProcesses(processes, []string{"myapp", "worker"})
	if len(result) != 2 {
		t.Errorf("expected 2 matches, got %d", len(result))
	}
}

func TestCollectProcesses_NoMatch(t *testing.T) {
	processes := []ProcessInfo{
		{Pid: 100, Name: "myapp", RSS: 1024},
	}
	result := collectProcesses(processes, []string{"worker"})
	if len(result) != 0 {
		t.Errorf("expected no matches, got %v", result)
	}
}

func TestCollectProcesses_ExcludesPID1(t *testing.T) {
	processes := []ProcessInfo{
		{Pid: 1, Name: "init", RSS: 512},
		{Pid: 100, Name: "myapp", RSS: 1024},
	}
	result := collectProcesses(processes, []string{"init", "myapp"})
	for _, p := range result {
		if p.Pid == 1 {
			t.Error("should not include PID 1")
		}
	}
}

func TestCollectProcesses_ExcludesOwnPID(t *testing.T) {
	ownPID := os.Getpid()
	processes := []ProcessInfo{
		{Pid: ownPID, Name: "testprocess", RSS: 1024},
		{Pid: 100, Name: "testprocess", RSS: 2048},
	}
	result := collectProcesses(processes, []string{"testprocess"})
	for _, p := range result {
		if p.Pid == ownPID {
			t.Error("should not include own PID")
		}
	}
	if len(result) != 1 || result[0].Pid != 100 {
		t.Errorf("expected only PID 100, got %v", result)
	}
}

// FindProcesses integration tests

func TestFindProcesses_EmptyTerms(t *testing.T) {
	_, err := FindProcesses([]string{})
	if err == nil {
		t.Error("expected error for empty terms")
	}
}

func TestFindProcesses_ExcludesPID1(t *testing.T) {
	results, err := FindProcesses([]string{"a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range results {
		if p.Pid == 1 {
			t.Error("should not include PID 1")
		}
	}
}
