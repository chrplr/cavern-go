package main

import "math"

// Rect is an axis-aligned rectangle with pygame-Rect-style collision tests.
type Rect struct {
	X, Y, W, H float64
}

// collidepoint reports whether (px, py) lies inside the rectangle (left/top
// inclusive, right/bottom exclusive), matching pygame.Rect.collidepoint.
func (r Rect) collidepoint(px, py float64) bool {
	return px >= r.X && px < r.X+r.W && py >= r.Y && py < r.Y+r.H
}

// sign returns -1 for negative numbers and 1 otherwise (note: sign(0) == 1),
// matching the Python original.
func sign(x float64) float64 {
	if x < 0 {
		return -1
	}
	return 1
}

// floorDiv is Python-style floor division for ints.
func floorDiv(a, b int) int {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}

// pmod is Python-style modulo (result has the sign of the divisor).
func pmod(a, b int) int {
	m := a % b
	if m != 0 && ((m < 0) != (b < 0)) {
		m += b
	}
	return m
}

func absf(x float64) float64 { return math.Abs(x) }
