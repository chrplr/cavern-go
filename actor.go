package main

// Anchor kinds per axis.
const (
	akCenter = 0
	akBottom = 1
)

// Anchor mirrors Pygame Zero's anchor tuples. Cavern only needs centre and
// centre-bottom.
type Anchor struct {
	yKind int
}

var (
	AnchorCentre       = Anchor{akCenter}
	AnchorCentreBottom = Anchor{akBottom}
)

// offset returns the pixel offset from the actor's (X, Y) anchor position to the
// sprite's top-left corner.
func (an Anchor) offset(w, h float64) (float64, float64) {
	ax := w / 2
	ay := h / 2
	if an.yKind == akBottom {
		ay = h
	}
	return ax, ay
}

// Actor is a positioned sprite with an anchor - the base for every game object.
type Actor struct {
	X, Y   float64
	Image  string
	anchor Anchor
}

func newActor(image string, x, y float64, anchor Anchor) Actor {
	return Actor{X: x, Y: y, Image: image, anchor: anchor}
}

func (a *Actor) anchorOffset(as *Assets) (float64, float64) {
	w, h := as.Size(a.Image)
	return a.anchor.offset(w, h)
}

func (a *Actor) spriteRect(as *Assets) Rect {
	ax, ay := a.anchorOffset(as)
	w, h := as.Size(a.Image)
	return Rect{a.X - ax, a.Y - ay, w, h}
}

func (a *Actor) top(as *Assets) float64 {
	_, ay := a.anchorOffset(as)
	return a.Y - ay
}

func (a *Actor) bottom(as *Assets) float64 {
	_, ay := a.anchorOffset(as)
	w, h := as.Size(a.Image)
	_ = w
	return a.Y - ay + h
}

func (a *Actor) centerPoint(as *Assets) (float64, float64) {
	r := a.spriteRect(as)
	return r.X + r.W/2, r.Y + r.H/2
}

func (a *Actor) collidepoint(as *Assets, px, py float64) bool {
	return a.spriteRect(as).collidepoint(px, py)
}

// drawImage blits the sprite at its anchored screen position.
func (a *Actor) drawImage(as *Assets) {
	ax, ay := a.anchorOffset(as)
	as.Blit(a.Image, a.X-ax, a.Y-ay)
}

// CollideActor moves through the level one pixel at a time, stopping at blocks
// and the level edges.
type CollideActor struct {
	Actor
}

// move steps up to speed pixels along (dx, dy). One of dx/dy is zero. Returns
// true if a block or the level edge blocked the movement.
func (c *CollideActor) move(g *Game, dx, dy, speed float64) bool {
	newX, newY := float64(int(c.X)), float64(int(c.Y))
	idx, idy := int(dx), int(dy)

	for i := 0; i < int(speed); i++ {
		newX += dx
		newY += dy

		if newX < 70 || newX > 730 {
			// Collided with edge of level.
			return true
		}

		nx, ny := int(newX), int(newY)
		if (idy > 0 && pmod(ny, GridBlockSize) == 0 ||
			idx > 0 && pmod(nx, GridBlockSize) == 0 ||
			idx < 0 && pmod(nx, GridBlockSize) == GridBlockSize-1) &&
			g.block(newX, newY) {
			return true
		}

		c.X, c.Y = newX, newY
	}
	return false
}

// GravityActor is a CollideActor subject to gravity (player, robots, fruit).
type GravityActor struct {
	CollideActor
	velY   float64
	landed bool
}

func (ga *GravityActor) gravUpdate(g *Game, detect bool) {
	ga.velY = minf(ga.velY+1, 10)

	if detect {
		if ga.move(g, 0, sign(ga.velY), absf(ga.velY)) {
			// Landed on a block. move applies no collision when moving up.
			ga.velY = 0
			ga.landed = true
		}
		if ga.top(g.assets) >= Height {
			// Fallen off the bottom - reappear at the top.
			ga.Y = 1
		}
	} else {
		ga.Y += ga.velY
	}
}

func minf(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
