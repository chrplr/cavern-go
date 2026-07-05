package main

import "strconv"

// Fruit is a pickup: normal fruit gives score, and orbs holding the dangerous
// robot type can yield extra health or an extra life.
type Fruit struct {
	GravityActor
	ftype      int
	timeToLive int
}

func NewFruit(x, y float64, trappedEnemyType int) *Fruit {
	f := &Fruit{timeToLive: 500}
	f.Actor = newActor("blank", x, y, AnchorCentreBottom)

	if trappedEnemyType == RobotTypeNormal {
		f.ftype = choiceInt([]int{FruitApple, FruitRaspberry, FruitLemon})
	} else {
		// A dangerous enemy: mostly ordinary fruit, sometimes extra health, and
		// rarely an extra life (proportions match the Python weighted list).
		var types []int
		for i := 0; i < 10; i++ {
			types = append(types, FruitApple, FruitRaspberry, FruitLemon)
		}
		for i := 0; i < 9; i++ {
			types = append(types, FruitExtraHealth)
		}
		types = append(types, FruitExtraLife)
		f.ftype = choiceInt(types)
	}
	return f
}

func (f *Fruit) Update(g *Game) {
	f.gravUpdate(g, true)

	// Collected by the player?
	if g.player != nil {
		cx, cy := f.centerPoint(g.assets)
		if g.player.collidepoint(g.assets, cx, cy) {
			switch f.ftype {
			case FruitExtraHealth:
				g.player.health = minInt(3, g.player.health+1)
				g.playSound("bonus", 1)
			case FruitExtraLife:
				g.player.lives++
				g.playSound("bonus", 1)
			default:
				g.player.score += (f.ftype + 1) * 100
				g.playSound("score", 1)
			}
			f.timeToLive = 0
		} else {
			f.timeToLive--
		}
	} else {
		f.timeToLive--
	}

	if f.timeToLive <= 0 {
		g.pops = append(g.pops, NewPop(f.X, f.Y-27, 0))
	}

	animFrame := []int{0, 1, 2, 1}[(g.timer/6)%4]
	f.Image = "fruit" + strconv.Itoa(f.ftype) + strconv.Itoa(animFrame)
}

func (f *Fruit) Draw(g *Game) { f.drawImage(g.assets) }

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
