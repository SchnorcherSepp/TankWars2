package gui

/*
  This file offers an editor with which you can create your own maps.
*/

import (
	"TankWars2/core"
	"TankWars2/gui/resources"
	"TankWars2/maps"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image"
	"image/color"
	"os"
	"slices"
	"time"
)

// interface check: ebiten.Game
var _ ebiten.Game = (*Editor)(nil)

// Editor implements the ebiten.Game interface and manages the Map-Editor.
type Editor struct {
	world *core.World
	file  string

	xWidth       int
	yHeight      int
	screenWidth  int
	screenHeight int

	activeTile *core.Tile
	activeType byte
}

// RunEditor initializes and runs the map editor with the specified file and world data.
// It loads the world data from the provided file, checks its validity, and configures the editor's settings.
//
// This function is blocking!
func RunEditor(file string, world *core.World) {

	// load world from file
	loadWorld, err := maps.Loader(file)
	if err != nil {
		println(err.Error())
	}

	// check world
	if loadWorld != nil && loadWorld.XWidth != 0 && loadWorld.YHeight != 0 && len(loadWorld.Tiles) != 0 {
		// use file world
		println("load world from file")
		world = loadWorld
	} else {
		// use arg world
		if world == nil {
			// arg world is invalid
			println("world is nil")
			os.Exit(1)
		} else {
			// ok
			println("create new world")
		}
	}

	// config editor
	editor := &Editor{
		world:        world,
		file:         file,
		xWidth:       world.XWidth,  // world dimension X
		yHeight:      world.YHeight, // world dimension Y
		screenWidth:  world.XWidth*(tileX+1) + 20 + tileX/2,
		screenHeight: world.YHeight*(tileY*0.8) + 20,
	}

	// init basic tiles
	for xCol := 0; xCol < editor.xWidth; xCol++ {
		for yRow := 0; yRow < editor.yHeight; yRow++ {
			tile := editor.world.Tiles[xCol][yRow]
			if tile == nil || tile.Type == 0 {
				println("init tile", xCol, yRow)
				editor.world.Tiles[xCol][yRow] = core.NewTile(core.GRASS, xCol, yRow)
			}
		}
	}

	// config window
	ebiten.SetWindowTitle("EDITOR: " + file)
	ebiten.SetWindowIcon([]image.Image{resources.Imgs.Logo})
	ebiten.SetWindowSize(editor.screenWidth, editor.screenHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(10) // default: 60 ticks per second

	// run (BLOCKING)
	if err = ebiten.RunGame(editor); err != nil {
		println(err.Error())
	}
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
func (e *Editor) Layout(_, _ int) (int, int) {
	return e.screenWidth, e.screenHeight
}

// Update updates an img by one tick. The given argument represents a screen image.
//
// Update updates only the img logic and Draw draws the screen.
//
// In the first frame, it is ensured that Update is called at least once before Draw. You can use Update
// to initialize the img state.
//
// After the first frame, Update might not be called or might be called once
// or more for one frame. The frequency is determined by the current TPS (tick-per-second).
func (e *Editor) Update() error {
	return nil
}

// Draw draws the img screen by one frame.
//
// The give argument represents a screen image. The updated content is adopted as the img screen.
func (e *Editor) Draw(screen *ebiten.Image) {
	e.processUserInput()
	e.drawTiles(screen)
	e.drawUnits(screen)
	e.writeGlobalText(screen)
}

//---------------- INPUT ---------------------------------------------------------------------------------------------//

// processUserInput handles the user input interactions in the map editor. It performs various actions based on the
// user's mouse clicks and keyboard inputs.
func (e *Editor) processUserInput() {

	// Identify the tile under the cursor
	xCol, yRow := calcTile(ebiten.CursorPosition())

	// Set the active tile when the LEFT mouse button is pressed
	if e.world != nil && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		// set active tile
		e.activeTile = e.world.Tile(xCol, yRow)
		// Change the tile or set the unit based on the active type
		if e.activeTile != nil {
			if e.activeType != 0 {
				if slices.Contains(core.UNITS, e.activeType) {
					// SET UNIT
					e.activeTile.Unit = core.NewUnit(0, e.activeType)
				} else {
					// SET TILE
					e.activeTile.Type = e.activeType
				}
			} else {
				e.activeTile.Unit = nil // remove unit
			}
		}
	}

	// Set the player's ownership for the active tile and unit
	e.changeOwner()

	// Create tiles with specific types for the editor
	e.setActiveType()

	// Save the map to a file when the Control + S keys are pressed
	if ebiten.IsKeyPressed(ebiten.KeyControl) && inpututil.IsKeyJustPressed(ebiten.KeyS) {
		println("SAVE:", e.file)
		// Write the map data to the file
		if err := os.WriteFile(e.file, []byte(e.world.Json()), 0600); err != nil {
			println(err.Error())
		}
		// Avoid potential double writing by waiting for a short duration
		time.Sleep(1 * time.Second)
	}
}

// changeOwner updates the ownership of the active tile and its associated unit based on the currently pressed keyboard key.
func (e *Editor) changeOwner() {

	// Get the active tile
	tile := core.NewTile(core.DIRT, 0, 0) // create fake tile
	if e.activeTile != nil {
		tile = e.activeTile // set active tile if BASE
	}

	// Get the active unit
	unit := core.NewUnit(0, core.TANK) // create fake unit
	if tile != nil && tile.Unit != nil {
		unit = tile.Unit // set unit if exist
	}

	// Set the owner/player based on the pressed keys
	if ebiten.IsKeyPressed(ebiten.Key1) {
		tile.Owner = 1
		unit.Player = 1
	} else if ebiten.IsKeyPressed(ebiten.Key2) {
		tile.Owner = 2
		unit.Player = 2
	} else if ebiten.IsKeyPressed(ebiten.Key3) {
		tile.Owner = 3
		unit.Player = 3
	} else if ebiten.IsKeyPressed(ebiten.Key4) {
		tile.Owner = 4
		unit.Player = 4
	} else if ebiten.IsKeyPressed(ebiten.Key5) {
		tile.Owner = 5
		unit.Player = 5
	} else if ebiten.IsKeyPressed(ebiten.Key6) {
		tile.Owner = 6
		unit.Player = 6
	} else if ebiten.IsKeyPressed(ebiten.KeyDigit0) {
		tile.Owner = 0
		unit.Player = 0
	}

	// Clear the owner at the wrong tile type (non-BASE)
	if tile.Type != core.BASE {
		tile.Owner = 0
	}
}

// setActiveType sets the active type based on the currently pressed keyboard key.
func (e *Editor) setActiveType() {
	if ebiten.IsKeyPressed(ebiten.KeyX) {
		e.activeType = 0
	} else if ebiten.IsKeyPressed(ebiten.KeyB) {
		e.activeType = core.BASE
	} else if ebiten.IsKeyPressed(ebiten.KeyD) {
		e.activeType = core.DIRT
	} else if ebiten.IsKeyPressed(ebiten.KeyF) {
		e.activeType = core.FOREST
	} else if ebiten.IsKeyPressed(ebiten.KeyG) {
		e.activeType = core.GRASS
	} else if ebiten.IsKeyPressed(ebiten.KeyH) {
		e.activeType = core.HILL
	} else if ebiten.IsKeyPressed(ebiten.KeyO) {
		e.activeType = core.HOLE
	} else if ebiten.IsKeyPressed(ebiten.KeyM) {
		e.activeType = core.MOUNTAIN
	} else if ebiten.IsKeyPressed(ebiten.KeyS) {
		e.activeType = core.STRUCTURE
	} else if ebiten.IsKeyPressed(ebiten.KeyW) {
		e.activeType = core.WATER
	} else if ebiten.IsKeyPressed(ebiten.KeyA) {
		e.activeType = core.ARTILLERY
	} else if ebiten.IsKeyPressed(ebiten.KeyT) {
		e.activeType = core.TANK
	} else if ebiten.IsKeyPressed(ebiten.KeyU) {
		e.activeType = core.SOLDIER
	}
}

//---------------- DRAW ----------------------------------------------------------------------------------------------//

// drawTiles draws the tiles on the screen, including their images, ownership indicators, and supply information.
func (e *Editor) drawTiles(screen *ebiten.Image) {

	for xCol := 0; xCol < e.xWidth; xCol++ {
		for yRow := 0; yRow < e.yHeight; yRow++ {

			posX, posY := calcScreenPosition(xCol, yRow, false)

			// image config
			op := new(ebiten.DrawImageOptions)
			op.GeoM.Translate(posX, posY)
			op.Filter = ebiten.FilterLinear

			// load image
			tile := e.world.Tile(xCol, yRow)
			img := resources.Imgs.ErrorTile
			switch tile.Type {
			case core.BASE:
				list := resources.Imgs.TilesBase
				id := int(tile.ImageID) % len(list)
				img = list[id] // load random image
			case core.DIRT:
				list := resources.Imgs.TilesDirt
				id := int(tile.ImageID) % len(list)
				img = list[id] // load random image
			case core.FOREST:
				list := resources.Imgs.TilesForest
				id := int(tile.ImageID) % len(list)
				img = list[id] // load random image
			case core.GRASS:
				list := resources.Imgs.TilesGrass
				id := int(tile.ImageID) % len(list)
				img = list[id] // load random image
			case core.HILL:
				list := resources.Imgs.TilesHill
				id := int(tile.ImageID) % len(list)
				img = list[id] // load random image
			case core.HOLE:
				list := resources.Imgs.TilesHole
				id := int(tile.ImageID) % len(list)
				img = list[id] // load random image
			case core.MOUNTAIN:
				list := resources.Imgs.TilesMountain
				id := int(tile.ImageID) % len(list)
				img = list[id] // load random image
			case core.STRUCTURE:
				list := resources.Imgs.TilesStructure
				id := int(tile.ImageID) % len(list)
				img = list[id] // load random image
			case core.WATER:
				list := resources.Imgs.TilesWater
				id := int(tile.ImageID) % len(list)
				img = list[id] // load random image
			}

			// draw tile
			screen.DrawImage(img, op)

			// draw active tile
			if e.activeTile != nil && e.activeTile.XCol == xCol && e.activeTile.YRow == yRow {
				bgColor := color.RGBA{R: 66, G: 66, B: 0, A: 0}
				vector.DrawFilledCircle(screen, float32(posX+tileX/2), float32(posY+tileY/2), tileX/2, bgColor, false)
			}

			// draw tile owner
			if tile.Owner > 0 {
				clr := color.RGBA{R: 77, G: 77, B: 77, A: 222}
				x := posX + tileX/2
				y := posY + 14
				vector.DrawFilledCircle(screen, float32(x), float32(y), float32(10), clr, false)
				txt := fmt.Sprintf("%d", tile.Owner)
				ebitenutil.DebugPrintAt(screen, txt, int(x-3), int(y-8))
			}
		}
	}
}

// drawUnits draws all units, including tanks, soldiers, etc., on the screen with their respective images and colors.
func (e *Editor) drawUnits(screen *ebiten.Image) {
	const unitX = 50
	const unitY = 50

	// all units
	for _, t := range e.world.Units(0) {
		posX, posY := calcScreenPosition(t.XCol, t.YRow, true)

		// load image
		u := t.Unit
		if u == nil {
			continue // skip
		}
		img := resources.Imgs.Error

		switch u.Type {
		case core.ARTILLERY:
			img = resources.Imgs.Artillery
		case core.SOLDIER:
			img = resources.Imgs.Soldier
		case core.TANK:
			img = resources.Imgs.Tank
		}

		// select color
		newColor := color.RGBA{A: 255} // black (default)
		switch u.Player {
		case core.RED:
			newColor = color.RGBA{R: 0xff, A: 255}
		case core.BLUE:
			newColor = color.RGBA{B: 0xff, A: 255}
		case core.GREEN:
			newColor = color.RGBA{G: 0xff, A: 255}
		case core.YELLOW:
			newColor = color.RGBA{R: 0xff, G: 0xff, A: 255}
		case core.WHITE:
			newColor = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 255}
		case core.BLACK:
			newColor = color.RGBA{A: 255}
		}

		// change color
		changeColorsExceptTransparent(img, newColor)

		// image config
		op := new(ebiten.DrawImageOptions)
		op.GeoM.Translate(posX-unitX/2, posY-unitY/2)
		op.Filter = ebiten.FilterLinear

		// draw unit
		screen.DrawImage(img, op)
	}
}

// writeGlobalText write the global text top left.
func (e *Editor) writeGlobalText(screen *ebiten.Image) {
	s := "\n"
	s += "  'Left click' set tile or unit\n"
	s += "  'Strg + S' save to file\n"
	s += "  '0-6' change owner\n"
	s += "  'X' remove unit\n"
	s += "\n"
	s += "  'B' BASE\n"
	s += "  'D' DIRT\n"
	s += "  'F' FOREST\n"
	s += "  'G' GRASS\n"
	s += "  'H' HILL\n"
	s += "  'O' HOLE\n"
	s += "  'M' MOUNTAIN\n"
	s += "  'S' STRUCTURE\n"
	s += "  'W' WATER\n"
	s += "\n"
	s += "  'A' ARTILLERY\n"
	s += "  'U' SOLDIER\n"
	s += "  'T' TANK\n"
	s += "\n"
	ebitenutil.DebugPrint(screen, s)
}
