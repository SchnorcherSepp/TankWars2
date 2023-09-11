package gui

/*
  This file contains functions responsible for drawing units, visualizing unit activities, and displaying unit
  statistics on the screen.
*/

import (
	"TankWars2/core"
	"TankWars2/gui/resources"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
	"strconv"
)

// drawUnits draws all units, including tanks, soldiers, etc., on the screen with their respective images and colors.
func (g *Game) drawUnits(screen *ebiten.Image) {
	const unitX = 50
	const unitY = 50

	// all units
	for _, t := range g.world.Units(0) {
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

		// change unit color (blink) with activity
		if u.Activity != nil && g.world.Iteration%8 <= 4 {
			newColor = color.RGBA{R: 0x66, G: 0x66, B: 0x66, A: 255}
		}

		// change color
		changeColorsExceptTransparent(img, newColor)

		// image config
		op := new(ebiten.DrawImageOptions)
		op.GeoM.Translate(posX-unitX/2, posY-unitY/2)
		op.Filter = ebiten.FilterLinear

		// draw unit
		screen.DrawImage(img, op)

		// draw activity
		drawActivity(screen, newColor, g.world, u.Activity)

		// draw unit stats
		drawUnitStats(screen, t)

		// write tile TEXT
		txt := fmt.Sprintf("%d", u.Health)
		posX = posX - (6 / 2 * float64(len(txt)))
		posY = posY - 8

		ebitenutil.DebugPrintAt(screen, txt, int(posX), int(posY))
	}
}

// drawActivity visualizes unit activities by drawing lines and progress indicators between source and target tiles.
// Different activities (MOVE, FIRE) are represented with varying line thickness and shapes (rectangles/circles).
// Explosions are displayed based on timing, and corresponding sounds are played.
func drawActivity(screen *ebiten.Image, clr color.RGBA, world *core.World, activity *core.Activity) {
	if activity == nil {
		return
	}

	// source and target tile
	from := world.Tile(activity.From[0], activity.From[1])
	to := world.Tile(activity.To[0], activity.To[1])

	// tile coordinates
	x1, y1 := calcScreenPosition(from.XCol, from.YRow, true)
	x2, y2 := calcScreenPosition(to.XCol, to.YRow, true)

	// line thickness
	thickness := float32(10)
	if activity.Name == core.FIRE {
		thickness = 1
	}

	// draw line
	vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), thickness, clr, false)

	// draw progress
	totalRange := float64(activity.End - activity.Start)
	progress := float64(world.Iteration - activity.Start)
	percent := progress / totalRange
	xx := x1 - ((x1 - x2) * percent)
	yy := y1 - ((y1 - y2) * percent)
	if activity.Name == core.MOVE {
		vector.DrawFilledRect(screen, float32(xx-10), float32(yy-10), float32(20), float32(20), clr, false)
	} else {
		vector.DrawFilledCircle(screen, float32(xx), float32(yy), float32(6), clr, false)
	}

	// draw explosion
	if activity.End-core.GameSpeed/3 < world.Iteration && activity.Name == core.FIRE {
		op := new(ebiten.DrawImageOptions)
		op.GeoM.Translate(x2-32, y2-32) // basic image 64x64
		op.Filter = ebiten.FilterLinear // Specify linear filter.
		screen.DrawImage(resources.Imgs.Explosion, op)
	}

	// play sounds
	if activity.Start == world.Iteration && activity.Name == core.FIRE {
		resources.PlaySound(resources.Sounds.Fire)
	}
	if activity.End == world.Iteration && activity.Name == core.FIRE {
		resources.PlaySound(resources.Sounds.Explosion)
	}
}

// drawUnitStats displays unit statistics and status symbols near each unit on the screen.
// Symbols for demoralized and hidden units are positioned based on unit status.
// Armor and ammunition indicators are drawn using shapes and colors, along with numerical values.
func drawUnitStats(screen *ebiten.Image, tile *core.Tile) {
	unit := tile.Unit
	if unit == nil {
		return // skip
	}

	posX, posY := calcScreenPosition(tile.XCol, tile.YRow, true)

	// demoralized (!) symbol top left
	if unit.Demoralized {
		img := resources.Imgs.Danger
		op := new(ebiten.DrawImageOptions)
		op.GeoM.Translate(posX-31, posY-27)
		op.Filter = ebiten.FilterLinear
		screen.DrawImage(img, op)
	}

	// hidden (eye) symbol top right
	if unit.Hidden {
		img := resources.Imgs.Eye
		op := new(ebiten.DrawImageOptions)
		op.GeoM.Translate(posX+16, posY-29)
		op.Filter = ebiten.FilterLinear
		screen.DrawImage(img, op)
	}

	// armour
	{
		// position
		x := float32(posX) + 9
		y := float32(posY) + 20
		size := float32(17)

		// select color
		clr := valueToColor(unit.Armour, 4, true)

		// shadow
		vector.DrawFilledCircle(screen, x+size/2+2, y+size/2+2, size/2, color.Black, false)
		vector.DrawFilledRect(screen, x+2, y+2, size, size/2, color.Black, false)
		// symbol
		vector.DrawFilledCircle(screen, x+size/2, y+size/2, size/2, clr, false)
		vector.DrawFilledRect(screen, x, y, size, size/2, clr, false)
		// text
		ebitenutil.DebugPrintAt(screen, strconv.Itoa(unit.Armour), int(x)+5, int(y))
	}

	// ammunition
	{
		// position
		x := float32(posX) - 9 - 16
		y := float32(posY) + 20
		size := float32(14)

		// select color
		clr := valueToColor(int(unit.Ammunition), 5, true)

		// shadow
		vector.DrawFilledCircle(screen, x+size/2+2, y+size/2+2, size/2, color.Black, false)
		vector.DrawFilledRect(screen, x+2, y+size/2+2, size, size/2+3, color.Black, false)
		// symbol
		vector.DrawFilledCircle(screen, x+size/2, y+size/2, size/2, clr, false)
		vector.DrawFilledRect(screen, x, y+size/2, size, size/2+3, clr, false)
		// text
		ebitenutil.DebugPrintAt(screen, strconv.Itoa(int(unit.Ammunition)), int(x)+4, int(y)+1)
	}
}
