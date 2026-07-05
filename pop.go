package main

import "strconv"

// Pop is a short-lived burst animation (from popped orbs and expiring fruit).
type Pop struct {
	Actor
	ptype int
	timer int
}

func NewPop(x, y float64, ptype int) *Pop {
	p := &Pop{ptype: ptype, timer: -1}
	p.Actor = newActor("blank", x, y, AnchorCentre)
	return p
}

func (p *Pop) Update(g *Game) {
	p.timer++
	p.Image = "pop" + strconv.Itoa(p.ptype) + strconv.Itoa(p.timer/2)
}

func (p *Pop) Draw(g *Game) { p.drawImage(g.assets) }
