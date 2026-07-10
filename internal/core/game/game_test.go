package game

import (
	"testing"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestNew(t *testing.T) {
	processes := []process.Info{
		{Pid: 1, Name: "a", Rss: 100},
		{Pid: 2, Name: "b", Rss: 200},
	}

	g := New(processes, Config{ConfirmMode: true, Speed: 3.5, TimeLimit: 60})

	if !g.cfg.ConfirmMode {
		t.Error("expected confirmMode=true")
	}
	if g.velocity.Speed() != 3.5 {
		t.Errorf("expected speed=3.5, got %f", g.velocity.Speed())
	}
	if g.cfg.TimeLimit != 60 {
		t.Errorf("expected timeLimit=60, got %d", g.cfg.TimeLimit)
	}
	if g.State() != Pending {
		t.Error("expected game in Pending state")
	}
	if g.kills != 0 {
		t.Errorf("expected kills=0, got %d", g.kills)
	}
	if g.freedMem != 0 {
		t.Errorf("expected freedMem=0, got %d", g.freedMem)
	}
}

func TestStart_TransitionsToRunning(t *testing.T) {
	processes := []process.Info{{Pid: 1, Name: "a", Rss: 100}}
	g := New(processes, Config{})

	g.Start(80, 24)

	if !g.IsRunning() {
		t.Error("expected game in Running state")
	}
	if len(g.targets) != 1 {
		t.Errorf("expected 1 target, got %d", len(g.targets))
	}
}

func TestStop_TransitionsToStopped(t *testing.T) {
	g := New(nil, Config{})
	g.Start(0, 0)

	g.Stop()

	if g.State() != Stopped {
		t.Errorf("expected Stopped after Stop, got %d", g.State())
	}
}

func TestStart_PanicsWhenRunning(t *testing.T) {
	g := New(nil, Config{})
	g.Start(0, 0)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when starting a running game")
		}
	}()
	g.Start(80, 24)
}

func TestStart_PanicsWhenStopped(t *testing.T) {
	g := New(nil, Config{})
	g.Start(0, 0)
	g.Stop()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when starting a stopped game")
		}
	}()
	g.Start(80, 24)
}
