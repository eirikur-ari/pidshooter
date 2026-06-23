package game

import (
	"fmt"
	"os"
	"syscall"

	"github.com/eirikur-ari/pidshooter/internal/process"
)

// EntityState represents the current state of a process entity.
type EntityState int

const (
	// Alive means the entity is flying around normally.
	Alive EntityState = iota
	// Killing means the kill animation is playing.
	Killing
	// Dead means the entity should be removed.
	Dead
)

// KillAnimFrames is the number of frames the kill animation lasts.
const KillAnimFrames = 12

// Entity represents a process displayed as a flying label in the terminal.
type Entity struct {
	process.Info
	Motion
	State         EntityState
	KillAnimFrame int
}

// NewEntity creates a new entity at a random position with random velocity.
func NewEntity(info process.Info, maxX, maxY int) *Entity {
	labelLen := len(fmt.Sprintf("[%d %s]", info.Pid(), info.Name()))
	return &Entity{
		Info:    info,
		Motion: newMotion(maxX, maxY, labelLen),
		State:   Alive,
	}
}

// Label returns the display string for this entity.
func (e *Entity) Label() string {
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
func (e *Entity) Update(maxX, maxY int, speed float64) {
	switch e.State {
	case Killing:
		e.KillAnimFrame++
		if e.KillAnimFrame >= KillAnimFrames {
			e.State = Dead
		}
	case Alive:
		e.Motion.Update(maxX, maxY, float64(len(e.Label())), speed)
	}
}

// Contains checks if the given screen coordinates are within this entity's label.
func (e *Entity) Contains(x, y int) bool {
	if e.State != Alive {
		return false
	}

	entityY := int(e.PosY)
	entityX := int(e.PosX)
	labelLen := len(e.Label())

	return y == entityY && x >= entityX && x < entityX+labelLen
}

// StartKillAnim transitions the entity to the killing state.
func (e *Entity) StartKillAnim() {
	e.State = Killing
	e.KillAnimFrame = 0
}

// Kill sends SIGKILL to the process this entity represents.
func (e *Entity) Kill() error {
	proc, err := os.FindProcess(e.Pid())
	if err != nil {
		return err
	}
	return proc.Signal(syscall.SIGKILL)
}
