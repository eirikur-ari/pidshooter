package runner

import (
	"errors"
	"testing"

	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func TestStart_FinderError(t *testing.T) {
	cfg := Config{Patterns: []string{"foo"}, Speed: 2.0, TimeLimit: 30}
	err := start(cfg, &testutil.FakeFinder{Err: errors.New("ps failed")})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestStart_NoProcesses(t *testing.T) {
	cfg := Config{Patterns: []string{"nonexistent"}, Speed: 2.0, TimeLimit: 30}
	err := start(cfg, &testutil.FakeFinder{})
	if err != nil {
		t.Errorf("expected nil error for empty results, got %v", err)
	}
}
