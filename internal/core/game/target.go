package game

import (
	"fmt"
	"math/rand"
	"unicode/utf8"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

// Vector is a 2D vector used for both position and velocity.
type Vector struct {
	X, Y float64
}

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

// KillAnimFrames is the number of frames the kill animation lasts.
const KillAnimFrames = 12

// Target represents a process displayed as a flying label in the terminal.
type Target struct {
	process.Info
	Position      Vector
	Velocity      Vector
	State         TargetState
	KillAnimFrame int
}

// NewTarget creates a new entity at a random position with random velocity.
func NewTarget(info process.Info, maxX, maxY int) *Target {
	labelLen := utf8.RuneCountInString(fmt.Sprintf("[%d %s]", info.Pid, info.Name))

	// Ensure entity fits within bounds
	spawnMaxX := maxX - labelLen - 1
	if spawnMaxX < 1 {
		spawnMaxX = 1
	}
	spawnMaxY := maxY - 2 // Leave room for status bar
	if spawnMaxY < 1 {
		spawnMaxY = 1
	}

	// Random velocity scaled by speed multiplier (default slower)
	velX := rand.Float64()*0.8 + 0.2
	if rand.Intn(2) == 0 {
		velX = -velX
	}
	velY := rand.Float64()*0.4 + 0.1
	if rand.Intn(2) == 0 {
		velY = -velY
	}

	// Random position
	return &Target{
		Info:     info,
		Position: Vector{X: float64(rand.Intn(spawnMaxX) + 1), Y: float64(rand.Intn(spawnMaxY) + 1)},
		Velocity: Vector{X: velX, Y: velY},
		State:    Alive,
	}
}

// Label returns the display string for this entity.
func (e *Target) Label() string {
	switch e.State {
	case Killing:
		// Kill animation frames
		frames := []string{"💥", "✦ KILLED ✦", "· · ·", "  ·  ", "     "}
		idx := e.KillAnimFrame * len(frames) / KillAnimFrames
		if idx >= len(frames) {
			idx = len(frames) - 1
		}
		return frames[idx]
	case Dead:
		return ""
	default:
		return fmt.Sprintf("[%d %s]", e.Pid, e.Name)
	}
}

// Update advances the kill animation or moves the entity and bounces off walls.
func (e *Target) Update(maxX, maxY int, speed float64) {
	switch e.State {
	case Killing:
		e.KillAnimFrame++
		if e.KillAnimFrame >= KillAnimFrames {
			e.State = Dead
		}
	case Alive:
		labelLen := float64(utf8.RuneCountInString(e.Label()))
		e.Position.X += e.Velocity.X * speed
		e.Position.Y += e.Velocity.Y * speed

		// Bounce off horizontal walls
		if e.Position.X < 0 {
			e.Position.X = 0
			e.Velocity.X = -e.Velocity.X
		}
		rightBound := float64(maxX) - labelLen
		if rightBound < 0 {
			rightBound = 0
		}
		if e.Position.X > rightBound {
			e.Position.X = rightBound
			e.Velocity.X = -e.Velocity.X
		}

		// Bounce off vertical walls (leave bottom row for status)
		if e.Position.Y < 0 {
			e.Position.Y = 0
			e.Velocity.Y = -e.Velocity.Y
		}
		bottomBound := float64(maxY - 2)
		if bottomBound < 0 {
			bottomBound = 0
		}
		if e.Position.Y > bottomBound {
			e.Position.Y = bottomBound
			e.Velocity.Y = -e.Velocity.Y
		}
	}
}

// Contains checks if the given screen coordinates are within this entity's label.
func (e *Target) Contains(x, y int) bool {
	if e.State != Alive {
		return false
	}
	labelLen := utf8.RuneCountInString(e.Label())
	return y == int(e.Position.Y) && x >= int(e.Position.X) && x < int(e.Position.X)+labelLen
}

// StartKillAnim transitions the entity to the killing state.
func (e *Target) StartKillAnim() {
	e.State = Killing
	e.KillAnimFrame = 0
}
