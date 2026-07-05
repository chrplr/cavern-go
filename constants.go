package main

const (
	Width  = 800
	Height = 480

	NumRows    = 18
	NumColumns = 28

	LevelXOffset  = 50
	GridBlockSize = 25

	targetFPS   = 60
	frameMillis = 1000 / targetFPS
)

// LEVELS holds the fixed level layouts. Each 'X' is a solid block, each space is
// empty. Every level has 17 rows; Game.nextLevel appends a copy of the first row
// as an 18th row (the floor), matching the Python original.
var LEVELS = [][]string{
	{
		"XXXXX     XXXXXXXX     XXXXX",
		"", "", "", "",
		"   XXXXXXX        XXXXXXX   ",
		"", "", "",
		"   XXXXXXXXXXXXXXXXXXXXXX   ",
		"", "", "",
		"XXXXXXXXX          XXXXXXXXX",
		"", "", "",
	},
	{
		"XXXX    XXXXXXXXXXXX    XXXX",
		"", "", "", "",
		"    XXXXXXXXXXXXXXXXXXXX    ",
		"", "", "",
		"XXXXXX                XXXXXX",
		"      X              X      ",
		"       X            X       ",
		"        X          X        ",
		"         X        X         ",
		"", "", "",
	},
	{
		"XXXX    XXXX    XXXX    XXXX",
		"", "", "", "",
		"  XXXXXXXX        XXXXXXXX  ",
		"", "", "",
		"XXXX      XXXXXXXX      XXXX",
		"", "", "",
		"    XXXXXX        XXXXXX    ",
		"", "", "",
	},
}

// Robot types.
const (
	RobotTypeNormal     = 0
	RobotTypeAggressive = 1
)

// Fruit types.
const (
	FruitApple       = 0
	FruitRaspberry   = 1
	FruitLemon       = 2
	FruitExtraHealth = 3
	FruitExtraLife   = 4
)

// Widths of the letters A to Z in the font images. For any other character
// (space, digits) the width of 'A' is used.
var charWidthTable = [26]int{
	27, 26, 25, 26, 25, 25, 26, 25, 12, 26, 26, 25, 33, 25, 26,
	25, 27, 26, 26, 25, 26, 26, 38, 25, 25, 25,
}
