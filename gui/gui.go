package gui

/*
  This file provides functionality for creating and managing the graphical user interface of the game.
  It includes functions for setting up the game window, processing user input, and rendering the game scene.
*/

import (
	"TankWars2/core"
	"TankWars2/gui/resources"
	"TankWars2/remote"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"image"
	"time"
)

// tileX and tileY define the width and height of each tile in pixels.
const (
	tileX = 120.0 // width
	tileY = 140.0 // height
)

// interface check: ebiten.Game
var _ ebiten.Game = (*Game)(nil)

// Game implements the ebiten.Game interface and manages the GUI.
type Game struct {
	world  *core.World
	remote *remote.Client
	speed  int

	xWidth       int
	yHeight      int
	screenWidth  int
	screenHeight int

	lastCommand       time.Time
	activeTile        *core.Tile
	activePlayer      uint8
	fireMode          bool
	toggleHelp        bool
	toggleCoordinates bool
	toggleSupply      bool
	toggleVisibility  bool
}

// RunGame initializes the game window and starts the GUI loop.
// The Update function from core.World is called with 30 Ticks per second (see core.GameSpeed).
//
// This function is blocking!
func RunGame(title string, world *core.World, remote *remote.Client, mute bool) error {
	resources.MuteSound = mute

	// config gui
	game := &Game{
		world:        world,
		remote:       remote,
		xWidth:       world.XWidth,  // world dimension X
		yHeight:      world.YHeight, // world dimension Y
		screenWidth:  world.XWidth*(tileX+1) + 20 + tileX/2,
		screenHeight: world.YHeight*(tileY*0.8) + 20,
	}

	// config window
	ebiten.SetWindowTitle(title)
	ebiten.SetWindowIcon([]image.Image{resources.Imgs.Logo})
	ebiten.SetWindowSize(game.screenWidth, game.screenHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(core.GameSpeed) // default: 60 ticks per second

	// run (BLOCKING)
	return ebiten.RunGame(game)
}

//--------------------------------------------------------------------------------------------------------------------//

// Layout accepts a native outside size in device-independent pixels and returns the img logical screen
// size.
//
// On desktops, the outside is a window or a monitor (fullscreen mode). On browsers, the outside is a body
// element. On mobiles, the outside is the view's size.
//
// Even though the outside size and the screen size differ, the rendering scale is automatically adjusted to
// fit with the outside.
//
// Layout is called almost every frame.
//
// It is ensured that Layout is invoked before Update is called in the first frame.
//
// If Layout returns non-positive numbers, the caller can panic.
//
// You can return a fixed screen size if you don't care, or you can also return a calculated screen size
// adjusted with the given outside size.
func (g *Game) Layout(_, _ int) (int, int) {
	return g.screenWidth, g.screenHeight
}

//---------------- DRAW ----------------------------------------------------------------------------------------------//

// Draw draws the img screen by one frame.
//
// The give argument represents a screen image. The updated content is adopted as the img screen.
func (g *Game) Draw(screen *ebiten.Image) {

	// process user input
	g.processUserInput()

	// draw tiles
	g.drawTiles(screen)
	g.drawShadow(screen)
	g.drawFireMode(screen)
	g.drawActiveTile(screen)
	g.writeTileText(screen)

	// draw units
	g.drawUnits(screen)

	// write global text
	g.writeGlobalText(screen)
}

// writeGlobalText write the global text top left.
func (g *Game) writeGlobalText(screen *ebiten.Image) {
	s := "\n"

	s += fmt.Sprintf("   Iteration: %d\n", g.world.Iteration)
	if g.toggleHelp {
		s += "  - Left click: select\n"
		s += "  - Right click: move\n"
		s += "  - Ctrl + Right click: attack\n"
		s += "  - 'Ctrl' (hold): fire mode\n"
		s += "  - 'C': coordinates\n"
		s += "  - 'S': supply\n"
		s += "  - 'V': visibility\n"
		s += "  - 0-9: player view\n"
		s += "\n"
	} else {
		s += "   Press 'H' for help\n"
		s += "\n"
	}
	ebitenutil.DebugPrint(screen, s)
}

//---------------- USER INPUT ----------------------------------------------------------------------------------------//

func (g *Game) processUserInput() {
	// tile with cursor on it
	xCol, yRow := calcTile(ebiten.CursorPosition())
	// delay after inputs
	if time.Now().Sub(g.lastCommand) < 200*time.Millisecond {
		return
	}

	// ----  commands  -------------------------------

	// set active tile [LEFT mouse button]
	setActiveTile(g, xCol, yRow)

	// set move command [RIGHT mouse button]  (Fire mode: OFF)
	setMoveCommand(g, xCol, yRow)

	// set fire command [RIGHT mouse button]  (file mode: ON)
	setFireCommand(g, xCol, yRow)

	// select player: 1, 2, 3, 4, 5, 6 and 0
	selectPlayer(g)

	// ----  toggle  -------------------------------

	// toggle KEY: help ['H']
	if ebiten.IsKeyPressed(ebiten.KeyH) {
		g.toggleHelp = !g.toggleHelp
		g.lastCommand = time.Now() // force delay after input
	}

	// toggle KEY: coordinates ['C']
	if ebiten.IsKeyPressed(ebiten.KeyC) {
		g.toggleCoordinates = !g.toggleCoordinates
		g.lastCommand = time.Now() // force delay after input
	}

	// toggle KEY: supply ['S']
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.toggleSupply = !g.toggleSupply
		g.lastCommand = time.Now() // force delay after input
	}

	// toggle KEY: visibility ['V']
	if ebiten.IsKeyPressed(ebiten.KeyV) {
		g.toggleVisibility = !g.toggleVisibility
		g.lastCommand = time.Now() // force delay after input
	}

	// activate fire mode
	g.fireMode = ebiten.IsKeyPressed(ebiten.KeyControl)
}

//---------------- HELPER --------------------------------------------------------------------------------------------//

func setActiveTile(g *Game, xCol, yRow int) {
	// [LEFT mouse button]
	if g.world != nil && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.activeTile = g.world.Tile(xCol, yRow)
		g.lastCommand = time.Now() // force delay after input
	}
}

func setMoveCommand(g *Game, xCol, yRow int) {
	//  [RIGHT mouse button]  (Fire mode: OFF)
	if g.world != nil && !g.fireMode && ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		target := g.world.Tile(xCol, yRow)

		// send command
		var err error
		if g.remote == nil {
			// local command
			target, err = g.world.Move(g.activeTile, target, 0)
		} else {
			// remote command
			err = g.remote.Move(g.activeTile.XCol, g.activeTile.YRow, target.XCol, target.YRow)
		}

		// error sound
		if err != nil {
			resources.PlaySound(resources.Sounds.Error) // play error
			fmt.Printf("%v\n", err)
		} else {
			// OK: move active tile
			g.activeTile = target
		}

		// command delay
		g.lastCommand = time.Now() // force delay after input
	}
}

func setFireCommand(g *Game, xCol, yRow int) {
	// [RIGHT mouse button]  (file mode: ON)
	if g.world != nil && g.fireMode && ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		target := g.world.Tile(xCol, yRow)

		// send command
		var err error
		if g.remote == nil {
			// local command
			err = g.world.Fire(g.activeTile, target, 0)
		} else {
			// remote command
			err = g.remote.Fire(g.activeTile.XCol, g.activeTile.YRow, target.XCol, target.YRow)
		}

		// error sound
		if err != nil {
			resources.PlaySound(resources.Sounds.Error) // play error
			fmt.Printf("%v\n", err)
		}

		// command delay
		g.lastCommand = time.Now() // force delay after input
	}
}

func selectPlayer(g *Game) {
	if ebiten.IsKeyPressed(ebiten.Key1) {
		g.activePlayer = 1
	}
	if ebiten.IsKeyPressed(ebiten.Key2) {
		g.activePlayer = 2
	}
	if ebiten.IsKeyPressed(ebiten.Key3) {
		g.activePlayer = 3
	}
	if ebiten.IsKeyPressed(ebiten.Key4) {
		g.activePlayer = 4
	}
	if ebiten.IsKeyPressed(ebiten.Key5) {
		g.activePlayer = 5
	}
	if ebiten.IsKeyPressed(ebiten.Key6) {
		g.activePlayer = 6
	}
	if ebiten.IsKeyPressed(ebiten.KeyDigit0) {
		g.activePlayer = 0
	}
}
