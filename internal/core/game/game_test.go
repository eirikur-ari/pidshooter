package game

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestNew(t *testing.T) {
	processes := []process.Info{
		{Pid: 1, Name: "a", Rss: 100},
		{Pid: 2, Name: "b", Rss: 200},
	}

	g := New(processes, Config{Confirm: true, Speed: 3.5, TimeLimit: 60})

	assert.True(t, g.cfg.Confirm)
	assert.Equal(t, 3.5, g.velocity.Speed())
	assert.Equal(t, 60, g.cfg.TimeLimit)
	assert.Equal(t, Pending, g.State())
	assert.Equal(t, 0, g.kills)
	assert.Equal(t, int64(0), g.freedMem)
}

func TestStart_TransitionsToRunning(t *testing.T) {
	processes := []process.Info{{Pid: 1, Name: "a", Rss: 100}}
	g := New(processes, Config{})

	g.Start(80, 24)

	assert.True(t, g.IsRunning())
	assert.Len(t, g.targets, 1)
}

func TestStop_TransitionsToStopped(t *testing.T) {
	g := New(nil, Config{})
	g.Start(0, 0)

	g.Stop()

	assert.Equal(t, Stopped, g.State())
}

func TestStart_PanicsWhenRunning(t *testing.T) {
	g := New(nil, Config{})
	g.Start(0, 0)

	assert.Panics(t, func() { g.Start(80, 24) })
}

func TestStart_PanicsWhenStopped(t *testing.T) {
	g := New(nil, Config{})
	g.Start(0, 0)
	g.Stop()

	assert.Panics(t, func() { g.Start(80, 24) })
}
