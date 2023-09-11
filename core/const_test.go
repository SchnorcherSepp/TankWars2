package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConstants(t *testing.T) {
	assert.Equal(t, 30, GameSpeed)
	assert.Equal(t, 15, MaxSupply)
	assert.Equal(t, 1.0, SupplySpeed)

	assert.Equal(t, 'B', BASE)
	assert.Equal(t, 'D', DIRT)
	assert.Equal(t, 'F', FOREST)
	assert.Equal(t, 'G', GRASS)
	assert.Equal(t, 'H', HILL)
	assert.Equal(t, 'O', HOLE)
	assert.Equal(t, 'M', MOUNTAIN)
	assert.Equal(t, 'S', STRUCTURE)
	assert.Equal(t, 'W', WATER)

	assert.Equal(t, 'A', ARTILLERY)
	assert.Equal(t, 'U', SOLDIER)
	assert.Equal(t, 'T', TANK)

	assert.Equal(t, 1, RED)
	assert.Equal(t, 2, BLUE)
	assert.Equal(t, 3, GREEN)
	assert.Equal(t, 4, YELLOW)
	assert.Equal(t, 5, WHITE)
	assert.Equal(t, 6, BLACK)

	assert.Equal(t, 0, FogOfWar)
	assert.Equal(t, 1, NormalView)
	assert.Equal(t, 2, CloseView)

	assert.Equal(t, "MOVE", MOVE)
	assert.Equal(t, "FIRE", FIRE)

	assert.Equal(t, []byte{BASE, DIRT, FOREST, GRASS, HILL, HOLE, MOUNTAIN, STRUCTURE, WATER}, TILES)
	assert.Equal(t, []byte{ARTILLERY, SOLDIER, TANK}, UNITS)
	assert.Equal(t, []byte{RED, BLUE, GREEN, YELLOW, WHITE, BLACK}, PLAYERS)
}
