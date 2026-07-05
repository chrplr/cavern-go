package main

import "strconv"

const boltSpeed = 7

// Bolt is a laser fired by a robot. It travels horizontally until it hits a
// block, an orb or the player.
type Bolt struct {
	CollideActor
	directionX float64
	active     bool
}

func NewBolt(x, y, dirX float64) *Bolt {
	b := &Bolt{directionX: dirX, active: true}
	b.Actor = newActor("blank", x, y, AnchorCentre)
	return b
}

func (b *Bolt) Update(g *Game) {
	if b.move(g, b.directionX, 0, boltSpeed) {
		// Collided with a block.
		b.active = false
	} else {
		// Check for a collision with an orb or the player.
		for _, o := range g.orbs {
			if o.hitTest(g, b) {
				b.active = false
				break
			}
		}
		if b.active && g.player != nil && g.player.hitTest(g, b) {
			b.active = false
		}
	}

	dirIdx := "0"
	if b.directionX > 0 {
		dirIdx = "1"
	}
	b.Image = "bolt" + dirIdx + strconv.Itoa((g.timer/4)%2)
}

func (b *Bolt) Draw(g *Game) { b.drawImage(g.assets) }
