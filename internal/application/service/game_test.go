package service

import (
	"errors"
	"testing"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/core/event"
	"github.com/eirikur-ari/pidshooter/internal/core/game"
	"github.com/eirikur-ari/pidshooter/internal/testutil/fake"
)

type stubRenderer struct{}

func (r *stubRenderer) Init() error         { return nil }
func (r *stubRenderer) Cleanup()            {}
func (r *stubRenderer) Size() (int, int)    { return 80, 24 }
func (r *stubRenderer) Render(_ game.Frame) {}

type stubEventSource struct{ ch chan event.InputEvent }

func newStubEventSource() *stubEventSource {
	return &stubEventSource{ch: make(chan event.InputEvent, 100)}
}
func (e *stubEventSource) Events() <-chan event.InputEvent { return e.ch }

func TestGameService_FinderError(t *testing.T) {
	svc := NewGameService(
		&fake.Process{FindErr: errors.New("ps failed")},
		&fake.Store{},
		&stubRenderer{},
		newStubEventSource(),
	)
	err := svc.Play(inbound.GamePlayConfig{Patterns: []string{"foo"}, Speed: 2.0, TimeLimit: 30})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGameService_NoProcesses(t *testing.T) {
	svc := NewGameService(
		&fake.Process{},
		&fake.Store{},
		&stubRenderer{},
		newStubEventSource(),
	)
	err := svc.Play(inbound.GamePlayConfig{Patterns: []string{"nonexistent"}, Speed: 2.0, TimeLimit: 30})
	if err != nil {
		t.Errorf("expected nil error for empty results, got %v", err)
	}
}
