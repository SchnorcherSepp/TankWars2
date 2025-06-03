package gui

/*
  The file implements functions for visually representing tiles and elements in a game's graphical user interface (GUI).
  Additionally, shadows are added based on the active player's visibility, and in fire mode, tiles in the fire area are
  highlighted. Debug information such as coordinates, supply values, and visibility are also displayed on screen when needed.
*/

import (
	"TankWars2_AI/core"
	"TankWars2_AI/gui/resources"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
	"sort"
)

// drawTiles draws the tiles on the screen, including their images, ownership indicators, and supply information.
func (g *Game) drawTiles(screen *ebiten.Image) {

	for xCol := 0; xCol < g.xWidth; xCol++ {
		for yRow := 0; yRow < g.yHeight; yRow++ {

			posX, posY := calcScreenPosition(xCol, yRow, false)

			// image config
			op := new(ebiten.DrawImageOptions)
			op.GeoM.Translate(posX, posY)
			op.Filter = ebiten.FilterLinear

			// load image
			tile := g.world.Tile(xCol, yRow)
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

			// draw the base owner
			if tile.Type == core.BASE && tile.Owner > 0 {
				clr := color.RGBA{R: 77, G: 77, B: 77, A: 222}
				x := posX + tileX/2
				y := posY + 14
				vector.DrawFilledCircle(screen, float32(x), float32(y), float32(10), clr, false)
				txt := fmt.Sprintf("%d", tile.Owner)
				ebitenutil.DebugPrintAt(screen, txt, int(x-3), int(y-8))
			}

			// draw supply
			if g.toggleSupply && g.activePlayer > 0 && len(tile.Supply) > 0 {
				supply := tile.Supply[g.activePlayer]
				const size = 16

				if supply > 0 {
					clr := valueToColor(supply, core.MaxSupply, false)

					x := posX + tileX - size - 3
					y := posY + tileY/4 + 3
					vector.DrawFilledRect(screen, float32(x), float32(y), size, size, clr, false)

					txt := fmt.Sprintf("%d", supply)
					ebitenutil.DebugPrintAt(screen, txt, int(x)-(3*len(txt))+size/2, int(y-8+size/2))
				}
			}
		}
	}
}

// drawShadow draws shadows over the tiles based on the visibility of the active player.
func (g *Game) drawShadow(screen *ebiten.Image) {

	if g.toggleVisibility && g.activePlayer > 0 {
		img := resources.Imgs.EmptyTile

		fogOfWarImg := ebiten.NewImageFromImage(img)
		changeColorsExceptTransparent(fogOfWarImg, color.RGBA{R: 66, A: 100})

		closeViewImg := ebiten.NewImageFromImage(img)
		changeColorsExceptTransparent(closeViewImg, color.RGBA{R: 99, G: 99, B: 99, A: 99})

		for xCol := 0; xCol < g.xWidth; xCol++ {
			for yRow := 0; yRow < g.yHeight; yRow++ {

				posX, posY := calcScreenPosition(xCol, yRow, false)

				// image config
				op := new(ebiten.DrawImageOptions)
				op.GeoM.Translate(posX, posY)
				op.Filter = ebiten.FilterLinear

				// get tile
				tile := g.world.Tile(xCol, yRow)

				// draw shadow
				switch tile.Visibility[g.activePlayer] {
				case core.FogOfWar:
					screen.DrawImage(fogOfWarImg, op)
				case core.CloseView:
					screen.DrawImage(closeViewImg, op)
				}
			}
		}
	}
}

// drawFireMode highlights the tiles within the firing range of the active unit when in fire mode.
func (g *Game) drawFireMode(screen *ebiten.Image) {
	activeTile := g.activeTile
	if g.fireMode && activeTile != nil {

		unit := activeTile.Unit
		if unit != nil {

			// load image
			img := resources.Imgs.EmptyTile
			img = ebiten.NewImageFromImage(img)
			changeColorsExceptTransparent(img, color.RGBA{R: 188, G: 1, B: 1, A: 22})

			// all tiles in fire range
			neighbors := g.world.ExtNeighbors(activeTile, unit.FireRange)
			for _, tmp := range neighbors {
				for _, t := range tmp {
					posX, posY := calcScreenPosition(t.XCol, t.YRow, false)

					// image config
					op := new(ebiten.DrawImageOptions)
					op.GeoM.Translate(posX, posY)
					op.Filter = ebiten.FilterLinear

					// draw tile
					screen.DrawImage(img, op)
				}
			}
		}
	}
}

// drawActiveTile highlights the currently active tile on the screen.
func (g *Game) drawActiveTile(screen *ebiten.Image) {
	if g.activeTile == nil {
		return
	}

	posX, posY := calcScreenPosition(g.activeTile.XCol, g.activeTile.YRow, true)

	bgColor := color.RGBA{R: 66, G: 66, B: 0, A: 0}
	vector.DrawFilledCircle(screen, float32(posX), float32(posY), tileX/2, bgColor, false)
}

// writeTileText writes debug information about each tile on the screen, including coordinates, supply values, and visibility.
func (g *Game) writeTileText(screen *ebiten.Image) {

	for xCol := 0; xCol < g.xWidth; xCol++ {
		for yRow := 0; yRow < g.yHeight; yRow++ {
			tile := g.world.Tile(xCol, yRow)
			txt := ""

			// coordinates
			if g.toggleCoordinates {
				txt += fmt.Sprintf("[%d,%d]", tile.XCol, tile.YRow)
			}

			// supply array
			if g.activePlayer == 0 && g.toggleSupply {
				// add new line
				if len(txt) > 0 {
					txt += "\n"
				}
				// get all Keys
				keys := make([]uint8, 0, len(tile.Supply))
				for key := range tile.Supply {
					keys = append(keys, key)
				}
				// sort keys
				sort.Slice(keys, func(i, j int) bool {
					return keys[i] < keys[j]
				})
				// add text
				txt += fmt.Sprintf("S: ")
				for _, key := range keys {
					txt += fmt.Sprintf("%d=%d_", key, tile.Supply[key])
				}
			}

			// visibility array
			if g.activePlayer == 0 && g.toggleVisibility {
				// add new line
				if len(txt) > 0 {
					txt += "\n"
				}
				// get all Keys
				keys := make([]uint8, 0, len(tile.Visibility))
				for key := range tile.Visibility {
					keys = append(keys, key)
				}
				// sort keys
				sort.Slice(keys, func(i, j int) bool {
					return keys[i] < keys[j]
				})
				// add text
				txt += fmt.Sprintf("V: ")
				for _, key := range keys {
					value := tile.Visibility[key]
					txt += fmt.Sprintf("%d=%d_", key, value)
				}
			}

			// write text
			if len(txt) > 0 {
				posX, posY := calcScreenPosition(xCol, yRow, true)
				posX = posX - (6 / 2 * float64(len(txt)))
				posY = posY - 8
				ebitenutil.DebugPrintAt(screen, txt, int(posX), int(posY))
			}
		}
	}
}
