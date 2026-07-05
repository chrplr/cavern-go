package main

import (
	"flag"
	"fmt"

	"github.com/chrplr/pgzgo"
)

type State int

const (
	StateMenu State = iota
	StatePlay
	StateGameOver
)

var (
	state  State
	game   *Game
	assets *Assets
	audio  *Audio
)

func update() {
	switch state {
	case StateMenu:
		if spacePressed() {
			state = StatePlay
			game = NewGame(NewPlayer(), assets, audio)
		} else {
			game.Update()
		}

	case StatePlay:
		if game.player.lives < 0 {
			game.playSound("over", 1)
			state = StateGameOver
		} else {
			game.Update()
		}

	case StateGameOver:
		if spacePressed() {
			state = StateMenu
			game = NewGame(nil, assets, audio)
		}
	}
}

func draw() {
	game.Draw()

	switch state {
	case StateMenu:
		assets.Blit("title", 0, 0)
		// "Press SPACE" animation: 10 frames, holding on frame 9 most of the time.
		animFrame := ((game.timer + 40) % 160) / 4
		if animFrame > 9 {
			animFrame = 9
		}
		assets.Blit("space"+itoa(animFrame), 130, 280)

	case StatePlay:
		drawStatus(assets, game)

	case StateGameOver:
		drawStatus(assets, game)
		assets.Blit("over", 0, 0)
	}
}

func main() {
	selftest := flag.Bool("selftest", false, "run several levels headlessly, then exit")
	flag.Parse()

	a, err := pgzgo.New(pgzgo.Config{
		Title:  "Cavern",
		Width:  Width,
		Height: Height,
		Images: imagesFS,
		Audio:  audioFS,
	})
	if err != nil {
		panic(err)
	}
	defer a.Close()

	app = a
	assets = a.Screen
	audio = a.Audio

	if *selftest {
		g := NewGame(NewPlayer(), assets, audio)
		// Free-running phase: exercises robots, bolts, fruit, gravity, collisions.
		for step := 0; step < 1500; step++ {
			g.Update()
		}
		// Orb phase: blow orbs and let them trap the robots, then float and pop.
		for step := 0; step < 1500; step++ {
			if step%30 == 0 && len(g.orbs) < 5 {
				o := NewOrb(g.player.X, g.player.Y-35, 1)
				g.orbs = append(g.orbs, o)
			}
			g.Update()
		}
		// Level-cycle phase: load and step every level layout in turn.
		for lvl := 0; lvl < len(LEVELS)*2; lvl++ {
			g.nextLevel()
			for step := 0; step < 120; step++ {
				g.Update()
			}
			fmt.Printf("level %d: %d grid rows, %d enemies, %d pending, %d fruits\n",
				g.level, len(g.grid), len(g.enemies), len(g.pendingEnemies), len(g.fruits))
		}
		// Verify embedded assets actually decode into real textures/sounds.
		loaded, total := 0, 0
		for _, n := range []string{"title", "over", "life", "block0", "still", "orb0", "space0"} {
			total++
			if w, h := assets.Size(n); w > 0 && h > 0 {
				loaded++
			}
		}
		fmt.Printf("embedded textures decoded: %d/%d; embedded sounds: %d\n", loaded, total, audio.SoundCount())
		fmt.Printf("SELFTEST OK: score %d, lives %d, health %d\n",
			g.player.score, g.player.lives, g.player.health)
		return
	}

	audio.PlayMusic("theme", 0.3)

	state = StateMenu
	game = NewGame(nil, assets, audio)

	a.Loop(
		func(*pgzgo.App) { update() },
		func(*pgzgo.App) { draw() },
	)
}
