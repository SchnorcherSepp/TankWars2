package core

/*
  This file defines the base attribute of the units and the tile attribute bonus.
  This allows an essential part of the game balance to be configured.
  The functions contained in this file are called by the Update() function.
*/

// stats calculates and returns essential attributes for a military unit, considering its
// unit type (see UNITS) and the terrain type (see TILES) it occupies. It determines
// the unit's visibility capabilities, defensive properties, attack range, ammunition status,
// mobility, and camouflage potential.
//
// Parameters:
//
//	unitType - The type of military unit (ARTILLERY, TANK, SOLDIER).
//	tileType - The type of terrain on which the unit is positioned (BASE, DIRT, FOREST, etc.).
//
// Returns:
//
//	view - The distance in tiles that enemy units can be seen.
//	closeView - The visibility range for camouflaged enemy units (usually smaller than 'view').
//	armour - The unit's armor value, which includes both its base armor and terrain bonus.
//	fireRange - The range of tiles over which the unit can launch attacks.
//	maxAmmunition - The unit's ammunition level. One ammunition is consumed per attack and
//	                replenishes slowly from supply depots.
//	speed - The number of rounds required for the unit to move to an adjacent tile.
//	fireSpeed - The interval between two consecutive attacks.
//	hidden - Indicates whether the unit is camouflaged due to terrain bonuses. Camouflaged
//	         units can only be detected by enemies within closeView range.
func stats(unitType, tileType byte) (view, closeView, armour, fireRange, maxAmmunition int, speed, fireSpeed uint64, hidden bool) {

	// view for all units
	view = 3
	closeView = 1

	// unit base value
	switch unitType {
	case ARTILLERY:
		armour = 1
		fireRange = 4
		maxAmmunition = 2
		speed = 150     // Iteration
		fireSpeed = 100 // Iteration

	case TANK:
		armour = 2
		fireRange = 2
		maxAmmunition = 3
		speed = 70     // Iteration
		fireSpeed = 60 // Iteration

	case SOLDIER:
		armour = 0
		fireRange = 1
		maxAmmunition = 9
		speed = 90     // Iteration
		fireSpeed = 69 // Iteration

	default:
		return 0, 0, 0, 0, 0, 0, 0, false // ERROR
	}

	//-----------------------------------------------

	// bonus value by tile
	switch tileType {
	case BASE:
		armour += 2
		fireRange = 0  // disable weapon (range = 0)!
		closeView += 2 // base can scan hidden units

	case DIRT:
		// no bonus

	case FOREST:
		hidden = true // everyone is hidden in the forest
		view -= 1
		if unitType != SOLDIER {
			speed = uint64(float64(speed) * 1.2) // reduce speed for tanks
		}

	case GRASS:
		if unitType == SOLDIER {
			hidden = true // soldiers are hidden in the grass
		}

	case HILL:
		fireRange += 1
		view += 1
		closeView += 1
		if unitType != SOLDIER {
			speed = uint64(float64(speed) * 1.2) // reduce speed for tanks
		}

	case HOLE:
		armour += 1
		if unitType != SOLDIER {
			speed = uint64(float64(speed) * 1.2) // reduce speed for tanks
		}

	case MOUNTAIN:
		fireRange += 1
		view += 1
		closeView += 1
		speed = uint64(float64(speed) * 1.4)

	case STRUCTURE:
		armour += 2
		hidden = true // everyone is hidden in a building

	case WATER:
		fireRange = 0 // disable weapon (range = 0)!
		speed = uint64(float64(speed) * 1.4)

	default:
		return 0, 0, 0, 0, 0, 0, 0, false // ERROR
	}

	return
}
