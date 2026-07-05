# Cavern — Python vs. Go implementation comparison

This document analyses how the Go port in this folder relates to the original
`cavern.py`. It covers the structural mapping, the language‑paradigm differences
that shaped the port, the framework substitutions, and a set of subtle
numeric/semantic details that had to be reproduced exactly for the game to
behave the same way.

The goal throughout the port was **behavioural fidelity**: the Go code is a
faithful translation of the game logic, deviating only where a language or
library difference forces a different expression of the same idea, or where a
platform feature (game controllers) is out of scope for the port.

Cavern is the simplest of the games ported in this repo: a single‑screen
*Bubble Bobble*‑style platformer with hardcoded level layouts, no scrolling, no
tilemaps, no save files. The Python original is ~780 lines in one file; the Go
port is ~1,050 lines across 16 focused files. The extra volume is almost
entirely boilerplate that Python gets for free: explicit struct definitions,
per‑type slice filters, and the asset/sound plumbing that Pygame Zero hides.

---

## 1. High‑level architecture

Both versions share the same conceptual design:

- A **single‑screen platformer**. The level is a small ASCII grid of blocks;
  the player runs, jumps, and blows **orbs** (bubbles) that trap robots. A
  trapped robot floats up in its orb; popping the orb turns it into fruit, which
  the player collects for score, extra health, or extra lives.
- An **actor hierarchy** — `Actor → CollideActor → GravityActor → Player/Robot/
  Fruit` — where each layer adds behaviour (anchored drawing → pixel‑stepped
  block collision → gravity).
- A **menu → play → game‑over** state machine, where the menu runs a
  player‑less game in the background as an attract screen.
- **Endless levels**: three layouts cycle forever, getting harder (more enemies,
  more aggressive ones, higher fire rate).

The two largest pieces of logic in both — `Player.update`/`(*Player).Update` (the
movement/orb/jump state machine) and `Game.update`/`(*Game).Update` (spawning,
culling, and level‑end detection) — are ported statement‑for‑statement.

### File layout

| Concern | Python | Go |
|---|---|---|
| Constants / levels / tables | top of `cavern.py` | `constants.go` |
| Rect, sign, floor div | inline helpers | `util.go` |
| Random helpers | `random.*` | `rng.go` |
| Actor / Collide / Gravity actors | those classes | `actor.go` |
| Orb / Bolt / Pop | those classes | `orb.go`, `bolt.go`, `pop.go` |
| Fruit / Player / Robot | those classes | `fruit.go`, `player.go`, `robot.go` |
| Game (levels, spawning, draw) | `Game` | `game.go` |
| Sprite font + status bar | `draw_text`/`draw_status` | `text.go` |
| Input | `keyboard.*`/`space_pressed` | `input.go` |
| Assets | Pygame Zero `images`/`screen.blit` | `assets.go` |
| Audio | Pygame Zero `sounds`/`music` | `audio.go` |
| State machine / entry point | `update`/`draw`/module code | `main.go` |

---

## 2. Language paradigm: classes/inheritance → structs/embedding

### Python: classical inheritance

```python
class CollideActor(Actor):
    def move(self, dx, dy, speed): ...
class GravityActor(CollideActor):
    def update(self, detect=True): ...
class Player(GravityActor):
    def update(self):
        super().update(self.health > 0)   # dynamic dispatch + super()
        ...
```

### Go: struct embedding + explicit forwarding

Go has no inheritance, so the port uses **struct embedding** for code reuse.
The chain becomes embedded fields:

```go
type CollideActor struct{ Actor }
type GravityActor struct {
    CollideActor
    velY   float64
    landed bool
}
type Player struct {
    GravityActor
    lives, score int
    ...
}
```

`super().update(self.health > 0)` becomes an **explicit call to the embedded
method**: `p.gravUpdate(g, p.health > 0)`. Each concrete type gets its own
`Update(g *Game)` and `Draw(g *Game)`.

Notably, **Cavern needs no `self` back‑reference interface** (unlike the Eggzy
port). In Eggzy, the base `get_rect` had to call a subclass‑overridden
`get_collidable_width`; here `CollideActor.move` only ever touches the actor's
own position and calls `game.block`, with no virtual method going back down to
the concrete type. So plain embedding suffices and the port is simpler.

### Polymorphic update/draw lists

Python builds heterogeneous lists and iterates them polymorphically:

```python
for obj in self.fruits + self.bolts + self.enemies + self.pops + [self.player] + self.orbs:
    if obj: obj.update()
```

Go has no covariant "list of anything with `.Update()`" without declaring an
interface, and each slice here is already a concrete type, so the port unrolls
the concatenation into typed loops in the original order:

```go
for _, f := range g.fruits  { f.Update(g) }
for _, b := range g.bolts   { b.Update(g) }
for _, e := range g.enemies { e.Update(g) }
for _, p := range g.pops    { p.Update(g) }
if g.player != nil          { g.player.Update(g) }
for _, o := range g.orbs    { o.Update(g) }
```

The draw order (`fruits, bolts, enemies, pops, orbs`, then `player` last) is
likewise unrolled verbatim, since it determines what is drawn on top of what.

---

## 3. List comprehensions → typed slice filters

Python culls dead objects with list comprehensions:

```python
self.fruits  = [f for f in self.fruits  if f.time_to_live > 0]
self.bolts   = [b for b in self.bolts   if b.active]
self.enemies = [e for e in self.enemies if e.alive]
self.pops    = [p for p in self.pops    if p.timer < 12]
self.orbs    = [o for o in self.orbs    if o.timer < 250 and o.y > -40]
```

Without generics used in this codebase's style, Go gets one small
`filterX(slice, keep)` helper per element type in `game.go`:

```go
g.fruits = filterFruits(g.fruits, func(f *Fruit) bool { return f.timeToLive > 0 })
g.orbs   = filterOrbs(g.orbs,   func(o *Orb)   bool { return o.timer < 250 && o.Y > -40 })
```

Verbose, but a direct transliteration of each predicate.

---

## 4. The `game` global → an explicit `*Game` parameter

Python reaches a module‑level `game` global from inside every actor method
(`game.player`, `game.play_sound(...)`, `game.orbs.append(...)`,
`game.block(...)` via the free `block` function). The Go port has no such
global; instead **`g *Game` is threaded as a parameter** through every method
that needs it: `(*Player).Update(g)`, `(*Robot).Update(g)`, `(*Orb).hitTest(g,
b)`, and so on. The free function `block(x, y)` becomes a method `(*Game).block`.

This is mechanical but pervasive — nearly every method signature gains a
`g *Game`. It makes the data flow explicit and removes any initialization‑order
ambiguity.

---

## 5. Framework substitution: Pygame Zero → go‑sdl3

| Pygame Zero feature | Go / go‑sdl3 replacement |
|---|---|
| `Actor("name", pos, anchor)` auto‑loads `images/name.png` | `Assets.Texture(name)` lazily loads + caches `*sdl.Texture` |
| `screen.blit(name, (x,y))` | `Assets.Blit` → `renderer.RenderTexture` |
| anchor tuples resolved internally | `Anchor` struct + `offset(w,h)` (§7) |
| `keyboard.left`, `keyboard.space` | `sdl.GetKeyboardState()` snapshot in `input.go` |
| `sounds.pop3.play()` via `getattr` | `Audio.PlaySound(name, count)` (§8) |
| `music.play`/`music.set_volume` | `Audio.PlayMusic` with a looping mixer track |
| the `update()`/`draw()` game loop | explicit `sdl.RunLoop` in `main.go` with a frame delay |
| `draw_text` sprite font | `text.go` `drawTextX`/`drawTextCentre` |

### The game loop and frame timing

Pygame Zero calls `update()` and `draw()` at a fixed 60 Hz. The Go `main.go`
runs the loop itself: poll events → `refreshKeys()` → clear → `update()` →
`draw()` → present, then `sdl.Delay` to pad the frame out to `1000/60` ms. Both
games are frame‑count driven (`game.timer` / `g.timer` increments once per
update; animation frames are `timer // interval`), so the timing model matches
as long as the loop runs at 60 Hz.

---

## 6. Input and the space‑bar edge latch

The trickiest input detail is `space_pressed()`. Pygame Zero exposes only the
*current* key state, so the original detects a fresh press with a module‑level
latch:

```python
space_down = False
def space_pressed():
    global space_down
    if keyboard.space:
        if space_down:  return False   # still held
        space_down = True;  return True
    space_down = False;  return False
```

The Go port reproduces this **exactly**, latch and all, in `input.go`:

```go
var spaceDown bool
func spacePressed() bool {
    if keySpace() {
        if spaceDown { return false }
        spaceDown = true
        return true
    }
    spaceDown = false
    return false
}
```

This is a case where using the port's own edge‑detection machinery would have
been *wrong*. `space_pressed()` is designed to be called **at most once per
frame** because each call mutates the latch — and there is a genuine quirk that
depends on it: inside `Player.update` the call sits in the *not‑hurt* branch, so
during the frames when the player is recoiling (`hurt_timer > 100`) the latch is
**not updated at all**. Reproducing the single‑call‑per‑frame discipline (menu,
game‑over, and the player's not‑hurt branch each call it once) keeps that
behaviour faithful. Held keys (`keyboard.left/right/up/space`) map to simple
`keyDown` queries against the SDL snapshot.

---

## 7. The anchor system

Pygame Zero anchors are heterogeneous tuples: `("center", "center")` and
`("center", "bottom")`. Cavern only uses those two, so `actor.go` models them
with a minimal struct tagging the **y** axis as centre or bottom (x is always
centre):

```go
type Anchor struct{ yKind int }   // akCenter | akBottom
var AnchorCentre       = Anchor{akCenter}
var AnchorCentreBottom = Anchor{akBottom}
```

`Anchor.offset(w, h)` returns the pixel offset from the actor's `(X, Y)` to the
sprite's top‑left, so the drawing code and the point‑collision code
(`collidepoint`, `centerPoint`, `top`, `bottom`) agree on where each sprite
sits. `CollideActor` uses centre; `GravityActor` (player/robots/fruit) uses
centre‑bottom, so `y` is the character's feet — which is why robots test
`orb.y >= self.top and orb.y < self.bottom` against the anchored rectangle.

---

## 8. Sprite font and audio

**Font.** Both games render text from per‑glyph PNGs and a hardcoded width table
for A–Z (any other character uses the width of 'A'). Python:

```python
CHAR_WIDTH = [27, 26, 25, ...]
def char_width(char): return CHAR_WIDTH[max(0, ord(char) - 65)]
def draw_text(text, y, x=None):
    ...
    screen.blit("font0"+str(ord(char)), (x, y))
```

`text.go` mirrors this: `charWidthOf` indexes `charWidthTable` with
`ord(char) - 65` clamped to `[0, 25]`, and glyphs are blitted as
`"font0" + itoa(int(c))` (so `'A'` → `font065`, space → `font032`). The Python
`draw_text(text, y, x=None)` overload — centre when `x` is `None` — becomes two
functions, `drawTextX` and `drawTextCentre`. `draw_status` (score right‑aligned,
level centred, life/health icons) ports directly, including the same `IMAGE_WIDTH`
advances.

**Audio.** Python's `play_sound(name, count)` picks a random numbered variant
with `getattr(sounds, name + str(randint(0, count-1)))`. Go's
`Audio.PlaySound(name, count)` does the same explicitly (`name + index`, loaded
and cached). Both are best‑effort: a missing file or absent sound hardware is
non‑fatal. `Game.play_sound` keeps the "no sound when there is no player" guard
(`if g.player == nil { return }`), which is what silences the attract‑mode menu.

---

## 9. Numeric and semantic details reproduced exactly

Platformer feel is sensitive to integer/float behaviour, so several details were
transliterated with care:

- **`sign` returns 1 for zero.** Cavern's `sign` is `-1 if x < 0 else 1` — it
  never returns 0. This matters where it feeds `move(0, sign(vel_y), ...)` and
  robot direction choices, so the Go `sign` replicates it precisely (distinct
  from the three‑way `sign` used in some other ports).
- **Integer pixel stepping.** `CollideActor.move` starts from `int(self.x),
  int(self.y)` and steps one pixel at a time, so positions land on integer
  coordinates and an actor never tunnels into a block. The Go port floors to
  `int` at entry the same way and steps identically.
- **The block‑collision modulo test.** The condition
  `(dy>0 and new_y%B==0) or (dx>0 and new_x%B==0) or (dx<0 and new_x%B==B-1)`
  — checking only the direction of travel, and only when aligned to the grid — is
  copied verbatim, using Python‑style modulo (`pmod`) for safety.
- **The appended floor row.** Each level array has 17 rows; `next_level` appends
  a copy of the first row as an 18th (`self.grid = self.grid + [self.grid[0]]`),
  giving the `NUM_ROWS == 18` the draw loop and `block` expect. The Go port
  builds a fresh 18‑row slice (`append(base..., base[0])`) so it never mutates
  the shared `LEVELS` data — the exact concern the Python comment calls out.
- **Truncating division in progression.** `num_strong_enemies = 1 +
  int(self.level / 1.5)` truncates toward zero; Go uses
  `1 + int(float64(g.level)/1.5)`. `max_enemies` and `fire_probability` use the
  same integer/`min` arithmetic.
- **Timers start at −1.** `Game.timer`, `Orb.timer`, `Pop.timer` all start at
  −1 and are incremented at the top of `update`, so the first live frame is 0.
  Spawn checks like `timer % 100 == 0` therefore fire on the first frame — the
  Go port preserves the −1 starts.
- **Orb "None" trapped type.** Python uses `trapped_enemy_type = None` vs. `0/1`.
  Go can't put `None` in an `int`, so the port uses **`-1` as the sentinel**
  (`trappedEnemyType int`), matching every `!= None` / `== None` test.

---

## 10. Intentional differences (out of scope for the port)

- **Game controllers.** The Python original is keyboard‑only already (there is no
  joystick class here), so the Go port matches: arrow keys move, **SPACE** blows
  orbs / holds to blow further, **UP** jumps.
- **Version checks** for Python / Pygame Zero at startup have no Go analogue.
- **A `-selftest` flag** is *added* in Go: it runs a free‑running phase (robots,
  bolts, fruit, gravity, collisions), an orb phase (blowing orbs that trap
  robots, then float and pop), and a level‑cycle phase (loading and stepping
  every layout), printing per‑level object counts. It exists only to verify the
  port without a display and has no Python counterpart.

---

## 11. Summary

The port is a close, behaviour‑preserving translation, and — because Cavern is a
small, self‑contained game — one of the more direct ones. The substantive
rewrites are all forced by the language or framework:

- inheritance → **struct embedding with explicit forwarding** (no `self`
  interface needed here);
- list comprehensions → **typed slice filters**;
- the `game` global → an **explicit `*Game` parameter** everywhere;
- Pygame Zero's implicit asset/sound/font/loop machinery → **explicit go‑sdl3
  plumbing** (`assets.go`, `audio.go`, `input.go`, `text.go`, `main.go`);
- the space‑bar latch, anchor tuples, and the `None` orb‑trap sentinel → small
  faithful Go equivalents.

Everything that affects how the game *plays* — the movement/jump/orb state
machine, the pixel‑stepped collision and its modulo grid test, the `sign`‑of‑zero
convention, the enemy spawn/fire progression, the appended floor row, and the
once‑per‑frame space latch — is reproduced as‑is. The verification path is
`go build` + `-selftest` (all level layouts load and step, orbs trap and pop,
fruit is collected, the player dies and respawns — all without panics); on‑screen
visuals and audio require a real display to confirm.
