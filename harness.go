package main

// Glue between the game and the pgzgo harness: the embedded assets (the
// //go:embed directives must live in this package) and the input helpers the
// game code calls, adapted onto the harness keyboard snapshot.

import (
	"embed"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/chrplr/pgzgo"
)

// Assets and Audio are the game's names for the harness drawing surface and mixer.
type Assets = pgzgo.Screen
type Audio = pgzgo.Audio

//go:embed images
var imagesFS embed.FS

//go:embed sounds music
var audioFS embed.FS

// app is the running harness; the input wrappers read from its keyboard snapshot.
var app *pgzgo.App

// Held-key helpers mirroring Pygame Zero's keyboard.left / .right / .up / .space.
func keyLeft() bool  { return app.Keyboard.Held(sdl.SCANCODE_LEFT) }
func keyRight() bool { return app.Keyboard.Held(sdl.SCANCODE_RIGHT) }
func keyUp() bool    { return app.Keyboard.Held(sdl.SCANCODE_UP) }
func keySpace() bool { return app.Keyboard.Held(sdl.SCANCODE_SPACE) }

// spacePressed reports whether space was just pressed this frame (rising edge).
func spacePressed() bool { return app.Keyboard.Pressed(sdl.SCANCODE_SPACE) }
