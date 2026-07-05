package main

import "strconv"

// charWidthOf returns the pixel width of a character. For anything other than
// A-Z (space, digits) the width of 'A' is used, matching the Python original.
func charWidthOf(c rune) int {
	index := int(c) - 65
	if index < 0 {
		index = 0
	}
	if index >= len(charWidthTable) {
		index = 0
	}
	return charWidthTable[index]
}

func textWidth(text string) int {
	total := 0
	for _, c := range text {
		total += charWidthOf(c)
	}
	return total
}

// drawTextX draws text with its left edge at x.
func drawTextX(as *Assets, text string, x, y float64) {
	for _, c := range text {
		as.Blit("font0"+strconv.Itoa(int(c)), x, y)
		x += float64(charWidthOf(c))
	}
}

// drawTextCentre draws text centred horizontally on the screen.
func drawTextCentre(as *Assets, text string, y float64) {
	x := float64((Width - textWidth(text)) / 2)
	drawTextX(as, text, x, y)
}

var imageWidth = map[string]float64{"life": 44, "plus": 40, "health": 40}

// drawStatus draws the score, level number, lives and health.
func drawStatus(as *Assets, g *Game) {
	// Score, right-justified.
	numberWidth := charWidthTable[0]
	s := itoa(g.player.score)
	drawTextX(as, s, float64(Width-2-(numberWidth*len(s))), 451)

	// Level number, centred.
	drawTextCentre(as, "LEVEL "+itoa(g.level+1), 451)

	// Lives and health. At most two life icons are shown; a plus means more.
	var livesHealth []string
	for i := 0; i < minInt(2, g.player.lives); i++ {
		livesHealth = append(livesHealth, "life")
	}
	if g.player.lives > 2 {
		livesHealth = append(livesHealth, "plus")
	}
	if g.player.lives >= 0 {
		for i := 0; i < g.player.health; i++ {
			livesHealth = append(livesHealth, "health")
		}
	}

	x := 0.0
	for _, image := range livesHealth {
		as.Blit(image, x, 450)
		x += imageWidth[image]
	}
}

func itoa(i int) string { return strconv.Itoa(i) }
