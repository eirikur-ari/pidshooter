package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestRoster_Spawn_CreatesTargetForEachProcess(t *testing.T) {
	// Given
	processes := []process.Info{
		process.NewInfo(1, "a", 0, 0),
		process.NewInfo(2, "b", 0, 0),
	}
	r := newRoster(processes)
	bounds := movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1})

	// When
	r.spawn(bounds)

	// Then
	assert.Len(t, r.targets, 2)
}

func TestRoster_Move_AdvancesTargets(t *testing.T) {
	// Given
	target := &Target{
		Info:   process.NewInfo(1, "x", 0, 0),
		Motion: movement.Motion{Position: movement.Vector{X: 10, Y: 10}, Velocity: movement.Vector{X: 2.0, Y: 1.0}},
		State:  Alive,
	}
	r := roster{targets: []*Target{target}}
	bounds := movement.NewBounds(movement.WindowSize{Width: 80, Height: 24}, movement.ChromeSize{Top: 1, Bottom: 1})

	// When
	r.move(bounds, 3.0)

	// Then
	assert.Equal(t, 16.0, target.Motion.Position.X)
	assert.Equal(t, 13.0, target.Motion.Position.Y)
}

func TestRoster_AllDead_ReturnsFalseWhenRosterHasNoSpawnedTargets(t *testing.T) {
	// Given
	r := roster{}

	// When
	result := r.allDead()

	// Then
	assert.False(t, result)
}

func TestRoster_AllDead_ReturnsTrueWhenRosterSpawnedTargetsAreAllDead(t *testing.T) {
	// Given
	r := roster{targets: []*Target{
		{Info: process.NewInfo(1, "a", 0, 0), State: Dead},
		{Info: process.NewInfo(2, "b", 0, 0), State: Dead},
	}}

	// When
	result := r.allDead()

	// Then
	assert.True(t, result)
}

func TestRoster_AllDead_ReturnsFalseWhenRosterHasSpawnedTargetThatIsStillAlive(t *testing.T) {
	// Given
	r := roster{targets: []*Target{
		{Info: process.NewInfo(1, "a", 0, 0), State: Dead},
		{Info: process.NewInfo(2, "b", 0, 0), State: Alive},
	}}

	// When
	result := r.allDead()

	// Then
	assert.False(t, result)
}

func TestRoster_HitAt_ReturnsTargetOnHit(t *testing.T) {
	// Given
	target := &Target{
		Info:   process.NewInfo(1, "x", 0, 0),
		Motion: movement.Motion{Position: movement.Vector{X: 10, Y: 5}},
		State:  Alive,
	}
	r := roster{targets: []*Target{target}}

	// When
	result := r.hitAt(10, 5)

	// Then
	assert.Equal(t, target, result)
}

func TestRoster_HitAt_ReturnsNilWhenShotHasAlreadyBeenFired(t *testing.T) {
	// Given
	target := &Target{
		Info:   process.NewInfo(1, "x", 0, 0),
		Motion: movement.Motion{Position: movement.Vector{X: 10, Y: 5}},
		State:  Alive,
	}
	r := roster{targets: []*Target{target}}

	// When
	target.FireShot()
	result := r.hitAt(10, 5)

	// Then
	assert.Nil(t, result, "a repeat hit on a target with a shot already fired at it should be ignored")
}

func TestRoster_HitAt_ReturnsNilOnMiss(t *testing.T) {
	// Given
	target := &Target{
		Info:   process.NewInfo(1, "x", 0, 0),
		Motion: movement.Motion{Position: movement.Vector{X: 10, Y: 5}},
		State:  Alive,
	}
	r := roster{targets: []*Target{target}}

	// When
	result := r.hitAt(0, 0)

	// Then
	assert.Nil(t, result)
}

func TestRoster_Available_ExcludesDeadTargets(t *testing.T) {
	// Given
	r := roster{targets: []*Target{
		{Info: process.NewInfo(1, "a", 0, 0), State: Dead},
	}}

	// When
	targets, alive := r.available()

	// Then
	assert.Empty(t, targets)
	assert.Equal(t, 0, alive)
}

func TestRoster_Available_CountsAlive(t *testing.T) {
	// Given
	r := roster{targets: []*Target{
		{Info: process.NewInfo(1, "a", 0, 0), State: Killing},
		{Info: process.NewInfo(2, "b", 0, 0), State: Alive},
	}}

	// When
	targets, alive := r.available()

	// Then
	require.Len(t, targets, 2) // killing + alive both available
	assert.Equal(t, 1, alive)
}
