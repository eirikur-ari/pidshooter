package game

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestKillTransitionsToKilling(t *testing.T) {
	tgt := &Target{Info: process.NewInfo(1, "target", 2048), State: Alive}
	g := &Game{}

	g.Kill(tgt)

	assert.Equal(t, Killing, tgt.State)
	assert.Equal(t, 1, g.kills)
	assert.Equal(t, int64(2048), g.freedMem)
}

func TestKillNoOpWhenNotAlive(t *testing.T) {
	tgt := &Target{Info: process.NewInfo(1, "target", 2048), State: Dead}
	g := &Game{}

	g.Kill(tgt)

	assert.Equal(t, 0, g.kills)
}

func TestUpdateStopsWhenTimeLimitExpired(t *testing.T) {
	g := New(nil, Config{TimeLimit: 1})
	g.Start(0, 0)
	g.timer.start = time.Now().Add(-2 * time.Second)

	g.Update(80, 24)

	assert.False(t, g.IsRunning())
}

func TestUpdateStopsWhenAllTargetsDead(t *testing.T) {
	tgt := &Target{Info: process.NewInfo(1, "target", 0), State: Dead}
	g := &Game{targets: []*Target{tgt}, throttle: movement.NewThrottle(movement.MinSpeed)}
	g.Start(0, 0)

	g.Update(80, 24)

	assert.False(t, g.IsRunning())
}
