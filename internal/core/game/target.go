package game

import (
	"fmt"
	"math/rand"
	"unicode/utf8"

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

// Target represents a process displayed as a flying label in the terminal.
type Target struct {
	process.Info
	Position          FrameVector
	Velocity          FrameVector
	State             TargetState
	KillAnimationTick int
}

// NewTarget creates a new entity at a random position with random velocity.
func NewTarget(info process.Info, bounds FrameBounds) *Target {
	return &Target{
		Info:     info,
		Position: newRandomPosition(info, bounds.W, bounds.H),
		Velocity: newRandomVelocity(),
		State:    Alive,
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
		return fmt.Sprintf("[%d %s]", t.Pid, t.Name)
	}
}

// Update advances the kill animation or moves the entity and bounces off walls.
func (t *Target) Update(bounds FrameBounds, speed float64) {
	switch t.State {
	case Killing:
		t.doomsdayTick()
	case Alive:
		t.move(bounds, speed)
	case Dead:
		// nothing to do
	}
}

// IsHitAt reports whether the given game-space coordinates (x=column, y=row)
// fall within this target's tag.
func (t *Target) IsHitAt(x, y int) bool {
	if t.State != Alive {
		return false
	}
	width := utf8.RuneCountInString(t.Tag())
	return y == int(t.Position.Y) && x >= int(t.Position.X) && x < int(t.Position.X)+width
}

// StartKillAnimation transitions the entity to the killing state.
func (t *Target) StartKillAnimation() {
	t.State = Killing
	t.KillAnimationTick = 0
}

// newRandomPosition returns a random spawn position that keeps the target within bounds.
func newRandomPosition(info process.Info, maxX, maxY int) FrameVector {
	labelLen := utf8.RuneCountInString(fmt.Sprintf("[%d %s]", info.Pid, info.Name))
	x := maxX - labelLen - 1
	if x < 1 {
		x = 1
	}
	y := maxY - 2 // Leave room for status bar
	if y < 1 {
		y = 1
	}
	return FrameVector{X: float64(rand.Intn(x) + 1), Y: float64(rand.Intn(y) + 1)}
}

// newRandomVelocity returns a FrameVector with randomized direction and magnitude
// suitable for initial target velocity.
func newRandomVelocity() FrameVector {
	x := rand.Float64()*0.5 + 0.5
	if rand.Intn(2) == 0 {
		x = -x
	}
	y := rand.Float64()*0.25 + 0.25
	if rand.Intn(2) == 0 {
		y = -y
	}
	return FrameVector{X: x, Y: y}
}

func (t *Target) doomsdayTick() {
	t.KillAnimationTick++
	if t.KillAnimationTick >= KillAnimationDuration {
		t.State = Dead
	}
}

func killAnimationTagFor(tick int) string {
	tags := []string{"💥", "✦ KILLED ✦", "· · ·", "  ·  ", "     "}
	idx := tick * len(tags) / KillAnimationDuration
	if idx >= len(tags) {
		idx = len(tags) - 1
	}
	return tags[idx]
}

// move advances the target's position and bounces it off the frame walls.
func (t *Target) move(bounds FrameBounds, speed float64) {
	t.Velocity.Apply(&t.Position, speed)
	width := float64(utf8.RuneCountInString(t.Tag()))
	bounds.Bounce(&t.Position, &t.Velocity, width)
}
