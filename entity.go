package main

import (
	"fmt"
	"math/rand"
)

// EntityState represents the current state of a process entity.
type EntityState int

const (
	// StateAlive means the entity is flying around normally.
	StateAlive EntityState = iota
	// StateKilling means the kill animation is playing.
	StateKilling
	// StateDead means the entity should be removed.
	StateDead
)

// killAnimFrames is the number of frames the kill animation lasts.
const killAnimFrames = 12

// Entity represents a process displayed as a flying label in the terminal.
type Entity struct {
	PID           int
	Name          string
	RSS           int64 // Memory in bytes
	X, Y          float64
	VelX, VelY    float64
	State         EntityState
	KillAnimFrame int
}

// NewEntity creates a new entity at a random position with random velocity.
func NewEntity(pid int, name string, rss int64, maxX, maxY int) *Entity {
	label := fmt.Sprintf("[%d %s]", pid, name)
	labelLen := len(label)

	// Ensure entity fits within bounds
	spawnMaxX := maxX - labelLen - 1
	if spawnMaxX < 1 {
		spawnMaxX = 1
	}
	spawnMaxY := maxY - 2 // Leave room for status bar
	if spawnMaxY < 1 {
		spawnMaxY = 1
	}

	// Random position
	x := float64(rand.Intn(spawnMaxX) + 1)
	y := float64(rand.Intn(spawnMaxY) + 1)

	// Random velocity scaled by speed multiplier (default slower)
	velX := (rand.Float64()*0.8 + 0.2)
	if rand.Intn(2) == 0 {
		velX = -velX
	}
	velY := (rand.Float64()*0.4 + 0.1)
	if rand.Intn(2) == 0 {
		velY = -velY
	}

	return &Entity{
		PID:   pid,
		Name:  name,
		RSS:   rss,
		X:     x,
		Y:     y,
		VelX:  velX,
		VelY:  velY,
		State: StateAlive,
	}
}

// Label returns the display string for this entity.
func (e *Entity) Label() string {
	switch e.State {
	case StateKilling:
		// Kill animation frames
		frames := []string{"💥", "✦ KILLED ✦", "· · ·", "  ·  ", "     "}
		idx := e.KillAnimFrame * len(frames) / killAnimFrames
		if idx >= len(frames) {
			idx = len(frames) - 1
		}
		return frames[idx]
	case StateDead:
		return ""
	default:
		return fmt.Sprintf("[%d %s]", e.PID, e.Name)
	}
}

// Update moves the entity and bounces off walls. Speed multiplier scales velocity.
func (e *Entity) Update(maxX, maxY int, speed float64) {
	if e.State == StateKilling {
		e.KillAnimFrame++
		if e.KillAnimFrame >= killAnimFrames {
			e.State = StateDead
		}
		return
	}
	if e.State == StateDead {
		return
	}

	labelLen := float64(len(e.Label()))

	e.X += e.VelX * speed
	e.Y += e.VelY * speed

	// Bounce off horizontal walls
	if e.X < 0 {
		e.X = 0
		e.VelX = -e.VelX
	}
	rightBound := float64(maxX) - labelLen
	if rightBound < 0 {
		rightBound = 0
	}
	if e.X > rightBound {
		e.X = rightBound
		e.VelX = -e.VelX
	}

	// Bounce off vertical walls (leave bottom row for status)
	if e.Y < 0 {
		e.Y = 0
		e.VelY = -e.VelY
	}
	bottomBound := float64(maxY - 2)
	if bottomBound < 0 {
		bottomBound = 0
	}
	if e.Y > bottomBound {
		e.Y = bottomBound
		e.VelY = -e.VelY
	}
}

// Contains checks if the given screen coordinates are within this entity's label.
func (e *Entity) Contains(x, y int) bool {
	if e.State != StateAlive {
		return false
	}

	entityY := int(e.Y)
	entityX := int(e.X)
	labelLen := len(e.Label())

	return y == entityY && x >= entityX && x < entityX+labelLen
}

// StartKillAnim transitions the entity to the killing state.
func (e *Entity) StartKillAnim() {
	e.State = StateKilling
	e.KillAnimFrame = 0
}
