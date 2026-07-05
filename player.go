package main

import "strconv"

// Player is the controllable character: runs, jumps and blows orbs.
type Player struct {
	GravityActor
	lives int
	score int

	directionX float64
	fireTimer  int
	hurtTimer  int
	health     int
	blowingOrb *Orb
}

func NewPlayer() *Player {
	p := &Player{lives: 2, score: 0}
	p.Actor = newActor("blank", 0, 0, AnchorCentreBottom)
	return p
}

func (p *Player) reset() {
	p.X, p.Y = Width/2, 100
	p.velY = 0
	p.directionX = 1
	p.fireTimer = 0
	p.hurtTimer = 100 // Invulnerable for this many frames.
	p.health = 3
	p.blowingOrb = nil
}

// hitTest checks for a collision with a bolt and applies damage/knockback.
func (p *Player) hitTest(g *Game, b *Bolt) bool {
	if p.collidepoint(g.assets, b.X, b.Y) && p.hurtTimer < 0 {
		p.hurtTimer = 200
		p.health--
		p.velY = -12
		p.landed = false
		p.directionX = b.directionX
		if p.health > 0 {
			g.playSound("ouch", 4)
		} else {
			g.playSound("die", 1)
		}
		return true
	}
	return false
}

func (p *Player) Update(g *Game) {
	// Fall out of the level (no block collision) once health hits zero.
	p.gravUpdate(g, p.health > 0)

	p.fireTimer--
	p.hurtTimer--

	if p.landed {
		p.hurtTimer = minInt(p.hurtTimer, 100)
	}

	dx := 0.0
	if p.hurtTimer > 100 {
		// Just been hurt: either the sideways knockback, or (if dead) drop out of
		// the level and respawn once far enough down.
		if p.health > 0 {
			p.move(g, p.directionX, 0, 4)
		} else if p.top(g.assets) >= Height*1.5 {
			p.lives--
			p.reset()
		}
	} else {
		if keyLeft() {
			dx = -1
		} else if keyRight() {
			dx = 1
		}

		if dx != 0 {
			p.directionX = dx
			if p.fireTimer < 10 {
				p.move(g, dx, 0, 4)
			}
		}

		// Blow a new orb: space pressed, cooldown elapsed, at most 5 orbs.
		if spacePressed() && p.fireTimer <= 0 && len(g.orbs) < 5 {
			x := p.X + p.directionX*38
			if x < 70 {
				x = 70
			} else if x > 730 {
				x = 730
			}
			y := p.Y - 35
			p.blowingOrb = NewOrb(x, y, p.directionX)
			g.orbs = append(g.orbs, p.blowingOrb)
			g.playSound("blow", 4)
			p.fireTimer = 20
		}

		if keyUp() && p.velY == 0 && p.landed {
			p.velY = -16
			p.landed = false
			g.playSound("jump", 1)
		}
	}

	// Holding space keeps blowing the current orb further.
	if keySpace() {
		if p.blowingOrb != nil {
			p.blowingOrb.blownFrames += 4
			if p.blowingOrb.blownFrames >= 120 {
				p.blowingOrb = nil
			}
		}
	} else {
		p.blowingOrb = nil
	}

	// Choose sprite. When hurt, the sprite flashes on alternate frames.
	p.Image = "blank"
	if p.hurtTimer <= 0 || p.hurtTimer%2 == 1 {
		dirIndex := "0"
		if p.directionX > 0 {
			dirIndex = "1"
		}
		switch {
		case p.hurtTimer > 100:
			if p.health > 0 {
				p.Image = "recoil" + dirIndex
			} else {
				p.Image = "fall" + strconv.Itoa((g.timer/4)%2)
			}
		case p.fireTimer > 0:
			p.Image = "blow" + dirIndex
		case dx == 0:
			p.Image = "still"
		default:
			p.Image = "run" + dirIndex + strconv.Itoa((g.timer/8)%4)
		}
	}
}

func (p *Player) Draw(g *Game) { p.drawImage(g.assets) }
