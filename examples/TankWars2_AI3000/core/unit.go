package core

/*
  This file defines the structure and methods related to the game units.
*/

import "encoding/json"

//--------  Struct  --------------------------------------------------------------------------------------------------//

// Unit represents a combat unit in the game.
type Unit struct {

	// !!! ORIGIN FIX: START ----------------------
	MARK bool
	TEXT string // last unit command
	tile *Tile  // SET BY remote.client: updateTileRef()

	ObserverPhantom  uint64 // i know the position of this unit (1-100)%
	ObserverLastSeen uint64 // last seen (world iteration)
	ObserverVisible  bool   // i see this unit now
	// !!! ORIGIN FIX: END ------------------------

	Player uint8 // Player identifier.
	Type   byte  // Unit type (see UNITS).
	ID     int   // Unique unit identifier.
	Health int   // Current health points of the unit.

	// current commands
	Activity *Activity // Current activity the unit is engaged in.

	// Attributes set by 'update'
	View        int     // Visibility distance.
	CloseView   int     // Close visibility distance (see hidden).
	FireRange   int     // Firing range distance.
	Speed       uint64  // Movement speed (in game iterations).
	FireSpeed   uint64  // Firing speed (in game iterations).
	Hidden      bool    // Indicates if the unit is hidden.
	Armour      int     // Armor strength.
	Demoralized bool    // Indicates if the unit is demoralized (50% reduced damage).
	Ammunition  float32 // Remaining ammunition count.
}

// Activity represents a unit's ongoing activity or command.
type Activity struct {
	// !!! ORIGIN FIX: START ----------------------
	MARK bool
	// !!! ORIGIN FIX: END ------------------------

	Name  string // Name of the activity (MOVE, FIRE)
	From  [2]int // Starting coordinates of the activity.
	To    [2]int // Destination coordinates of the activity.
	Start uint64 // Start iteration of the activity.
	End   uint64 // End iteration of the activity.
}

// NewUnit creates a new unit with the specified player and type.
// It initializes the unit's attributes and returns a pointer to the newly created Unit struct.
func NewUnit(player uint8, typ byte, tile *Tile) *Unit {
	return &Unit{
		// !!! ORIGIN FIX: START ----------------------
		tile: tile,
		// !!! ORIGIN FIX: END ------------------------

		Player: player,
		Type:   typ,
		ID:     rnd.Int(), // random number as unique unit ID
		Health: 100,       // default health (100%)

		Ammunition: 99, // dummy ammunition (will be overwritten by update)
	}
}

//--------  Getter  --------------------------------------------------------------------------------------------------//

// Clone creates a deep copy of a Unit struct using JSON serialization and deserialization.
// If any error occurs during the cloning process, nil is returned.
func (u *Unit) Clone() *Unit {

	// Serialize the original Unit struct into JSON data
	origJSON, err := json.Marshal(u)
	if err != nil {
		return nil
	}

	// Create an empty Unit struct to hold the cloned data
	clone := Unit{}

	// Deserialize the JSON data into the clone struct
	if err = json.Unmarshal(origJSON, &clone); err != nil {
		return nil
	}

	// !!! ORIGIN FIX: START ----------------------
	if u != nil && u.tile != nil {
		//  clone tile
		clone.tile = u.tile.Clone()
	}
	// !!! ORIGIN FIX: END ------------------------

	// Return a pointer to the cloned Unit struct
	return &clone
}
