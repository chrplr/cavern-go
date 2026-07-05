package main

import "strconv"

const orbMaxTimer = 250

// Orb is a bubble blown by the player. It drifts horizontally, then floats up.
// It can trap an enemy; when it pops a trapped enemy becomes fruit.
type Orb struct {
	CollideActor
	directionX       float64
	floating         bool
	trappedEnemyType int // -1 = none, else the trapped robot's type (0 or 1)
	timer            int
	blownFrames      int
}

func NewOrb(x, y, dirX float64) *Orb {
	o := &Orb{
		directionX:       dirX,
		trappedEnemyType: -1,
		timer:            -1,
		blownFrames:      6,
	}
	o.Actor = newActor("blank", x, y, AnchorCentre)
	return o
}

// hitTest checks for a collision with a bolt; a hit fast-forwards the orb to pop.
func (o *Orb) hitTest(g *Game, b *Bolt) bool {
	collided := o.collidepoint(g.assets, b.X, b.Y)
	if collided {
		o.timer = orbMaxTimer - 1
	}
	return collided
}

func (o *Orb) Update(g *Game) {
	o.timer++

	if o.floating {
		// Float upwards.
		o.move(g, 0, -1, float64(randInt(1, 2)))
	} else {
		// Move horizontally; start floating on hitting a block.
		if o.move(g, o.directionX, 0, 4) {
			o.floating = true
		}
	}

	if o.timer == o.blownFrames {
		o.floating = true
	} else if o.timer >= orbMaxTimer || o.Y <= -40 {
		// Pop when our lifetime runs out or we leave the top of the screen.
		g.pops = append(g.pops, NewPop(o.X, o.Y, 1))
		if o.trappedEnemyType != -1 {
			g.fruits = append(g.fruits, NewFruit(o.X, o.Y, o.trappedEnemyType))
		}
		g.playSound("pop", 4)
	}

	if o.timer < 9 {
		// Grow to full size over 9 frames (a new frame every 3).
		o.Image = "orb" + strconv.Itoa(o.timer/3)
	} else if o.trappedEnemyType != -1 {
		o.Image = "trap" + strconv.Itoa(o.trappedEnemyType) + strconv.Itoa((o.timer/4)%8)
	} else {
		o.Image = "orb" + strconv.Itoa(3+(((o.timer-9)/8)%4))
	}
}

func (o *Orb) Draw(g *Game) { o.drawImage(g.assets) }
