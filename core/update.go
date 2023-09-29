package core

/*
  The central focus of this file is the Update() function. This function serves as
  the heart of the gameplay mechanics, orchestrating a series of interconnected
  operations. These operations include base ownership determination, supply network
  management, unit movement and combat processing, attribute adjustments, healing
  procedures, and visibility range recalculations. All other functions within the
  file are integral components called exclusively by Update(), effectively making
  them integral to the comprehensive Update operation. Sub functions on the subject
  of damage and attack have been moved to the separate file `attack.go`.
*/

import (
	"github.com/SchnorcherSepp/TankWars2/gui/resources"
)

//--------  Setter  --------------------------------------------------------------------------------------------------//

// Update processes a single iteration of the game world, applying various updates
// and calculations.
//
// It performs the following steps in sequence:
// - Sets the owner of bases on the map based on unit presence and proximity.
// - Updates the supply levels on the map, considering changes in base ownership.
// - Processes movement and firing commands of units.
// - Updates unit statistics and attributes, including ammunition and health.
// - Heals units stationed at bases over time and fixes demoralization status.
// - Updates visibility ranges for units on the map.
// - Advances the iteration count to mark the completion of the current iteration.
func (w *World) Update() {
	w.lock.Lock()         // Acquire the lock to ensure thread safety
	defer w.lock.Unlock() // Release the lock when the function exits

	// enforce freeze
	if w.Freeze {
		return // so nothing
	}

	// Set the owner of bases
	updateBaseOwner(w)

	// Update supply levels on the map
	updateSupply(w)

	// process commands
	processMove(w)
	processFire(w)

	// Update unit attributes and ammunition
	updateUnitAttributes(w)
	healUnits(w)

	// Update visibility ranges for units
	updateVisibility(w)

	// Advance the iteration count
	w.Iteration++
}

//--------  Helper  --------------------------------------------------------------------------------------------------//

// updateBaseOwner identifies all supply depots (BASE) on the world map and assigns
// ownership of the base to whoever has units positioned on the field.
func updateBaseOwner(world *World) {
	if world == nil {
		return
	}

	// Iterate through all units on the world map
	for _, tile := range world.Units(0) {
		unit := tile.Unit
		if unit == nil {
			continue // Skip non-occupied tiles
		}
		player := unit.Player

		// Assign tile owner if it is a base
		if tile.Type == BASE {
			tile.Owner = player // set new owner
		}
	}
}

// updateSupply calculates and updates the supply network for military bases in a game world.
// The supply network is created by determining which fields can be supplied from each
// own military base within a specified maximum distance. The lower the supply value, the
// closer the field is to a supply depot and the better the supply (1 to MaxSupply).
//
// The supply network is stored in the form of values within a grid, where each value
// indicates how far the supply network extends from the base. Each Tile's supply data is
// stored using a map, where the key is the player ID and the value is the corresponding
// supply value.
//
// Example of supply data storage for a Tile:
//
//	tile.Supply = map[uint8]int{
//	    1: 3,   // Player 1's supply value is 3
//	    2: 6,   // Player 2's supply value is 6
//	}
//
// The above example indicates that the Tile provides a supply value of 3 for Player 1
// and a supply value of 6 for Player 2.
func updateSupply(world *World) {
	if world == nil {
		return
	}

	// Clear supply values for all tiles
	for _, t := range world.TileList(0) {
		t.Supply = make(map[uint8]int)
	}

	// Iterate through all bases on the world map
	for _, base := range world.TileList(BASE) {
		if base.Owner == 0 {
			continue // ignore neutral bases
		}

		// Set supply value for own base
		base.Supply[base.Owner] = 1

		// Process all neighbors (map wide) within specified supply range
		for lvl, tmp := range world.ExtNeighbors(base, MaxSupply) {
			lvl += 1
			for _, tile := range tmp {

				// Set the best supply value
				value, ok := tile.Supply[base.Owner]
				if !ok || value > lvl {
					tile.Supply[base.Owner] = lvl
				}
			}
		}
	}
}

// updateVisibility updates the visibility of units on the game map based on their
// attributes. It clears and then recalculates the visibility map for each player's
// units, taking into account the unit's viewing range and other factors.
//
// The tiles can have exactly three states for the player: FogOfWar, NormalView and CloseView.
// This is determined by the visibility of the units (attributes view and closeView).
//
// Example of visibility data storage for a Tile:
//
//	tile.Visibility = map[uint8]int{
//	    1: 3,   // Player 1's vision on this tile is 3
//	    2: 6,   // Player 2's vision on this tile is 6
//	}
func updateVisibility(world *World) {
	if world == nil {
		return
	}

	// Clear visibility for all tiles
	for _, t := range world.TileList(0) {
		t.Visibility = make(map[uint8]int)
	}

	// Iterate through all units on the world map
	for _, tile := range world.Units(0) {
		unit := tile.Unit
		if unit == nil {
			continue // Skip non-occupied tiles
		}
		player := unit.Player

		// Set own visibility for the unit's player to CloseView
		tile.Visibility[player] = CloseView

		// Process visibility for all neighboring tiles (map-wide)
		for lvl, tmp := range world.ExtNeighbors(tile, unit.View) {
			for _, t := range tmp {
				set := FogOfWar
				if lvl < unit.CloseView {
					set = CloseView // Set visibility to CloseView
				} else if lvl < unit.View {
					set = NormalView // Set visibility to NormalView
				}

				// Set the best visibility value
				value, ok := t.Visibility[player]
				if set > 0 && (!ok || value < set) {
					t.Visibility[player] = set
				}
			}
		}
	}
}

// updateUnitAttributes updates the attributes of all units on the game map based on
// their type, tile, and supply. It calculates and assigns values such as view
// range, armor, firing range, speed, ammunition, and hidden status for each unit.
func updateUnitAttributes(world *World) {
	if world == nil {
		return
	}

	// Iterate through all units on the world map
	for _, tile := range world.Units(0) {
		unit := tile.Unit
		if unit == nil {
			continue // Skip non-occupied tiles
		}
		player := unit.Player
		supply := tile.Supply[player]

		// Get basic statistics for the unit (see stats)
		view, closeView, armour, fireRange, maxAmmunition, speed, fireSpeed, hidden := stats(unit.Type, tile.Type)

		// Update unit attributes with the calculated values
		unit.View = view
		unit.CloseView = closeView
		unit.Armour = armour
		unit.FireRange = fireRange
		unit.Speed = speed
		unit.FireSpeed = fireSpeed
		unit.Hidden = hidden

		// refill ammunition based on supply availability, considering maximum supply range
		if supply > 0 && supply <= MaxSupply {
			unit.Ammunition += (1 / (float32(supply) * 2 * 30)) * SupplySpeed
		}

		// Limit ammunition to the maximum allowed amount
		if unit.Ammunition > float32(maxAmmunition) {
			unit.Ammunition = float32(maxAmmunition)
		}
	}
}

// healUnits heals units stationed at bases (BASE) over time, gradually restoring their
// health and removing demoralization. Units located at bases are healed incrementally,
// and their demoralized status is fixed.
func healUnits(world *World) {
	if world == nil {
		return
	}

	// Iterate through all bases on the world map
	for _, base := range world.TileList(BASE) {
		unit := base.Unit

		// Check if a unit is stationed at the base and heal every 100 iteration
		if unit != nil && world.Iteration%100 == 0 {

			// heal from demoralized
			unit.Demoralized = false

			// Incrementally increase unit health up to a maximum of 100
			if unit.Health < 100 {
				unit.Health++ // Heal the unit
			}
		}
	}
}

// processMove handles the movement of units in the game world. It checks all units on
// the world map to determine if they are currently in the process of moving. If a unit
// is in the process of moving, it calculates the mid-point of the movement activity and
// determines whether the unit should be moved from its source tile to the destination tile.
func processMove(world *World) {
	if world == nil {
		return
	}

	// Iterate through all units on the world map
	for _, tile := range world.Units(0) {
		unit := tile.Unit
		if unit == nil {
			continue // no unit found -> skip
		}

		// Check if the unit is currently moving
		if unit.Activity == nil || unit.Activity.Name != MOVE {
			continue // nothing to do -> skip
		}

		// remove old activity if it has ended
		if unit.Activity.End < world.Iteration {
			unit.Activity = nil // disable
			continue            // my job is done -> skip
		}

		// Calculate mid-point of the movement activity
		diff := unit.Activity.End - unit.Activity.Start
		switchPoint := unit.Activity.Start + diff/2

		// Check if it's time to switch tiles
		if world.Iteration == switchPoint {

			// Retrieve source and destination tile coordinates
			fromX := unit.Activity.From[0]
			fromY := unit.Activity.From[1]
			toX := unit.Activity.To[0]
			toY := unit.Activity.To[1]

			// Get source and destination tiles
			from := world.Tile(fromX, fromY)
			to := world.Tile(toX, toY)

			// Check if the destination is already occupied
			if to.Unit != nil {
				resources.PlaySound(resources.Sounds.Error) // play error sound
				unit.Activity = nil                         // ABORT moving!
				continue                                    // my job is done -> skip
			}

			// MOVE UNIT
			to.Unit = from.Unit // move unit
			from.Unit = nil     // Clear source tile
		}
	}
}
