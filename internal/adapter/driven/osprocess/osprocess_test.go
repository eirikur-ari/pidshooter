package osprocess

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/eirikur-ari/pidshooter/internal/domain/process/ports/driven"
)

func newTestFinder(t *testing.T) *Finder {
	t.Helper()
	path, err := exec.LookPath("ps")
	if err != nil {
		t.Fatalf("ps not found: %v", err)
	}
	return &Finder{psPath: path}
}

func TestValidate_EmptySlice(t *testing.T) {
	if err := validate([]string{}); err == nil {
		t.Error("expected error for empty slice")
	}
}

func TestValidate_EmptyTerm(t *testing.T) {
	if err := validate([]string{""}); err == nil {
		t.Error("expected error for empty term")
	}
}

func TestValidate_TooLong(t *testing.T) {
	long := strings.Repeat("a", driven.MaxPatternLength+1)
	if err := validate([]string{long}); err == nil {
		t.Error("expected error for term exceeding max length")
	}
}

func TestValidate_Valid(t *testing.T) {
	if err := validate([]string{"myapp", "worker"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFind_EmptyPatterns(t *testing.T) {
	_, err := newTestFinder(t).Find([]string{})
	if err == nil {
		t.Error("expected error for empty patterns")
	}
}

func TestFind_EmptyTerm(t *testing.T) {
	_, err := newTestFinder(t).Find([]string{""})
	if err == nil {
		t.Error("expected error for empty term")
	}
}

func TestFilter_MatchesByName(t *testing.T) {
	processes := []driven.Info{
		&proc{pid: 100, name: "myapp", rss: 1024},
		&proc{pid: 200, name: "worker", rss: 2048},
	}
	result := filter(processes, []string{"myapp"})
	if len(result) != 1 || result[0].Pid() != 100 {
		t.Errorf("expected to match 'myapp', got %v", result)
	}
}

func TestFilter_SubstringMatch(t *testing.T) {
	processes := []driven.Info{
		&proc{pid: 100, name: "myapp-worker", rss: 1024},
	}
	result := filter(processes, []string{"app"})
	if len(result) != 1 {
		t.Error("expected substring match")
	}
}

func TestFilter_CaseInsensitive(t *testing.T) {
	processes := []driven.Info{
		&proc{pid: 100, name: "MyApp", rss: 1024},
	}
	result := filter(processes, []string{"myapp"})
	if len(result) != 1 {
		t.Error("expected case-insensitive match")
	}
}

func TestFilter_MultipleTerms(t *testing.T) {
	processes := []driven.Info{
		&proc{pid: 100, name: "myapp", rss: 1024},
		&proc{pid: 200, name: "worker", rss: 2048},
		&proc{pid: 300, name: "other", rss: 512},
	}
	result := filter(processes, []string{"myapp", "worker"})
	if len(result) != 2 {
		t.Errorf("expected 2 matches, got %d", len(result))
	}
}

func TestFilter_NoMatch(t *testing.T) {
	processes := []driven.Info{
		&proc{pid: 100, name: "myapp", rss: 1024},
	}
	result := filter(processes, []string{"worker"})
	if len(result) != 0 {
		t.Errorf("expected no matches, got %v", result)
	}
}

func TestFilter_ExcludesPID1(t *testing.T) {
	processes := []driven.Info{
		&proc{pid: 1, name: "init", rss: 512},
		&proc{pid: 100, name: "myapp", rss: 1024},
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
	processes := []driven.Info{
		&proc{pid: ownPID, name: "testprocess", rss: 1024},
		&proc{pid: 100, name: "testprocess", rss: 2048},
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

func TestProc_Fields(t *testing.T) {
	p := &proc{pid: 42, name: "myapp", rss: 8192}
	if p.Pid() != 42 {
		t.Errorf("expected Pid=42, got %d", p.Pid())
	}
	if p.Name() != "myapp" {
		t.Errorf("expected Name=myapp, got %q", p.Name())
	}
	if p.Rss() != 8192 {
		t.Errorf("expected Rss=8192, got %d", p.Rss())
	}
}

func TestKiller_InvalidPID(t *testing.T) {
	k := NewKiller()
	// PID -1 is invalid on all platforms and should return an error.
	err := k.Kill(-1)
	if err == nil {
		t.Error("expected error killing invalid PID -1, got nil")
	}
}

func TestKiller_RefusesPID0(t *testing.T) {
	k := NewKiller()
	err := k.Kill(0)
	if err == nil {
		t.Error("expected error killing PID 0, got nil")
	}
}

func TestKiller_RefusesPID1(t *testing.T) {
	k := NewKiller()
	err := k.Kill(1)
	if err == nil {
		t.Error("expected error killing PID 1, got nil")
	}
}

func TestKiller_RefusesNegativePID(t *testing.T) {
	k := NewKiller()
	for _, pid := range []int{-1, -100, -99999} {
		if err := k.Kill(pid); err == nil {
			t.Errorf("expected error killing PID %d, got nil", pid)
		}
	}
}

func TestFilter_ExcludesPID0(t *testing.T) {
	processes := []driven.Info{
		&proc{pid: 0, name: "swapper", rss: 0},
		&proc{pid: 100, name: "myapp", rss: 1024},
	}
	result := filter(processes, []string{"swapper", "myapp"})
	for _, p := range result {
		if p.Pid() <= 1 {
			t.Errorf("filter should exclude PID %d", p.Pid())
		}
	}
	if len(result) != 1 || result[0].Pid() != 100 {
		t.Errorf("expected only PID 100, got %v", result)
	}
}

func TestValidate_ExactMaxLength(t *testing.T) {
	exact := strings.Repeat("a", driven.MaxPatternLength)
	if err := validate([]string{exact}); err != nil {
		t.Errorf("expected no error for exactly max-length pattern, got %v", err)
	}
}
