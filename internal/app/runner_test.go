package app

import (
	"errors"
	"testing"

	gamedriven "github.com/eirikur-ari/pidshooter/internal/domain/game/ports/driven"
	"github.com/eirikur-ari/pidshooter/internal/domain/game/ports/driving"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

type stubRenderer struct{}

func (r *stubRenderer) Init() error                  { return nil }
func (r *stubRenderer) Cleanup()                     {}
func (r *stubRenderer) Size() (int, int)             { return 80, 24 }
func (r *stubRenderer) Render(_ gamedriven.Frame)    {}

type stubEventSource struct{ ch chan gamedriven.InputEvent }

func newStubEventSource() *stubEventSource {
	return &stubEventSource{ch: make(chan gamedriven.InputEvent, 100)}
}
func (e *stubEventSource) Events() <-chan gamedriven.InputEvent { return e.ch }

func TestGameService_FinderError(t *testing.T) {
	svc := NewGameService(
		&testutil.FakeFinder{Err: errors.New("ps failed")},
		&testutil.FakeKiller{},
		&testutil.FakeStore{},
		&stubRenderer{},
		newStubEventSource(),
	)
	err := svc.Play(driving.Config{Patterns: []string{"foo"}, Speed: 2.0, TimeLimit: 30})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGameService_NoProcesses(t *testing.T) {
	svc := NewGameService(
		&testutil.FakeFinder{},
		&testutil.FakeKiller{},
		&testutil.FakeStore{},
		&stubRenderer{},
		newStubEventSource(),
	)
	err := svc.Play(driving.Config{Patterns: []string{"nonexistent"}, Speed: 2.0, TimeLimit: 30})
	if err != nil {
		t.Errorf("expected nil error for empty results, got %v", err)
	}
}
