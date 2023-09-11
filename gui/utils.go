package gui

/*
  This file provides utility functions for graphical user interface interactions in the game.
*/

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	"math"
)

// calcScreenPosition calculates the screen position (in pixels) for a given grid position (xCol, yRow).
// If 'center' is true, the position is adjusted to the center of the tile.
func calcScreenPosition(xCol, yRow int, center bool) (posX, posY float64) {

	// position
	posX = float64(xCol * (tileX + 1))
	posY = float64(yRow * (tileY + 1))

	// windows border
	posX += 10
	posY += 10

	// center
	if center {
		posX += tileX / 2
		posY += tileY / 2
	}

	// correct grid (hex)
	if yRow%2 == 1 {
		posX += tileX / 2
	}
	posY -= float64(yRow) * tileY / 4

	return
}

// calcTile calculates the grid position (xCol, yRow) for a given screen position (posX, posY) in pixels.
func calcTile(posX, posY int) (xCol, yRow int) {

	// center
	posX -= tileX / 2
	posY -= tileY / 2

	// windows border
	posX -= 10
	posY -= 10

	// position
	x := float64(posX) / (tileX + 1)
	y := float64(posY) / (tileY + 1)

	// correct hex grid (Y)
	y *= 1.3
	y = math.Round(y) // round

	// correct hex grid (X)
	if int(y)%2 == 1 {
		x -= 0.5
	}
	x = math.Round(x) // round

	return int(x), int(y)
}

// changeColorsExceptTransparent modifies the colors of a given image, replacing non-transparent pixels with 'newColor'.
func changeColorsExceptTransparent(image *ebiten.Image, newColor color.Color) {
	s := image.Bounds().Size()
	width, height := s.X, s.Y

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			currentColor := image.At(x, y)
			_, _, _, a := currentColor.RGBA()

			if a > 0 {
				image.Set(x, y, newColor)
			}
		}
	}
}

// valueToColor maps a numeric value to a color based on the provided 'max' value.
// If 'invert' is true, the color mapping is inverted, transitioning from green to red.
func valueToColor(value, max int, invert bool) color.RGBA {
	if value < 0 {
		value = 0
	}
	if value > max {
		value = max
	}
	if invert {
		value = max - value
	}

	normalizedValue := float64(value) / float64(max)
	r := uint8(255 * normalizedValue)
	g := uint8(255 * (1 - normalizedValue))
	return color.RGBA{R: r, G: g, A: 255}
}
