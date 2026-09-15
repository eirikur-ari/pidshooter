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
	// Dead means the entity has been removed.
	Dead
)

// KillAnimationDuration is the number of game ticks the kill animation plays before the target disappears.
const KillAnimationDuration = 12

// TODO: perhaps move target to process package, and have it implement a Target interface in the game package, so that the game package doesn't need to know about process.Info
// Target represents a process displayed as a flying tag in the terminal.
type Target struct {
	process.Info
	movement.Motion
	State             State
	KillAnimationTick int
	shotFired         bool
}

// NewTarget creates a new entity at a random position with random velocity.
func NewTarget(info process.Info, bounds movement.Bounds) *Target {
	return &Target{
		Info:   info,
		Motion: movement.NewMotion(bounds, utf8.RuneCountInString(tagFor(info))),
		State:  Alive,
	}
}

// Tag returns the display string for this target.
func (t *Target) Tag() string {
	switch t.State {
	case Killing:
		return killAnimationTagFor(t.KillAnimationTick)
	case Dead:
		return ""
	default:
		return tagFor(t.Info)
	}
}

// Update advances the kill animation or moves the entity and bounces off walls.
func (t *Target) Update(bounds movement.Bounds, speed float64) {
	switch t.State {
	case Killing:
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
	t.KillAnimationTick = 0
	return true
}

// Reap transitions the target straight to Dead, skipping the kill animation,
// for a target whose backing process already exited outside the game.
// Returns false without changing state if the target is not alive.
func (t *Target) Reap() bool {
	if !t.isAlive() {
		return false
	}
	t.State = Dead
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
	t.KillAnimationTick++
	if t.KillAnimationTick >= KillAnimationDuration {
		t.State = Dead
	}
}

func (t *Target) move(bounds movement.Bounds, speed float64) {
	t.Motion.Move(bounds, speed, float64(utf8.RuneCountInString(tagFor(t.Info))))
}

func tagFor(info process.Info) string {
	return fmt.Sprintf("[%d %s]", info.PID, info.Name)
}

// TODO: perhaps move tag string etc to infrastructure / outbound adapter
func killAnimationTagFor(tick int) string {
	tags := []string{"💥", "✦ KILLED ✦", "· · ·", "  ·  ", "     "}
	idx := tick * len(tags) / KillAnimationDuration
	if idx >= len(tags) {
		idx = len(tags) - 1
	}
	return tags[idx]
}
