package game

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestTarget_NewTarget_StartsAliveWithGivenInfo(t *testing.T) {
	// Given
	info := newInfoFixture()
	target := NewTarget(info, newBoundsFixture())

	// Then
	assert.Equal(t, info, target.Info)
	assert.Equal(t, Alive, target.State)
}

func TestTarget_Tag_ReturnsLabelWhenAlive(t *testing.T) {
	// Given
	info := newInfoFixture()
	info.PID = 42
	info.Name = "deep thought"
	target := &Target{Info: info, State: Alive}

	// When
	tag := target.Tag()

	// Then
	assert.Equal(t, "[42 deep thought]", tag)
}

func TestTarget_Tag_IsEmptyWhenNotAlive(t *testing.T) {
	for _, state := range []State{Killing, Fleeing, Dead} {
		// Given
		target := &Target{Info: newInfoFixture(), State: state}

		// When
		tag := target.Tag()

		// Then
		assert.Empty(t, tag, "state %d should not render a label; the renderer draws its own animation frame", state)
	}
}

func TestTarget_AnimationProgress_IsZeroWhenInAliveOrDeadState(t *testing.T) {
	for _, state := range []State{Alive, Dead} {
		// Given
		target := &Target{Info: newInfoFixture(), State: state, AnimationTick: 5}

		// When
		progress := target.AnimationProgress()

		// Then
		assert.Equal(t, 0.0, progress, "state %d has no animation to report progress on", state)
	}
}

func TestTarget_AnimationProgress_ReturnsElapsedFractionWhenInKillingOrFleeingState(t *testing.T) {
	for _, state := range []State{Killing, Fleeing} {
		// Given
		target := &Target{Info: newInfoFixture(), State: state, AnimationTick: AnimationDuration / 2}

		// When
		progress := target.AnimationProgress()

		// Then
		assert.Equal(t, 0.5, progress, "state %d should report its tick as a fraction of the animation duration", state)
	}
}

func TestTarget_Update_AdvancesAnimationTickWhenInKillingOrFleeingState(t *testing.T) {
	for _, state := range []State{Killing, Fleeing} {
		// Given
		target := &Target{Info: newInfoFixture(), State: state, AnimationTick: 3}

		// When
		target.Update(newBoundsFixture(), 1.0)

		// Then
		assert.Equal(t, state, target.State, "state %d should keep animating before the final tick", state)
		assert.Equal(t, 4, target.AnimationTick, "state %d should advance one tick per update", state)
	}
}

func TestTarget_Update_BecomesDeadOnFinalAnimationTickWhenInKillingOrFleeingState(t *testing.T) {
	for _, state := range []State{Killing, Fleeing} {
		// Given
		target := &Target{Info: newInfoFixture(), State: state, AnimationTick: AnimationDuration - 1}

		// When
		target.Update(newBoundsFixture(), 1.0)

		// Then
		assert.Equal(t, Dead, target.State, "state %d should become Dead once its animation completes", state)
	}
}

func TestTarget_Update_DoesNotMoveWhenInDeadState(t *testing.T) {
	// Given
	target := &Target{
		Info:   newInfoFixture(),
		Motion: movement.Motion{Position: movement.Vector{X: 10, Y: 10}, Velocity: movement.Vector{X: 1.0, Y: 1.0}},
		State:  Dead,
	}

	// When
	target.Update(newBoundsFixture(), 1.0)

	// Then
	assert.Equal(t, 10.0, target.Motion.Position.X, "dead target should not move")
	assert.Equal(t, 10.0, target.Motion.Position.Y, "dead target should not move")
}

func TestTarget_Update_MovesAliveTargetByVelocityScaledBySpeed(t *testing.T) {
	// Given
	info := newInfoFixture()
	info.PID = 1
	info.Name = "x"

	// tag "[1 x]" = 5 chars; at (40,10) with speed=3 there is no wall bounce.
	target := &Target{
		Info:   info,
		Motion: movement.Motion{Position: movement.Vector{X: 40, Y: 10}, Velocity: movement.Vector{X: 1.0, Y: 0.5}},
		State:  Alive,
	}

	// When
	target.Update(newBoundsFixture(), 3.0)

	// Then
	assert.Equal(t, 43.0, target.Motion.Position.X) // X: 40 + (1.0 * 3) = 43
	assert.Equal(t, 11.5, target.Motion.Position.Y) // Y: 10 + (0.5 * 3) = 11.5
}

func TestTarget_Update_DoesNotBounceAtRightWallWhenTagFitsByRuneWidth(t *testing.T) {
	// Given
	info := newInfoFixture()
	info.PID = 42
	info.Name = "café"
	// "[42 café]" is 9 runes but 10 UTF-8 bytes, so the right wall is at 90-9 = 81.
	// Moving from 70.5 by 0.5 lands exactly on the wall, which must not bounce.
	target := &Target{
		Info:   info,
		Motion: movement.Motion{Position: movement.Vector{X: 70.5, Y: 5}, Velocity: movement.Vector{X: 0.5, Y: 0}},
		State:  Alive,
	}
	windowSize := movement.WindowSize{Width: 90, Height: 24}
	chromeSize := chromeSizeFixture()
	bounds := movement.NewBounds(windowSize, chromeSize) // right wall at 90-9=81

	// When
	target.Update(bounds, 1.0)

	// Then
	assert.Equal(t, 71.0, target.Motion.Position.X)
	assert.Greater(t, target.Motion.Velocity.X, 0.0, "a tag landing exactly on the wall by rune width should not bounce")
}

func TestTarget_Kill_StartsKillAnimation(t *testing.T) {
	// Given
	target := &Target{Info: newInfoFixture(), State: Alive, AnimationTick: 5}

	// When
	killed := target.Kill()

	// Then
	assert.True(t, killed)
	assert.Equal(t, Killing, target.State)
	assert.Equal(t, 0, target.AnimationTick)
}

func TestTarget_Kill_IsRejectedWhenNotAlive(t *testing.T) {
	for _, state := range []State{Killing, Fleeing, Dead} {
		// Given
		target := &Target{Info: newInfoFixture(), State: state}

		// When
		killed := target.Kill()

		// Then
		assert.False(t, killed, "state %d cannot be killed", state)
		assert.Equal(t, state, target.State)
	}
}

func TestTarget_Reap_StartsFleeingAnimation(t *testing.T) {
	// Given
	target := &Target{Info: newInfoFixture(), State: Alive, AnimationTick: 5}

	// When
	reaped := target.Reap()

	// Then
	assert.True(t, reaped)
	assert.Equal(t, Fleeing, target.State)
	assert.Equal(t, 0, target.AnimationTick)
}

func TestTarget_Reap_IsRejectedWhenNotAlive(t *testing.T) {
	for _, state := range []State{Killing, Fleeing, Dead} {
		// Given
		target := &Target{Info: newInfoFixture(), State: state}

		// When
		reaped := target.Reap()

		// Then
		assert.False(t, reaped, "state %d cannot be reaped", state)
		assert.Equal(t, state, target.State)
	}
}

func TestTarget_isHitAt_IsTrueWithinTagBounds(t *testing.T) {
	// Given
	target := aliveTargetAt(42, "bash", 10, 5)
	width := utf8.RuneCountInString(target.Tag())

	// Then
	assert.True(t, target.isHitAt(10, 5))
	assert.True(t, target.isHitAt(10+width-1, 5))
}

func TestTarget_isHitAt_IsFalseOutsideTagBounds(t *testing.T) {
	// Given
	target := aliveTargetAt(42, "bash", 10, 5)
	width := utf8.RuneCountInString(target.Tag())

	// Then
	assert.False(t, target.isHitAt(9, 5), "one column before the tag's start should miss")
	assert.False(t, target.isHitAt(10+width, 5), "one column past the tag's end should miss")
	assert.False(t, target.isHitAt(10, 4), "the row above the tag should miss")
}

func TestTarget_isHitAt_MatchesRoundedRenderPositionNotTruncated(t *testing.T) {
	// Given
	target := aliveTargetAt(42, "bash", 10.6, 5.6)
	width := utf8.RuneCountInString(target.Tag())

	// Then
	assert.True(t, target.isHitAt(11, 6), "clicking at the rounded position, where the tag is actually rendered, should hit")
	assert.True(t, target.isHitAt(11+width-1, 6))
	assert.False(t, target.isHitAt(10, 6), "clicking at the truncated column, one left of where the tag is rendered, should miss")
	assert.False(t, target.isHitAt(11, 5), "clicking at the truncated row, one above where the tag is rendered, should miss")
}

func TestTarget_isHitAt_CountsTagInRunesNotBytes(t *testing.T) {
	// Given
	// "[42 café]" is 9 runes but 10 UTF-8 bytes.
	target := aliveTargetAt(42, "café", 10, 5)
	runeCount := utf8.RuneCountInString(target.Tag())

	// Then
	assert.True(t, target.isHitAt(10+runeCount-1, 5), "the tag's last rune should hit")
	assert.False(t, target.isHitAt(10+runeCount, 5), "one column past the tag's last rune should miss")
}

func TestTarget_isHitAt_IsFalseWhenNotAlive(t *testing.T) {
	for _, state := range []State{Killing, Fleeing, Dead} {
		// Given
		target := aliveTargetAt(42, "bash", 10, 5)

		// When
		target.State = state

		// Then
		assert.False(t, target.isHitAt(10, 5), "a target in state %d should not be hit", state)
	}
}

func TestTarget_isHitAt_IsFalseWhileShotFired(t *testing.T) {
	// Given
	target := aliveTargetAt(42, "bash", 10, 5)

	// When
	target.FireShot()
	hit := target.isHitAt(10, 5)

	// Then
	assert.False(t, hit, "a target with a shot already fired at it should not be hit again")
}

func TestTarget_isHitAt_IsTrueAfterCeaseFire(t *testing.T) {
	// Given
	target := aliveTargetAt(42, "bash", 10, 5)

	// When
	target.FireShot()
	target.CeaseFire()

	// Then
	assert.True(t, target.isHitAt(10, 5), "a target should be hittable again once fire has ceased")
}

func aliveTargetAt(pid int, name string, x, y float64) *Target {
	return &Target{
		Info:   process.NewInfo(pid, name, 0, 0),
		Motion: movement.Motion{Position: movement.Vector{X: x, Y: y}},
		State:  Alive,
	}
}

func TestTarget_move_IgnoresTargetWhenNotAlive(t *testing.T) {
	for _, state := range []State{Killing, Fleeing, Dead} {
		// Given
		target := &Target{
			Info:   newInfoFixture(),
			Motion: movement.Motion{Position: movement.Vector{X: 10, Y: 10}, Velocity: movement.Vector{X: 1.0, Y: 1.0}},
			State:  state,
		}

		// When
		target.move(newBoundsFixture(), 1.0)

		// Then
		assert.Equal(t, 10.0, target.Motion.Position.X, "move must not move a target in state %d", state)
		assert.Equal(t, 10.0, target.Motion.Position.Y, "move must not move a target in state %d", state)
	}
}
