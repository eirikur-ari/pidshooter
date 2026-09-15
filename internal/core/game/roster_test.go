package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestNewRosterEmptyBeforeSpawn(t *testing.T) {
	r := newRoster([]process.Info{process.NewInfo(1, "a", 0, 0)})
	assert.Empty(t, r.targets)
}

func TestRosterSpawnCreatesTargetForEachProcess(t *testing.T) {
	processes := []process.Info{
		process.NewInfo(1, "a", 0, 0),
		process.NewInfo(2, "b", 0, 0),
	}
	r := newRoster(processes)

	r.spawn(movement.NewBounds(80, 24))

	assert.Len(t, r.targets, 2)
}

func TestRosterMoveAdvancesTargets(t *testing.T) {
	tgt := &Target{
		Info:   process.NewInfo(1, "x", 0, 0),
		Motion: movement.Motion{Position: movement.Vector{X: 10, Y: 10}, Velocity: movement.Vector{X: 1.0, Y: 0}},
		State:  Alive,
	}
	r := roster{targets: []*Target{tgt}}

	r.move(movement.NewBounds(80, 24), 1.0)

	assert.NotEqual(t, 10.0, tgt.Position.X)
}

func TestRosterAllDeadFalseWhenEmpty(t *testing.T) {
	r := roster{}
	assert.False(t, r.allDead())
}

func TestRosterAllDeadTrueWhenAllDead(t *testing.T) {
	r := roster{targets: []*Target{
		{Info: process.NewInfo(1, "a", 0, 0), State: Dead},
		{Info: process.NewInfo(2, "b", 0, 0), State: Dead},
	}}
	assert.True(t, r.allDead())
}

func TestRosterAllDeadFalseWhenSomeAlive(t *testing.T) {
	r := roster{targets: []*Target{
		{Info: process.NewInfo(1, "a", 0, 0), State: Dead},
		{Info: process.NewInfo(2, "b", 0, 0), State: Alive},
	}}
	assert.False(t, r.allDead())
}

func TestRosterHitAtReturnsTargetOnHit(t *testing.T) {
	tgt := &Target{
		Info:   process.NewInfo(1, "x", 0, 0),
		Motion: movement.Motion{Position: movement.Vector{X: 10, Y: 5}},
		State:  Alive,
	}
	r := roster{targets: []*Target{tgt}}

	assert.Equal(t, tgt, r.hitAt(10, 5))
}

func TestRosterHitAtReturnsNilWhenShotAlreadyFired(t *testing.T) {
	tgt := &Target{
		Info:   process.NewInfo(1, "x", 0, 0),
		Motion: movement.Motion{Position: movement.Vector{X: 10, Y: 5}},
		State:  Alive,
	}
	tgt.FireShot()
	r := roster{targets: []*Target{tgt}}

	assert.Nil(t, r.hitAt(10, 5), "a repeat hit on a target with a shot already fired at it should be ignored")
}

func TestRosterHitAtReturnsNilOnMiss(t *testing.T) {
	tgt := &Target{
		Info:   process.NewInfo(1, "x", 0, 0),
		Motion: movement.Motion{Position: movement.Vector{X: 10, Y: 5}},
		State:  Alive,
	}
	r := roster{targets: []*Target{tgt}}

	assert.Nil(t, r.hitAt(0, 0))
}

func TestRosterAvailableExcludesDeadTargets(t *testing.T) {
	r := roster{targets: []*Target{
		{Info: process.NewInfo(1, "a", 0, 0), State: Dead},
	}}

	targets, alive := r.available()

	assert.Empty(t, targets)
	assert.Equal(t, 0, alive)
}

func TestRosterAvailableCountsAlive(t *testing.T) {
	r := roster{targets: []*Target{
		{Info: process.NewInfo(1, "a", 0, 0), State: Killing},
		{Info: process.NewInfo(2, "b", 0, 0), State: Alive},
	}}

	targets, alive := r.available()

	require.Len(t, targets, 2) // killing + alive both available
	assert.Equal(t, 1, alive)
}
