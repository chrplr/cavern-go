package main

import "strconv"

// Robot is an enemy. The aggressive type also shoots at orbs.
type Robot struct {
	GravityActor
	rtype          int
	speed          float64
	directionX     float64
	alive          bool
	changeDirTimer int
	fireTimer      int
}

func NewRobot(x, y float64, rtype int) *Robot {
	r := &Robot{
		rtype:          rtype,
		speed:          float64(randInt(1, 3)),
		directionX:     1,
		alive:          true,
		changeDirTimer: 0,
		fireTimer:      100,
	}
	r.Actor = newActor("blank", x, y, AnchorCentreBottom)
	return r
}

func (r *Robot) Update(g *Game) {
	r.gravUpdate(g, true)

	r.changeDirTimer--
	r.fireTimer++

	// Move in the current direction; turn around on hitting a wall.
	if r.move(g, r.directionX, 0, r.speed) {
		r.changeDirTimer = 0
	}

	if r.changeDirTimer <= 0 {
		// Pick a direction; with a player present, two of the three choices head
		// towards them.
		directions := []int{-1, 1}
		if g.player != nil {
			directions = append(directions, int(sign(g.player.X-r.X)))
		}
		r.directionX = float64(choiceInt(directions))
		r.changeDirTimer = randInt(100, 250)
	}

	// The aggressive type can deliberately shoot at orbs, turning to face them.
	if r.rtype == RobotTypeAggressive && r.fireTimer >= 24 {
		for _, orb := range g.orbs {
			if orb.Y >= r.top(g.assets) && orb.Y < r.bottom(g.assets) && absf(orb.X-r.X) < 200 {
				r.directionX = sign(orb.X - r.X)
				r.fireTimer = 0
				break
			}
		}
	}

	// Fire at the player.
	if r.fireTimer >= 12 {
		fireProbability := g.fireProbability()
		if g.player != nil && r.top(g.assets) < g.player.bottom(g.assets) && r.bottom(g.assets) > g.player.top(g.assets) {
			fireProbability *= 10
		}
		if randFloat() < fireProbability {
			r.fireTimer = 0
			g.playSound("laser", 4)
		}
	} else if r.fireTimer == 8 {
		// Frame 8 of the firing animation is when the bolt actually appears.
		g.bolts = append(g.bolts, NewBolt(r.X+r.directionX*20, r.Y-38, r.directionX))
	}

	// Colliding with an untrapped orb traps us inside it.
	for _, orb := range g.orbs {
		cx, cy := orb.centerPoint(g.assets)
		if orb.trappedEnemyType == -1 && r.collidepoint(g.assets, cx, cy) {
			r.alive = false
			orb.floating = true
			orb.trappedEnemyType = r.rtype
			g.playSound("trap", 4)
			break
		}
	}

	// Choose sprite.
	dirIdx := "0"
	if r.directionX > 0 {
		dirIdx = "1"
	}
	image := "robot" + strconv.Itoa(r.rtype) + dirIdx
	if r.fireTimer < 12 {
		image += strconv.Itoa(5 + (r.fireTimer / 4))
	} else {
		image += strconv.Itoa(1 + ((g.timer / 4) % 4))
	}
	r.Image = image
}

func (r *Robot) Draw(g *Game) { r.drawImage(g.assets) }
