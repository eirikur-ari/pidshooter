package process

import (
	"os"
	"strings"
	"testing"
)

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

func TestFilter_MatchesByName(t *testing.T) {
	processes := []Info{
		&process{pid: 100, name: "myapp", rss: 1024},
		&process{pid: 200, name: "worker", rss: 2048},
	}
	result := filter(processes, []string{"myapp"})
	if len(result) != 1 || result[0].Pid() != 100 {
		t.Errorf("expected to match 'myapp', got %v", result)
	}
}

func TestFilter_SubstringMatch(t *testing.T) {
	processes := []Info{
		&process{pid: 100, name: "myapp-worker", rss: 1024},
	}
	result := filter(processes, []string{"app"})
	if len(result) != 1 {
		t.Error("expected substring match")
	}
}

func TestFilter_CaseInsensitive(t *testing.T) {
	processes := []Info{
		&process{pid: 100, name: "MyApp", rss: 1024},
	}
	result := filter(processes, []string{"myapp"})
	if len(result) != 1 {
		t.Error("expected case-insensitive match")
	}
}

func TestFilter_MultipleTerms(t *testing.T) {
	processes := []Info{
		&process{pid: 100, name: "myapp", rss: 1024},
		&process{pid: 200, name: "worker", rss: 2048},
		&process{pid: 300, name: "other", rss: 512},
	}
	result := filter(processes, []string{"myapp", "worker"})
	if len(result) != 2 {
		t.Errorf("expected 2 matches, got %d", len(result))
	}
}

func TestFilter_NoMatch(t *testing.T) {
	processes := []Info{
		&process{pid: 100, name: "myapp", rss: 1024},
	}
	result := filter(processes, []string{"worker"})
	if len(result) != 0 {
		t.Errorf("expected no matches, got %v", result)
	}
}

func TestFilter_ExcludesPID1(t *testing.T) {
	processes := []Info{
		&process{pid: 1, name: "init", rss: 512},
		&process{pid: 100, name: "myapp", rss: 1024},
	}
	result := filter(processes, []string{"init", "myapp"})
	for _, p := range result {
		if p.Pid() == 1 {
			t.Error("should not include PID 1")
		}
	}
}

func TestFilter_ExcludesOwnPID(t *testing.T) {
	ownPID := os.Getpid()
	processes := []Info{
		&process{pid: ownPID, name: "testprocess", rss: 1024},
		&process{pid: 100, name: "testprocess", rss: 2048},
	}
	result := filter(processes, []string{"testprocess"})
	for _, p := range result {
		if p.Pid() == ownPID {
			t.Error("should not include own PID")
		}
	}
	if len(result) != 1 || result[0].Pid() != 100 {
		t.Errorf("expected only PID 100, got %v", result)
	}
}