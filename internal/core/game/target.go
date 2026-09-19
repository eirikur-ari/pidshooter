package game

import (
	"fmt"
	"unicode/utf8"

	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// State represents the current state of a process entity.
type State int

const (
	// Alive means the entity is flying around normally.
	Alive State = iota
	// Killing means the kill animation is playing.
	Killing
	// Fleeing means the flee animation is playing, for a target whose
	// backing process was found already gone before a kill could land.
	Fleeing
	// Dead means the entity has been removed.
	Dead
)

// AnimationDuration is the number of game ticks the kill or flee animation
// plays before the target disappears.
const AnimationDuration = 36

type Target struct {
	process.Info
	movement.Motion
	State         State
	AnimationTick int
	shotFired     bool
}

// NewTarget creates a new entity at a random position with random velocity.
func NewTarget(info process.Info, bounds movement.Bounds) *Target {
	return &Target{
		Info:   info,
		Motion: movement.NewMotion(bounds, utf8.RuneCountInString(tagFor(info))),
		State:  Alive,
	}
}

// Tag returns the display label for this target while it is alive, and
// empty otherwise.
func (t *Target) Tag() string {
	if t.State != Alive {
		return ""
	}
	return tagFor(t.Info)
}

// AnimationProgress reports how far through its kill or flee animation this
// target is, as a fraction from 0 to 1. Zero when neither Killing nor
// Fleeing is true.
func (t *Target) AnimationProgress() float64 {
	if t.State != Killing && t.State != Fleeing {
		return 0
	}
	return float64(t.AnimationTick) / float64(AnimationDuration)
}

// Update advances the kill or flee animation, or moves the entity and bounces off walls.
func (t *Target) Update(bounds movement.Bounds, speed float64) {
	switch t.State {
	case Killing, Fleeing:
		t.doomsdayTick()
	case Alive:
		t.move(bounds, speed)
	case Dead:
		// nothing to do
	}
}

// Kill transitions the target to the dying state and returns true.
// Returns false without changing state if the target is not alive.
func (t *Target) Kill() bool {
	if !t.isAlive() {
		return false
	}
	t.State = Killing
	t.AnimationTick = 0
	return true
}

// Reap transitions the target to the fleeing state, skipping the kill
// animation, for a target whose backing process already exited outside the
// game. The target plays its flee animation before disappearing, the same
// as a confirmed kill does. Returns false without changing state if the
// target is not alive.
func (t *Target) Reap() bool {
	if !t.isAlive() {
		return false
	}
	t.State = Fleeing
	t.AnimationTick = 0
	return true
}

// FireShot marks that a kill attempt for this target is in flight, so it is
// not returned as a hit again until fire has ceased.
func (t *Target) FireShot() { t.shotFired = true }

// CeaseFire clears the marker set by FireShot, regardless of how the shot resolved.
func (t *Target) CeaseFire() { t.shotFired = false }

// isAlive reports whether the target is flying around and can be shot.
func (t *Target) isAlive() bool { return t.State == Alive }

// isDead reports whether the target has finished its kill animation and been removed.
func (t *Target) isDead() bool { return t.State == Dead }

// isHitAt reports whether the given game-space coordinates (x=column, y=row)
// fall within this target's tag. A target with a shot already fired at it
// is not hit again until fire has ceased.
func (t *Target) isHitAt(x, y int) bool {
	if t.State != Alive || t.shotFired {
		return false
	}
	width := utf8.RuneCountInString(t.Tag())
	return y == int(t.Position.Y) && x >= int(t.Position.X) && x < int(t.Position.X)+width
}

func (t *Target) doomsdayTick() {
	t.AnimationTick++
	if t.AnimationTick >= AnimationDuration {
		t.State = Dead
	}
}

func (t *Target) move(bounds movement.Bounds, speed float64) {
	t.Motion.Move(bounds, speed, float64(utf8.RuneCountInString(tagFor(t.Info))))
}

func tagFor(info process.Info) string {
	return fmt.Sprintf("[%d %s]", info.PID, info.Name)
}
