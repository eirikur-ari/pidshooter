package application

import (
	"errors"
	"testing"

	"github.com/eirikur-ari/pidshooter/internal/application/contract"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

type stubRenderer struct{}

func (r *stubRenderer) Init() error         { return nil }
func (r *stubRenderer) Cleanup()            {}
func (r *stubRenderer) Size() (int, int)    { return 80, 24 }
func (r *stubRenderer) Render(_ game.Frame) {}

type stubEventSource struct{ ch chan contract.InputEvent }

func newStubEventSource() *stubEventSource {
	return &stubEventSource{ch: make(chan contract.InputEvent, 100)}
}
func (e *stubEventSource) Events() <-chan contract.InputEvent { return e.ch }

func TestGameRunner_FinderError(t *testing.T) {
	svc := NewGameRunner(
		&fake.Finder{Err: errors.New("ps failed")},
		&fake.Killer{},
		&fake.Store{},
		&stubRenderer{},
		newStubEventSource(),
	)
	err := svc.Run(contract.Config{Patterns: []string{"foo"}, Speed: 2.0, TimeLimit: 30})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGameRunner_NoProcesses(t *testing.T) {
	svc := NewGameRunner(
		&fake.Finder{},
		&fake.Killer{},
		&fake.Store{},
		&stubRenderer{},
		newStubEventSource(),
	)
	err := svc.Run(contract.Config{Patterns: []string{"nonexistent"}, Speed: 2.0, TimeLimit: 30})
	if err != nil {
		t.Errorf("expected nil error for empty results, got %v", err)
	}
}
