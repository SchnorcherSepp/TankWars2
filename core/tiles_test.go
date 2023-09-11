package core

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNewTile(t *testing.T) {
	tile := NewTile(GRASS, 3, 4)
	assert.NotNil(t, tile)
	assert.Equal(t, uint8(GRASS), tile.Type)
	assert.Equal(t, 3, tile.XCol)
	assert.Equal(t, 4, tile.YRow)
	assert.Nil(t, tile.Unit)
	assert.Equal(t, uint8(0), tile.Owner)
}

func TestNewTileRandomImageID(t *testing.T) {
	tile1 := NewTile(GRASS, 0, 0)
	tile2 := NewTile(GRASS, 0, 0)
	assert.NotEqual(t, tile1.ImageID, tile2.ImageID)
}

func TestNewTileSupplyAndVisibilityMaps(t *testing.T) {
	tile := NewTile(GRASS, 0, 0)
	assert.NotNil(t, tile.Supply)
	assert.NotNil(t, tile.Visibility)
}

func TestNewTileSupplyAndVisibilityMapsEmpty(t *testing.T) {
	tile := NewTile(GRASS, 0, 0)
	assert.Empty(t, tile.Supply)
	assert.Empty(t, tile.Visibility)
}

func TestTileClone(t *testing.T) {
	// Create a sample Tile
	original := &Tile{
		Type:    1,
		ImageID: 42,
		XCol:    3,
		YRow:    4,
		Unit: &Unit{
			Player: 1,
			Type:   2,
			ID:     123,
		},
		Owner: 2,
		Visibility: map[uint8]int{
			1: 10,
			2: 15,
		},
		Supply: map[uint8]int{
			3: 8,
			4: 12,
		},
	}

	cloned := original.Clone()

	if cloned == original {
		t.Errorf("no cloning")
	}
	if !reflect.DeepEqual(original, cloned) {
		t.Errorf("Cloned unit does not match the original.\nOriginal: %+v\nCloned: %+v", original, cloned)
	}
}
