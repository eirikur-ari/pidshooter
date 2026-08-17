package game

import (
	"fmt"
	"math"
	"unicode/utf8"

	"github.com/eirikur-ari/pidshooter/internal/core/movement"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// TargetState represents the current state of a process entity.
type TargetState int

const (
	// Alive means the entity is flying around normally.
	Alive TargetState = iota
	// Killing means the kill animation is playing.
	Killing
	// Dead means the entity has been removed.
	Dead
)

// KillAnimationDuration is the number of game ticks the kill animation plays before the target disappears.
const KillAnimationDuration = 12

// Target represents a process displayed as a flying tag in the terminal.
type Target struct {
	process.Info
	movement.Motion
	State             TargetState
	KillAnimationTick int
}

// TargetSnapshot is a point-in-time read model of a visible target.
type TargetSnapshot struct {
	X, Y    int
	Tag     string
	Killing bool
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

// IsDead reports whether the target has finished its kill animation and been removed.
func (t *Target) IsDead() bool { return t.State == Dead }

// IsAlive reports whether the target is flying around and can be shot.
func (t *Target) IsAlive() bool { return t.State == Alive }

// IsKilling reports whether the kill animation is playing.
func (t *Target) IsKilling() bool { return t.State == Killing }

// IsHitAt reports whether the given game-space coordinates (x=column, y=row)
// fall within this target's tag.
func (t *Target) IsHitAt(x, y int) bool {
	if t.State != Alive {
		return false
	}
	width := utf8.RuneCountInString(t.Tag())
	return y == int(t.Position.Y) && x >= int(t.Position.X) && x < int(t.Position.X)+width
}

// Kill transitions the target to the dying state and returns true.
// Returns false without changing state if the target is not alive.
func (t *Target) Kill() bool {
	if !t.IsAlive() {
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
	if !t.IsAlive() {
		return false
	}
	t.State = Dead
	return true
}

// Snapshot returns the current renderable state of this target.
func (t *Target) Snapshot() TargetSnapshot {
	return TargetSnapshot{
		X:       int(math.Round(t.Position.X)),
		Y:       int(math.Round(t.Position.Y)),
		Tag:     t.Tag(),
		Killing: t.IsKilling(),
	}
}

func (t *Target) doomsdayTick() {
	t.KillAnimationTick++
	if t.KillAnimationTick >= KillAnimationDuration {
		t.State = Dead
	}
}

func tagFor(info process.Info) string {
	return fmt.Sprintf("[%d %s]", info.Pid, info.Name)
}

func killAnimationTagFor(tick int) string {
	tags := []string{"💥", "✦ KILLED ✦", "· · ·", "  ·  ", "     "}
	idx := tick * len(tags) / KillAnimationDuration
	if idx >= len(tags) {
		idx = len(tags) - 1
	}
	return tags[idx]
}

func (t *Target) move(bounds movement.Bounds, speed float64) {
	t.Motion.Move(bounds, speed, float64(utf8.RuneCountInString(tagFor(t.Info))))
}
