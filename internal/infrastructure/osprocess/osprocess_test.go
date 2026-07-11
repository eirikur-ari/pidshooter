package osprocess

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func newTestProcess(t *testing.T) *Process {
	t.Helper()
	path, err := exec.LookPath("ps")
	if err != nil {
		t.Fatalf("ps not found: %v", err)
	}
	return &Process{psPath: path}
}

func TestValidate_EmptySlice(t *testing.T) {
	err := validate([]string{})
	if err == nil {
		t.Fatal("expected error for empty slice")
	}
	if !errors.Is(err, process.ErrNoPatterns) {
		t.Errorf("expected process.ErrNoPatterns, got %v", err)
	}
}

func TestValidate_EmptyTerm(t *testing.T) {
	if err := validate([]string{""}); err == nil {
		t.Error("expected error for empty term")
	}
}

func TestValidate_TooShort(t *testing.T) {
	for _, p := range []string{"a", "ab"} {
		if err := validate([]string{p}); err == nil {
			t.Errorf("expected error for pattern %q shorter than MinPatternLength", p)
		}
	}
}

func TestValidate_ExactMinLength(t *testing.T) {
	min := strings.Repeat("a", process.MinPatternLength)
	if err := validate([]string{min}); err != nil {
		t.Errorf("expected no error for exactly min-length pattern, got %v", err)
	}
}

func TestValidate_TooLong(t *testing.T) {
	long := strings.Repeat("a", process.MaxPatternLength+1)
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
	_, err := newTestProcess(t).Find([]string{})
	if err == nil {
		t.Error("expected error for empty patterns")
	}
}

func TestFind_EmptyTerm(t *testing.T) {
	_, err := newTestProcess(t).Find([]string{""})
	if err == nil {
		t.Error("expected error for empty term")
	}
}

func TestFilter_MatchesByName(t *testing.T) {
	processes := []process.Info{
		{Pid: 100, Name: "myapp", Rss: 1024},
		{Pid: 200, Name: "worker", Rss: 2048},
	}
	result := filter(processes, []string{"myapp"})
	if len(result) != 1 || result[0].Pid != 100 {
		t.Errorf("expected to match 'myapp', got %v", result)
	}
}

func TestFilter_SubstringMatch(t *testing.T) {
	processes := []process.Info{
		{Pid: 100, Name: "myapp-worker", Rss: 1024},
	}
	result := filter(processes, []string{"app"})
	if len(result) != 1 {
		t.Error("expected substring match")
	}
}

func TestFilter_CaseInsensitive(t *testing.T) {
	processes := []process.Info{
		{Pid: 100, Name: "MyApp", Rss: 1024},
	}
	result := filter(processes, []string{"myapp"})
	if len(result) != 1 {
		t.Error("expected case-insensitive match")
	}
}

func TestFilter_MultipleTerms(t *testing.T) {
	processes := []process.Info{
		{Pid: 100, Name: "myapp", Rss: 1024},
		{Pid: 200, Name: "worker", Rss: 2048},
		{Pid: 300, Name: "other", Rss: 512},
	}
	result := filter(processes, []string{"myapp", "worker"})
	if len(result) != 2 {
		t.Errorf("expected 2 matches, got %d", len(result))
	}
}

func TestFilter_NoMatch(t *testing.T) {
	processes := []process.Info{
		{Pid: 100, Name: "myapp", Rss: 1024},
	}
	result := filter(processes, []string{"worker"})
	if len(result) != 0 {
		t.Errorf("expected no matches, got %v", result)
	}
}

func TestFilter_ExcludesPID1(t *testing.T) {
	processes := []process.Info{
		{Pid: 1, Name: "init", Rss: 512},
		{Pid: 100, Name: "myapp", Rss: 1024},
	}
	result := filter(processes, []string{"init", "myapp"})
	for _, p := range result {
		if p.Pid == 1 {
			t.Error("should not include PID 1")
		}
	}
}

func TestFilter_ExcludesOwnPID(t *testing.T) {
	ownPID := os.Getpid()
	processes := []process.Info{
		{Pid: ownPID, Name: "testprocess", Rss: 1024},
		{Pid: 100, Name: "testprocess", Rss: 2048},
	}
	result := filter(processes, []string{"testprocess"})
	for _, p := range result {
		if p.Pid == ownPID {
			t.Error("should not include own PID")
		}
	}
	if len(result) != 1 || result[0].Pid != 100 {
		t.Errorf("expected only PID 100, got %v", result)
	}
}

func TestFilter_ExcludesPID0(t *testing.T) {
	processes := []process.Info{
		{Pid: 0, Name: "swapper", Rss: 0},
		{Pid: 100, Name: "myapp", Rss: 1024},
	}
	result := filter(processes, []string{"swapper", "myapp"})
	for _, p := range result {
		if p.Pid <= 1 {
			t.Errorf("filter should exclude PID %d", p.Pid)
		}
	}
	if len(result) != 1 || result[0].Pid != 100 {
		t.Errorf("expected only PID 100, got %v", result)
	}
}

func TestKiller_InvalidPID(t *testing.T) {
	_, err := newTestProcess(t).Kill(-1, "")
	if err == nil {
		t.Error("expected error killing invalid PID -1, got nil")
	}
}

func TestKiller_RefusesPID0(t *testing.T) {
	if _, err := newTestProcess(t).Kill(0, ""); err == nil {
		t.Error("expected error killing PID 0, got nil")
	}
}

func TestKiller_RefusesPID1(t *testing.T) {
	if _, err := newTestProcess(t).Kill(1, ""); err == nil {
		t.Error("expected error killing PID 1, got nil")
	}
}

func TestKiller_RefusesNegativePID(t *testing.T) {
	p := newTestProcess(t)
	for _, pid := range []int{-1, -100, -99999} {
		if _, err := p.Kill(pid, ""); err == nil {
			t.Errorf("expected error killing PID %d, got nil", pid)
		}
	}
}

func TestValidateProcessName_Match(t *testing.T) {
	if err := validateProcessName("myapp", "myapp"); err != nil {
		t.Errorf("expected no error for matching names, got %v", err)
	}
}

func TestValidateProcessName_Mismatch(t *testing.T) {
	if err := validateProcessName("myapp", "otherd"); err == nil {
		t.Error("expected error for mismatched names, got nil")
	}
}

func TestValidateProcessName_EmptyActual(t *testing.T) {
	if err := validateProcessName("myapp", ""); err == nil {
		t.Error("expected error when actual name is empty, got nil")
	}
}

func TestKiller_RefusesNameMismatch(t *testing.T) {
	ownPID := os.Getpid()
	killed, err := newTestProcess(t).Kill(ownPID, "definitely-not-this-process")
	if err != nil {
		t.Errorf("unexpected error on name mismatch: %v", err)
	}
	if killed {
		t.Error("expected not killed when name does not match current process name")
	}
}

func TestValidate_ExactMaxLength(t *testing.T) {
	exact := strings.Repeat("a", process.MaxPatternLength)
	if err := validate([]string{exact}); err != nil {
		t.Errorf("expected no error for exactly max-length pattern, got %v", err)
	}
}
