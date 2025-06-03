package core

/*
  This file defines the structure and methods related to the game world's hexagonal tiles.
*/

import (
	"encoding/json"
	"math"
)

//--------  Struct  --------------------------------------------------------------------------------------------------//

// Tile represents a single hexagonal tile within the game world grid.
type Tile struct {
	Type    byte  // tile type (see TILES)
	ImageID uint8 // ID used by the GUI to display random images
	XCol    int   // Column of the grid
	YRow    int   // Row of the grid

	// unit
	Unit *Unit // Combat unit located on this tile

	// set by 'update'
	Owner      uint8         // Player who last visited the tile
	Visibility map[uint8]int // Visibility level of each player on this tile (FogOfWar, NormalView, CloseView)
	Supply     map[uint8]int // Supply level of each player on this tile
}

// NewTile creates a new Tile object with the specified type and grid coordinates.
// It initializes the Tile's attributes and generates a random ImageID for GUI display.
func NewTile(t byte, xCol, yRow int) *Tile {
	return &Tile{
		Type:       t,
		ImageID:    uint8(rnd.Intn(math.MaxUint8)), // Random number in the range [0, 254]
		XCol:       xCol,
		YRow:       yRow,
		Visibility: make(map[uint8]int),
		Supply:     make(map[uint8]int),
	}
}

//--------  Getter  --------------------------------------------------------------------------------------------------//

// Clone creates a deep copy of a Tile struct using JSON serialization and deserialization.
// If any error occurs during the cloning process, nil is returned.
func (t *Tile) Clone() *Tile {

	// Serialize the original Tile struct into JSON data
	origJSON, err := json.Marshal(t)
	if err != nil {
		return nil
	}

	// Create an empty Tile struct to hold the cloned data
	clone := Tile{}

	// Deserialize the JSON data into the clone struct
	if err = json.Unmarshal(origJSON, &clone); err != nil {
		return nil
	}

	// Return a pointer to the cloned Tile struct
	return &clone
}
