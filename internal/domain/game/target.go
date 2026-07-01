package game

import (
	"fmt"
	"unicode/utf8"

	procdriven "github.com/eirikur-ari/pidshooter/internal/domain/process/ports/driven"
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

// KillAnimFrames is the number of frames the kill animation lasts.
const KillAnimFrames = 12

// Target represents a process displayed as a flying label in the terminal.
type Target struct {
	procdriven.Info
	Motion
	State         TargetState
	KillAnimFrame int
}

// NewTarget creates a new entity at a random position with random velocity.
func NewTarget(info procdriven.Info, maxX, maxY int) *Target {
	labelLen := utf8.RuneCountInString(fmt.Sprintf("[%d %s]", info.Pid(), info.Name()))
	return &Target{
		Info:   info,
		Motion: newMotion(maxX, maxY, labelLen),
		State:  Alive,
	}
}

// Label returns the display string for this entity.
func (e *Target) Label() string {
	switch e.State {
	case Killing:
		frames := []string{"💥", "✦ KILLED ✦", "· · ·", "  ·  ", "     "}
		idx := e.KillAnimFrame * len(frames) / KillAnimFrames
		if idx >= len(frames) {
			idx = len(frames) - 1
		}
		return frames[idx]
	case Dead:
		return ""
	default:
		return fmt.Sprintf("[%d %s]", e.Pid(), e.Name())
	}
}

// Update advances the kill animation or delegates movement to Motion.
func (e *Target) Update(maxX, maxY int, speed float64) {
	switch e.State {
	case Killing:
		e.KillAnimFrame++
		if e.KillAnimFrame >= KillAnimFrames {
			e.State = Dead
		}
	case Alive:
		e.Motion.Update(maxX, maxY, float64(utf8.RuneCountInString(e.Label())), speed)
	}
}

// Contains checks if the given screen coordinates are within this entity's label.
func (e *Target) Contains(x, y int) bool {
	if e.State != Alive {
		return false
	}
	entityY := int(e.PosY)
	entityX := int(e.PosX)
	labelLen := utf8.RuneCountInString(e.Label())
	return y == entityY && x >= entityX && x < entityX+labelLen
}

// StartKillAnim transitions the entity to the killing state.
func (e *Target) StartKillAnim() {
	e.State = Killing
	e.KillAnimFrame = 0
}
