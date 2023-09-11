package core

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWorld(t *testing.T) {
	world := NewWorld(10, 10)
	assert.NotNil(t, world)
	assert.Equal(t, 10, len(world.Tiles))
	assert.Equal(t, 10, len(world.Tiles[0]))
}

func TestTile(t *testing.T) {
	world := NewWorld(10, 10)
	tile := world.Tile(5, 5)
	assert.NotNil(t, tile)

	invalidTile := world.Tile(11, 5)
	assert.Nil(t, invalidTile)
}

func TestTileList(t *testing.T) {
	world := NewWorld(10, 10)
	tile := world.Tile(5, 5)
	tile.Type = BASE

	baseTiles := world.TileList(BASE)
	assert.NotNil(t, baseTiles)
	assert.Equal(t, 1, len(baseTiles))

	baseTiles = world.TileList(0)
	assert.NotNil(t, baseTiles)
	assert.Equal(t, 10*10, len(baseTiles))
}

func TestUnits(t *testing.T) {
	world := NewWorld(10, 10)
	unit1 := NewUnit(RED, SOLDIER)
	world.Tile(5, 5).Unit = unit1
	unit2 := NewUnit(BLUE, SOLDIER)
	world.Tile(6, 5).Unit = unit2

	redUnits := world.Units(RED)
	assert.NotNil(t, redUnits)
	assert.Equal(t, 1, len(redUnits))

	allUnits := world.Units(0)
	assert.NotNil(t, allUnits)
	assert.Equal(t, 2, len(allUnits))
}

func TestNeighbors(t *testing.T) {
	world := NewWorld(10, 10)
	tile := world.Tile(5, 5)

	neighbors := world.Neighbors(tile)
	assert.NotNil(t, neighbors)
	assert.Equal(t, 6, len(neighbors))

	tile = world.Tile(0, 0)
	neighbors = world.Neighbors(tile)
	assert.NotNil(t, neighbors)
	assert.Equal(t, 2, len(neighbors))
}

func TestExtNeighbors(t *testing.T) {
	world := NewWorld(10, 10)
	tile := world.Tile(5, 5)

	extNeighbors := world.ExtNeighbors(tile, 3)
	assert.NotNil(t, extNeighbors)
	assert.Equal(t, 3, len(extNeighbors))
}

func TestMove(t *testing.T) {
	world := NewWorld(10, 10)
	from := world.Tile(5, 5)
	to := world.Tile(5, 6)
	unit := NewUnit(RED, SOLDIER)
	world.Tile(5, 5).Unit = unit

	_, err := world.Move(from, to, 0)
	assert.NoError(t, err)
	assert.NotNil(t, unit.Activity)

	_, err = world.Move(from, to, 0)
	assert.Error(t, err)
}

func TestFire(t *testing.T) {
	world := NewWorld(10, 10)
	from := world.Tile(5, 5)
	to := world.Tile(5, 6)
	unit := NewUnit(RED, TANK)
	unit.Ammunition = 1
	from.Unit = unit
	from.Type = GRASS

	updateUnitAttributes(world) // set attributes!!

	err := world.Fire(from, to, 0)
	assert.NoError(t, err)
	assert.NotNil(t, unit.Activity)

	err = world.Fire(from, to, 0)
	assert.Error(t, err)
}

func TestClone(t *testing.T) {
	// Create a new world
	original := NewWorld(21, 13)

	// Perform any necessary setup and modifications to the world
	for _, t := range original.TileList(0) {
		t.Type = GRASS
		if t.XCol <= 2 || t.XCol >= original.XWidth-2-1 {
			t.Type = FOREST
		}
		if t.XCol == original.XWidth/2 || t.XCol == original.XWidth/2-1 {
			t.Type = WATER
		}
	}
	original.Tile(0, 0).Type = BASE
	original.Tile(original.XWidth-1, 0).Type = BASE
	original.Tile(original.XWidth/2-6, original.YHeight/2-3).Unit = NewUnit(RED, ARTILLERY)
	original.Tile(original.XWidth-1, original.YHeight-1).Unit = NewUnit(BLUE, SOLDIER)
	original.Tile(0+3, 4).Unit = NewUnit(RED, TANK)
	original.Tile(original.XWidth-5, original.YHeight-4).Unit = NewUnit(BLUE, TANK)

	// clone
	cloned := original.Clone()

	if cloned == original {
		t.Errorf("no cloning")
	}
	if !reflect.DeepEqual(original, cloned) {
		t.Errorf("Cloned unit does not match the original.\nOriginal: %+v\nCloned: %+v", original, cloned)
	}
}

func TestCensorship(t *testing.T) {
	// Create a sample game world for testing.
	world := NewWorld(3, 3)
	tile := world.Tile(1, 1)
	unit := &Unit{Player: 1, Hidden: true}
	tile.Unit = unit
	tile.Supply[0] = 5
	tile.Supply[1] = 10
	tile.Visibility[0] = NormalView
	tile.Visibility[1] = FogOfWar
	tile.Owner = 1

	// Apply censorship to the game world.
	censoredWorld := Censorship(world, 1)
	if censoredWorld == nil {
		t.Fatal("Censorship failed: Cloning error")
	}

	// Check the result after censorship.
	censoredTile := censoredWorld.Tile(1, 1)
	if censoredTile == nil {
		t.Fatal("Censorship failed: Tile not found")
	}

	if censoredTile.Unit != nil {
		t.Fatal("Censorship failed: Hidden unit not removed")
	}

	if censoredTile.Supply[0] != 0 || censoredTile.Supply[1] != 10 {
		t.Fatal("Censorship failed: Supply data not removed properly")
	}

	if censoredTile.Visibility[0] != 0 || censoredTile.Visibility[1] != FogOfWar {
		t.Fatal("Censorship failed: Visibility data not removed properly")
	}

	if censoredTile.Owner != 0 && censoredTile.Owner != 1 {
		t.Fatal("Censorship failed: Owner not reset properly")
	}
}
