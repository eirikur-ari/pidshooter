package game

import (
	"testing"

	gamedriven "github.com/eirikur-ari/pidshooter/internal/domain/game/ports/driven"
	procdriven "github.com/eirikur-ari/pidshooter/internal/domain/process/ports/driven"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

// stubRenderer and stubEventSource are package-internal test doubles.
// They can't live in testutil because testutil would need to import domain/game/ports,
// and domain/game/ports is a sub-package of domain/game — importing it from testutil
// and then importing testutil from domain/game's internal tests would form a cycle.
type stubRenderer struct {
	w, h   int
	frames []gamedriven.Frame
}

func (r *stubRenderer) Init() error        { return nil }
func (r *stubRenderer) Cleanup()           {}
func (r *stubRenderer) Size() (int, int)   { return r.w, r.h }
func (r *stubRenderer) Render(f gamedriven.Frame) { r.frames = append(r.frames, f) }

type stubEventSource struct{ ch chan gamedriven.InputEvent }

func newStubEventSource() *stubEventSource {
	return &stubEventSource{ch: make(chan gamedriven.InputEvent, 100)}
}
func (e *stubEventSource) Events() <-chan gamedriven.InputEvent { return e.ch }

func TestNew(t *testing.T) {
	processes := []procdriven.Info{
		testutil.NewFakeProcess(1, "a", 100),
		testutil.NewFakeProcess(2, "b", 200),
	}

	killer := &testutil.FakeKiller{}
	renderer := &stubRenderer{w: 80, h: 24}
	events := newStubEventSource()

	g := New(processes, true, 3.5, 60, killer, renderer, events)

	if !g.confirmMode {
		t.Error("expected confirmMode=true")
	}
	if g.speed != 3.5 {
		t.Errorf("expected speed=3.5, got %f", g.speed)
	}
	if g.timeLimit != 60 {
		t.Errorf("expected timeLimit=60, got %d", g.timeLimit)
	}
	if !g.running {
		t.Error("expected running=true")
	}
	if g.kills != 0 {
		t.Errorf("expected kills=0, got %d", g.kills)
	}
	if g.freedMem != 0 {
		t.Errorf("expected freedMem=0, got %d", g.freedMem)
	}
}
