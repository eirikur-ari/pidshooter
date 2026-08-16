package game

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

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
