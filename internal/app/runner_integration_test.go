//go:build integration

package app

import (
	"errors"
	"strings"
	"testing"

	gamedriven "github.com/eirikur-ari/pidshooter/internal/domain/game/ports/driven"
	"github.com/eirikur-ari/pidshooter/internal/domain/game/ports/driving"
	procdriven "github.com/eirikur-ari/pidshooter/internal/domain/process/ports/driven"
	"github.com/eirikur-ari/pidshooter/internal/testutil/capture"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

func TestGameService_HappyPath(t *testing.T) {
	events := newStubEventSource()
	events.ch <- gamedriven.KeyEvent{Ch: 'q'}

	store := &fake.Store{}
	svc := NewGameService(
		&fake.Finder{Processes: []procdriven.Info{fake.NewProcess(1, "target", 1024)}},
		&fake.Killer{},
		store,
		&stubRenderer{},
		events,
	)

	if err := svc.Play(driving.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if store.Saved == nil {
		t.Fatal("expected Save to be called, got nil")
	}
	if len(store.Saved.Scores) != 1 {
		t.Fatalf("expected 1 score entry, got %d", len(store.Saved.Scores))
	}
	entry := store.Saved.Scores[0]
	if entry.Kills != 0 {
		t.Errorf("expected 0 kills after immediate quit, got %d", entry.Kills)
	}
	if entry.Duration <= 0 {
		t.Errorf("expected positive duration (StartTime was recorded), got %f", entry.Duration)
	}
}

func TestGameService_SaveError_PrintsWarning(t *testing.T) {
	events := newStubEventSource()
	events.ch <- gamedriven.KeyEvent{Ch: 'q'}

	svc := NewGameService(
		&fake.Finder{Processes: []procdriven.Info{fake.NewProcess(1, "target", 1024)}},
		&fake.Killer{},
		&fake.Store{SaveErr: errors.New("disk full")},
		&stubRenderer{},
		events,
	)

	var err error
	stderr := capture.Stderr(func() {
		err = svc.Play(driving.Config{Patterns: []string{"target"}, Speed: 2.0, TimeLimit: 0})
	})

	if err != nil {
		t.Fatalf("expected Play to return nil, got %v", err)
	}
	if !strings.Contains(stderr, "score not saved") {
		t.Errorf("expected stderr warning about save failure, got: %q", stderr)
	}
	if !strings.Contains(stderr, "disk full") {
		t.Errorf("expected stderr to include underlying error, got: %q", stderr)
	}
}
