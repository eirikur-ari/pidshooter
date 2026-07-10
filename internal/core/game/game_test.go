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

	g := New(processes, true, 3.5, 60)

	if !g.confirmMode {
		t.Error("expected confirmMode=true")
	}
	if g.velocity.Speed() != 3.5 {
		t.Errorf("expected speed=3.5, got %f", g.velocity.Speed())
	}
	if g.timeLimit != 60 {
		t.Errorf("expected timeLimit=60, got %d", g.timeLimit)
	}
	if !g.running.Load() {
		t.Error("expected running=true")
	}
	if g.kills != 0 {
		t.Errorf("expected kills=0, got %d", g.kills)
	}
	if g.freedMem != 0 {
		t.Errorf("expected freedMem=0, got %d", g.freedMem)
	}
}
