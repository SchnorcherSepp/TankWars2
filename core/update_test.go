package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdateBaseOwner(t *testing.T) {
	world := NewWorld(10, 10)
	baseTile := world.Tile(5, 5)
	baseTile.Type = BASE
	baseTile.Owner = 0

	unit := NewUnit(RED, SOLDIER)
	baseTile.Unit = unit

	// test
	assert.Equal(t, uint8(0), baseTile.Owner)   // no base owner
	updateBaseOwner(world)                      // call function
	assert.Equal(t, uint8(RED), baseTile.Owner) // base owner
	baseTile.Unit = nil
	updateBaseOwner(world)                      // call function #2
	assert.Equal(t, uint8(RED), baseTile.Owner) // base owner

	// error
	updateBaseOwner(nil)
}

func TestUpdateSupply(t *testing.T) {
	world := NewWorld(10, 10)
	baseTile := world.Tile(9, 0)
	baseTile.Type = BASE
	baseTile.Owner = RED

	// test
	testTile := world.Tile(0, 0)

	assert.NotNil(t, testTile.Supply)
	assert.Equal(t, 0, len(testTile.Supply))

	updateSupply(world)

	assert.Equal(t, 1, len(testTile.Supply))
	assert.Equal(t, 9, testTile.Supply[RED])
	assert.Equal(t, 0, testTile.Supply[BLUE])

	// error
	updateSupply(nil)
}

func TestUpdateVisibility(t *testing.T) {
	world := NewWorld(10, 10)
	tile := world.Tile(9, 0)
	unit := NewUnit(RED, TANK)
	tile.Unit = unit
	tile.Type = HILL

	// test
	testTile := world.Tile(8, 0)

	assert.NotNil(t, testTile.Visibility)
	assert.Equal(t, 0, len(testTile.Visibility))

	updateUnitAttributes(world) // set attributes!!
	updateVisibility(world)

	assert.Equal(t, 1, len(testTile.Visibility))
	assert.Equal(t, CloseView, testTile.Visibility[RED])
	assert.Equal(t, FogOfWar, testTile.Visibility[BLUE])
	assert.Equal(t, CloseView, world.Tile(7, 0).Visibility[RED])
	assert.Equal(t, NormalView, world.Tile(6, 0).Visibility[RED])

	// error
	updateSupply(nil)
}

func TestUpdateUnitAttributes(t *testing.T) {
	world := NewWorld(10, 10)
	baseTile := world.Tile(5, 5)
	baseTile.Owner = RED
	baseTile.Type = GRASS

	unit := NewUnit(RED, SOLDIER)
	baseTile.Unit = unit

	updateUnitAttributes(world)

	assert.Equal(t, 3, unit.View)
	assert.Equal(t, 1, unit.CloseView)
	assert.Equal(t, 0, unit.Armour)
	assert.Equal(t, 1, unit.FireRange)
	assert.Equal(t, uint64(90), unit.Speed)
	assert.Equal(t, uint64(69), unit.FireSpeed)
	assert.True(t, unit.Hidden)
	assert.Equal(t, float32(9), unit.Ammunition)

	// error
	updateSupply(nil)
}

func TestHealUnits(t *testing.T) {
	world := NewWorld(10, 10)
	baseTile := world.Tile(5, 5)
	baseTile.Type = BASE
	baseTile.Owner = RED

	unit := NewUnit(RED, SOLDIER)
	unit.Health = 50
	world.Tile(5, 5).Unit = unit

	world.Iteration = 100
	healUnits(world)

	assert.Equal(t, 51, unit.Health)
	assert.False(t, unit.Demoralized)

	world.Iteration = 200
	unit.Demoralized = true
	healUnits(world)

	assert.False(t, unit.Demoralized)

	// error
	updateSupply(nil)
}
