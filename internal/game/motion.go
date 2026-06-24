package game

import "math/rand"

// Motion holds the position and velocity of a moving entity.
type Motion struct {
	PosX, PosY float64
	VelX, VelY float64
}

// newMotion returns a Motion value with a random spawn position and velocity
// that fits within the terminal bounds for an entity of the given label length.
func newMotion(maxX, maxY, labelLen int) Motion {
	spawnMaxX := maxX - labelLen - 1
	if spawnMaxX < 1 {
		spawnMaxX = 1
	}
	spawnMaxY := maxY - 2
	if spawnMaxY < 1 {
		spawnMaxY = 1
	}

	x := float64(rand.Intn(spawnMaxX) + 1)
	y := float64(rand.Intn(spawnMaxY) + 1)

	velX := rand.Float64()*0.8 + 0.2
	if rand.Intn(2) == 0 {
		velX = -velX
	}
	velY := rand.Float64()*0.4 + 0.1
	if rand.Intn(2) == 0 {
		velY = -velY
	}

	return Motion{PosX: x, PosY: y, VelX: velX, VelY: velY}
}

// Update moves the entity and bounces off walls. labelLen is the display
// width of the entity, needed to compute the right-wall boundary.
func (p *Motion) Update(maxX, maxY int, labelLen float64, speed float64) {
	p.PosX += p.VelX * speed
	p.PosY += p.VelY * speed

	if p.PosX < 0 {
		p.PosX = 0
		p.VelX = -p.VelX
	}
	rightBound := float64(maxX) - labelLen
	if rightBound < 0 {
		rightBound = 0
	}
	if p.PosX > rightBound {
		p.PosX = rightBound
		p.VelX = -p.VelX
	}

	if p.PosY < 0 {
		p.PosY = 0
		p.VelY = -p.VelY
	}
	bottomBound := float64(maxY - 2)
	if bottomBound < 0 {
		bottomBound = 0
	}
	if p.PosY > bottomBound {
		p.PosY = bottomBound
		p.VelY = -p.VelY
	}
}
