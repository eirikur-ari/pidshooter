package game

import (
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestNewTargetWithinBounds(t *testing.T) {
	maxX, maxY := 80, 24
	e := NewTarget(process.NewInfo(1234, "test", 1024), movement.NewBounds(maxX, maxY))

	assert.Equal(t, 1234, e.Pid)
	assert.Equal(t, "test", e.Name)
	assert.Equal(t, int64(1024), e.Rss)
	assert.Equal(t, Alive, e.State)

	tag := fmt.Sprintf("[%d %s]", e.Pid, e.Name)
	width := len(tag)
	spawnMaxX := maxX - width - 1
	spawnMaxY := maxY - 2
	assert.GreaterOrEqual(t, e.Position.X, 1.0)
	assert.LessOrEqual(t, int(e.Position.X), spawnMaxX)
	assert.GreaterOrEqual(t, e.Position.Y, 1.0)
	assert.LessOrEqual(t, int(e.Position.Y), spawnMaxY)
}

func TestNewTargetSmallTerminal(t *testing.T) {
	e := NewTarget(process.NewInfo(1, "xxx", 0), movement.NewBounds(5, 5))
	require.NotNil(t, e)
}

func TestTargetTagAlive(t *testing.T) {
	e := &Target{Info: process.NewInfo(42, "bash", 0), State: Alive}
	assert.Equal(t, "[42 bash]", e.Tag())
}

func TestTargetTagDead(t *testing.T) {
	e := &Target{Info: process.NewInfo(42, "bash", 0), State: Dead}
	assert.Equal(t, "", e.Tag())
}

func TestTargetTagKilling(t *testing.T) {
	e := &Target{Info: process.NewInfo(42, "bash", 0), State: Killing, KillAnimationTick: 0}
	assert.NotEmpty(t, e.Tag())
}

func TestTargetUpdateKillingState(t *testing.T) {
	e := &Target{
		Info:              process.NewInfo(1, "xxx", 0),
		State:             Killing,
		KillAnimationTick: KillAnimationDuration - 1,
	}
	e.Update(movement.NewBounds(80, 24), 1.0)
	assert.Equal(t, Dead, e.State)
}

func TestTargetUpdateDeadNoOp(t *testing.T) {
	e := &Target{
		Info:   process.NewInfo(1, "xxx", 0),
		Motion: movement.Motion{Position: movement.Vector{X: 10, Y: 10}, Velocity: movement.Vector{X: 1.0, Y: 1.0}},
		State:  Dead,
	}
	e.Update(movement.NewBounds(80, 24), 1.0)
	assert.Equal(t, 10.0, e.Position.X, "dead entity should not move")
	assert.Equal(t, 10.0, e.Position.Y, "dead entity should not move")
}

func TestTargetUpdateBounceLeft(t *testing.T) {
	e := &Target{
		Info:   process.NewInfo(1, "x", 0),
		Motion: movement.Motion{Position: movement.Vector{X: 0, Y: 5}, Velocity: movement.Vector{X: -1.0, Y: 0}},
		State:  Alive,
	}
	e.Update(movement.NewBounds(80, 24), 1.0)
	assert.GreaterOrEqual(t, e.Position.X, 0.0, "Position.X should not be negative after left bounce")
	assert.Greater(t, e.Velocity.X, 0.0, "Velocity.X should be positive after left bounce")
}

func TestTargetUpdateBounceRight(t *testing.T) {
	// tag "[1 x]" = 5 chars → rightBound = 80-5 = 75
	e := &Target{
		Info:   process.NewInfo(1, "x", 0),
		Motion: movement.Motion{Position: movement.Vector{X: 75, Y: 5}, Velocity: movement.Vector{X: 2.0, Y: 0}},
		State:  Alive,
	}
	e.Update(movement.NewBounds(80, 24), 1.0)
	rightBound := 75.0
	assert.LessOrEqual(t, e.Position.X, rightBound, "Position.X should not exceed right bound after right bounce")
	assert.Less(t, e.Velocity.X, 0.0, "Velocity.X should be negative after right bounce")
}

func TestTargetUpdateBounceTop(t *testing.T) {
	e := &Target{
		Info:   process.NewInfo(1, "x", 0),
		Motion: movement.Motion{Position: movement.Vector{X: 5, Y: 0}, Velocity: movement.Vector{X: 0, Y: -1.0}},
		State:  Alive,
	}
	e.Update(movement.NewBounds(80, 24), 1.0)
	assert.GreaterOrEqual(t, e.Position.Y, 0.0, "Position.Y should not be negative after top bounce")
	assert.Greater(t, e.Velocity.Y, 0.0, "Velocity.Y should be positive after top bounce")
}

func TestTargetUpdateBounceBottom(t *testing.T) {
	e := &Target{
		Info:   process.NewInfo(1, "x", 0),
		Motion: movement.Motion{Position: movement.Vector{X: 5, Y: 23}, Velocity: movement.Vector{X: 0, Y: 2.0}},
		State:  Alive,
	}
	e.Update(movement.NewBounds(80, 24), 1.0)
	bottomBound := float64(24 - 2)
	assert.LessOrEqual(t, e.Position.Y, bottomBound, "Position.Y should not exceed bottom bound after bottom bounce")
	assert.Less(t, e.Velocity.Y, 0.0, "Velocity.Y should be negative after bottom bounce")
}

func TestTargetUpdateSpeedMultiplier(t *testing.T) {
	// tag "[1 x]" = 5 chars; at (40,10) with speed=3 there is no wall bounce.
	e := &Target{
		Info:   process.NewInfo(1, "x", 0),
		Motion: movement.Motion{Position: movement.Vector{X: 40, Y: 10}, Velocity: movement.Vector{X: 1.0, Y: 0.5}},
		State:  Alive,
	}
	e.Update(movement.NewBounds(80, 24), 3.0)
	assert.Equal(t, 43.0, e.Position.X)
	assert.Equal(t, 11.5, e.Position.Y)
}

func TestTargetUpdateMultiByteRightWall(t *testing.T) {
	// "[42 café]" is 9 runes but 10 UTF-8 bytes.
	// With the byte-count bug, rightBound = maxX - 10 = 70.
	// With the fix, rightBound = maxX - 9 = 71.
	// Place the entity at Position.X=70.5 moving right at speed=1. After one update:
	//   fix:  new Position.X = 71.0 — at the correct boundary, no bounce yet.
	//   bug:  new Position.X > 70 → bounce, Velocity.X flips negative.
	e := &Target{
		Info:   process.NewInfo(42, "café", 0),
		Motion: movement.Motion{Position: movement.Vector{X: 70.5, Y: 5}, Velocity: movement.Vector{X: 0.5, Y: 0}},
		State:  Alive,
	}
	e.Update(movement.NewBounds(80, 24), 1.0)
	assert.Greater(t, e.Velocity.X, 0.0,
		"entity bounced prematurely at right wall — byte-count bug in Update? Position.X=%.1f Velocity.X=%.1f",
		e.Position.X, e.Velocity.X)
}

func TestTargetContains(t *testing.T) {
	e := &Target{
		Info:   process.NewInfo(42, "bash", 0),
		Motion: movement.Motion{Position: movement.Vector{X: 10, Y: 5}, Velocity: movement.Vector{X: 0, Y: 0}},
		State:  Alive,
	}
	tag := e.Tag()
	width := len(tag)

	assert.True(t, e.IsHitAt(10, 5))
	assert.True(t, e.IsHitAt(10+width-1, 5))
	assert.False(t, e.IsHitAt(9, 5))
	assert.False(t, e.IsHitAt(10+width, 5))
	assert.False(t, e.IsHitAt(10, 4))
}

func TestTargetContainsMultiByteProcessName(t *testing.T) {
	// "café" is 5 UTF-8 bytes but 4 runes → tag "[42 café]" is 10 bytes, 9 runes.
	// With the byte-count bug, IsHitAt over-counts by 1 and accepts column 19 as a hit.
	e := &Target{
		Info:   process.NewInfo(42, "café", 0),
		Motion: movement.Motion{Position: movement.Vector{X: 10, Y: 5}, Velocity: movement.Vector{X: 0, Y: 0}},
		State:  Alive,
	}
	tag := e.Tag()
	runeCount := utf8.RuneCountInString(tag)
	pastEnd := 10 + runeCount
	assert.False(t, e.IsHitAt(pastEnd, 5),
		"IsHitAt(%d, 5) should be false for tag %q (rune count %d) — byte-count bug?",
		pastEnd, tag, runeCount)
}

func TestTargetContainsNotAlive(t *testing.T) {
	e := &Target{
		Info:   process.NewInfo(42, "bash", 0),
		Motion: movement.Motion{Position: movement.Vector{X: 10, Y: 5}, Velocity: movement.Vector{X: 0, Y: 0}},
		State:  Killing,
	}
	assert.False(t, e.IsHitAt(10, 5), "non-alive entity should not be hit")
}

func TestTargetIsAlive(t *testing.T) {
	assert.True(t, (&Target{State: Alive}).IsAlive())
	assert.False(t, (&Target{State: Killing}).IsAlive())
	assert.False(t, (&Target{State: Dead}).IsAlive())
}

func TestTargetIsKilling(t *testing.T) {
	assert.True(t, (&Target{State: Killing}).IsKilling())
	assert.False(t, (&Target{State: Alive}).IsKilling())
	assert.False(t, (&Target{State: Dead}).IsKilling())
}

func TestTargetIsDead(t *testing.T) {
	assert.True(t, (&Target{State: Dead}).IsDead())
	assert.False(t, (&Target{State: Alive}).IsDead())
	assert.False(t, (&Target{State: Killing}).IsDead())
}

func TestTargetKill(t *testing.T) {
	e := &Target{Info: process.NewInfo(1, "xxx", 0), State: Alive, KillAnimationTick: 5}
	assert.True(t, e.Kill())
	assert.Equal(t, Killing, e.State)
	assert.Equal(t, 0, e.KillAnimationTick)
}

func TestTargetKillNoOpWhenNotAlive(t *testing.T) {
	e := &Target{Info: process.NewInfo(1, "xxx", 0), State: Dead}
	assert.False(t, e.Kill())
	assert.Equal(t, Dead, e.State)
}

func TestTargetSnapshotAliveTarget(t *testing.T) {
	e := &Target{
		Info:   process.NewInfo(1, "x", 0),
		Motion: movement.Motion{Position: movement.Vector{X: 10.6, Y: 5.4}, Velocity: movement.Vector{X: 0, Y: 0}},
		State:  Alive,
	}

	s := e.Snapshot()

	assert.Equal(t, 11, s.X)
	assert.Equal(t, 5, s.Y)
	assert.Equal(t, "[1 x]", s.Tag)
	assert.False(t, s.Killing)
}

func TestTargetSnapshotKillingTarget(t *testing.T) {
	e := &Target{
		Info:              process.NewInfo(1, "x", 0),
		Motion:            movement.Motion{Position: movement.Vector{X: 3.0, Y: 7.0}, Velocity: movement.Vector{X: 0, Y: 0}},
		State:             Killing,
		KillAnimationTick: 0,
	}

	s := e.Snapshot()

	assert.True(t, s.Killing)
}
