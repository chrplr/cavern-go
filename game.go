package main

// Game holds the level state and all active objects.
type Game struct {
	player      *Player
	levelColour int
	level       int
	grid        []string
	timer       int

	fruits         []*Fruit
	bolts          []*Bolt
	enemies        []*Robot
	pops           []*Pop
	orbs           []*Orb
	pendingEnemies []int

	assets *Assets
	audio  *Audio
}

func NewGame(player *Player, assets *Assets, audio *Audio) *Game {
	g := &Game{
		player:      player,
		levelColour: -1,
		level:       -1,
		assets:      assets,
		audio:       audio,
	}
	g.nextLevel()
	return g
}

// block reports whether there is a solid level block at pixel (x, y).
func (g *Game) block(x, y float64) bool {
	gridX := floorDiv(int(x)-LevelXOffset, GridBlockSize)
	gridY := floorDiv(int(y), GridBlockSize)
	if gridY > 0 && gridY < NumRows {
		row := g.grid[gridY]
		return gridX >= 0 && gridX < NumColumns && len(row) > 0 && gridX < len(row) && row[gridX] != ' '
	}
	return false
}

// fireProbability is the per-frame chance of each robot firing (rises by level).
func (g *Game) fireProbability() float64 {
	return 0.001 + (0.0001 * float64(minInt(100, g.level)))
}

// maxEnemies is the cap on simultaneous on-screen enemies (rises by level).
func (g *Game) maxEnemies() int {
	return minInt((g.level+6)/2, 8)
}

func (g *Game) nextLevel() {
	g.levelColour = (g.levelColour + 1) % 4
	g.level++

	// Build the grid: the level rows plus a copy of the first row as the floor.
	// Copy into a fresh slice so we never mutate the LEVELS data.
	base := LEVELS[g.level%len(LEVELS)]
	g.grid = make([]string, 0, len(base)+1)
	g.grid = append(g.grid, base...)
	g.grid = append(g.grid, base[0])

	g.timer = -1

	if g.player != nil {
		g.player.reset()
	}

	g.fruits = nil
	g.bolts = nil
	g.enemies = nil
	g.pops = nil
	g.orbs = nil

	// Build the list of pending enemies (0 = normal, 1 = aggressive), then shuffle.
	numEnemies := 10 + g.level
	numStrong := 1 + int(float64(g.level)/1.5)
	numWeak := numEnemies - numStrong
	g.pendingEnemies = nil
	for i := 0; i < numStrong; i++ {
		g.pendingEnemies = append(g.pendingEnemies, RobotTypeAggressive)
	}
	for i := 0; i < numWeak; i++ {
		g.pendingEnemies = append(g.pendingEnemies, RobotTypeNormal)
	}
	shuffleInt(g.pendingEnemies)

	g.playSound("level", 1)
}

// getRobotSpawnX finds an x for a new robot by scanning the top row for a gap.
func (g *Game) getRobotSpawnX() float64 {
	r := randInt(0, NumColumns-1)
	top := g.grid[0]
	for i := 0; i < NumColumns; i++ {
		gridX := (r + i) % NumColumns
		if gridX < len(top) && top[gridX] == ' ' {
			return float64(GridBlockSize*gridX + LevelXOffset + 12)
		}
	}
	return Width / 2
}

func (g *Game) Update() {
	g.timer++

	// Update all objects.
	for _, f := range g.fruits {
		f.Update(g)
	}
	for _, b := range g.bolts {
		b.Update(g)
	}
	for _, e := range g.enemies {
		e.Update(g)
	}
	for _, p := range g.pops {
		p.Update(g)
	}
	if g.player != nil {
		g.player.Update(g)
	}
	for _, o := range g.orbs {
		o.Update(g)
	}

	// Remove objects that are no longer wanted.
	g.fruits = filterFruits(g.fruits, func(f *Fruit) bool { return f.timeToLive > 0 })
	g.bolts = filterBolts(g.bolts, func(b *Bolt) bool { return b.active })
	g.enemies = filterRobots(g.enemies, func(e *Robot) bool { return e.alive })
	g.pops = filterPops(g.pops, func(p *Pop) bool { return p.timer < 12 })
	g.orbs = filterOrbs(g.orbs, func(o *Orb) bool { return o.timer < 250 && o.Y > -40 })

	// Every 100 frames spawn a random fruit (unless the level has no enemies left).
	if g.timer%100 == 0 && len(g.pendingEnemies)+len(g.enemies) > 0 {
		g.fruits = append(g.fruits, NewFruit(float64(randInt(70, 730)), float64(randInt(75, 400)), 0))
	}

	// Every 81 frames spawn a robot from the pending list, up to the cap.
	if g.timer%81 == 0 && len(g.pendingEnemies) > 0 && len(g.enemies) < g.maxEnemies() {
		robotType := g.pendingEnemies[len(g.pendingEnemies)-1]
		g.pendingEnemies = g.pendingEnemies[:len(g.pendingEnemies)-1]
		g.enemies = append(g.enemies, NewRobot(g.getRobotSpawnX(), -30, robotType))
	}

	// End the level when nothing is left to do (and no orbs still trap an enemy).
	if len(g.pendingEnemies)+len(g.fruits)+len(g.enemies)+len(g.pops) == 0 {
		trapped := 0
		for _, o := range g.orbs {
			if o.trappedEnemyType != -1 {
				trapped++
			}
		}
		if trapped == 0 {
			g.nextLevel()
		}
	}
}

func (g *Game) Draw() {
	// Background for this level.
	g.assets.Blit("bg"+itoa(g.levelColour), 0, 0)

	// Blocks.
	blockSprite := "block" + itoa(g.level%4)
	for rowY := 0; rowY < NumRows; rowY++ {
		row := g.grid[rowY]
		if len(row) > 0 {
			x := float64(LevelXOffset)
			for _, c := range row {
				if c != ' ' {
					g.assets.Blit(blockSprite, x, float64(rowY*GridBlockSize))
				}
				x += GridBlockSize
			}
		}
	}

	// Objects (player drawn last).
	for _, f := range g.fruits {
		f.Draw(g)
	}
	for _, b := range g.bolts {
		b.Draw(g)
	}
	for _, e := range g.enemies {
		e.Draw(g)
	}
	for _, p := range g.pops {
		p.Draw(g)
	}
	for _, o := range g.orbs {
		o.Draw(g)
	}
	if g.player != nil {
		g.player.Draw(g)
	}
}

// playSound plays a sound, but only when a player exists (silent on the menu).
func (g *Game) playSound(name string, count int) {
	if g.player == nil {
		return
	}
	g.audio.PlaySound(name, count)
}

// Typed slice filters (Go has no generic list comprehension in this codebase style).
func filterFruits(s []*Fruit, keep func(*Fruit) bool) []*Fruit {
	var out []*Fruit
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}
func filterBolts(s []*Bolt, keep func(*Bolt) bool) []*Bolt {
	var out []*Bolt
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}
func filterRobots(s []*Robot, keep func(*Robot) bool) []*Robot {
	var out []*Robot
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}
func filterPops(s []*Pop, keep func(*Pop) bool) []*Pop {
	var out []*Pop
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}
func filterOrbs(s []*Orb, keep func(*Orb) bool) []*Orb {
	var out []*Orb
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}
