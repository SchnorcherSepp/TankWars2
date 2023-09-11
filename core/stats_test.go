package core

import (
	"testing"
)

func Test_stats(t *testing.T) {
	tests := []struct {
		unitType              byte
		tileType              byte
		expectedView          int
		expectedCloseView     int
		expectedArmour        int
		expectedFireRange     int
		expectedMaxAmmunition int
		expectedSpeed         uint64
		expectedFireSpeed     uint64
		expectedHidden        bool
	}{
		// ERROR
		{0, BASE, 0, 0, 0, 0, 0, 0, 0, false},      // ERROR
		{0, DIRT, 0, 0, 0, 0, 0, 0, 0, false},      // ERROR
		{0, FOREST, 0, 0, 0, 0, 0, 0, 0, false},    // ERROR
		{0, GRASS, 0, 0, 0, 0, 0, 0, 0, false},     // ERROR
		{0, HILL, 0, 0, 0, 0, 0, 0, 0, false},      // ERROR
		{0, HOLE, 0, 0, 0, 0, 0, 0, 0, false},      // ERROR
		{0, MOUNTAIN, 0, 0, 0, 0, 0, 0, 0, false},  // ERROR
		{0, STRUCTURE, 0, 0, 0, 0, 0, 0, 0, false}, // ERROR
		{0, WATER, 0, 0, 0, 0, 0, 0, 0, false},     // ERROR
		{0, 0, 0, 0, 0, 0, 0, 0, 0, false},         // ERROR

		// ARTILLERY cases
		{ARTILLERY, BASE, 3, 3, 3, 0, 2, 150, 100, false},
		{ARTILLERY, DIRT, 3, 1, 1, 4, 2, 150, 100, false},
		{ARTILLERY, FOREST, 2, 1, 1, 4, 2, 180, 100, true},
		{ARTILLERY, GRASS, 3, 1, 1, 4, 2, 150, 100, false},
		{ARTILLERY, HILL, 4, 2, 1, 5, 2, 180, 100, false},
		{ARTILLERY, HOLE, 3, 1, 2, 4, 2, 180, 100, false},
		{ARTILLERY, MOUNTAIN, 4, 2, 1, 5, 2, 210, 100, false},
		{ARTILLERY, STRUCTURE, 3, 1, 3, 4, 2, 150, 100, true},
		{ARTILLERY, WATER, 3, 1, 1, 0, 2, 210, 100, false},
		{ARTILLERY, 0, 0, 0, 0, 0, 0, 0, 0, false}, // ERROR

		// TANK cases
		{TANK, BASE, 3, 3, 4, 0, 3, 70, 60, false},
		{TANK, DIRT, 3, 1, 2, 2, 3, 70, 60, false},
		{TANK, FOREST, 2, 1, 2, 2, 3, 84, 60, true},
		{TANK, GRASS, 3, 1, 2, 2, 3, 70, 60, false},
		{TANK, HILL, 4, 2, 2, 3, 3, 84, 60, false},
		{TANK, HOLE, 3, 1, 3, 2, 3, 84, 60, false},
		{TANK, MOUNTAIN, 4, 2, 2, 3, 3, 98, 60, false},
		{TANK, STRUCTURE, 3, 1, 4, 2, 3, 70, 60, true},
		{TANK, WATER, 3, 1, 2, 0, 3, 98, 60, false},
		{TANK, 0, 0, 0, 0, 0, 0, 0, 0, false}, // ERROR

		// SOLDIER cases
		{SOLDIER, BASE, 3, 3, 2, 0, 9, 90, 69, false},
		{SOLDIER, DIRT, 3, 1, 0, 1, 9, 90, 69, false},
		{SOLDIER, FOREST, 2, 1, 0, 1, 9, 90, 69, true},
		{SOLDIER, GRASS, 3, 1, 0, 1, 9, 90, 69, true},
		{SOLDIER, HILL, 4, 2, 0, 2, 9, 90, 69, false},
		{SOLDIER, HOLE, 3, 1, 1, 1, 9, 90, 69, false},
		{SOLDIER, MOUNTAIN, 4, 2, 0, 2, 9, 125, 69, false},
		{SOLDIER, STRUCTURE, 3, 1, 2, 1, 9, 90, 69, true},
		{SOLDIER, WATER, 3, 1, 0, 0, 9, 125, 69, false},
		{SOLDIER, 0, 0, 0, 0, 0, 0, 0, 0, false}, // ERROR
	}

	// check TILES and UNITS
	// rewrite test if this fail
	if len(TILES) != 9 || len(UNITS) != 3 {
		t.Errorf("TILES or UNITS changed: rewrite unit test")
	}

	// check function
	for _, test := range tests {
		view, closeView, armour, fireRange, maxAmmunition, speed, fireSpeed, hidden := stats(test.unitType, test.tileType)

		if view != test.expectedView ||
			closeView != test.expectedCloseView ||
			armour != test.expectedArmour ||
			fireRange != test.expectedFireRange ||
			maxAmmunition != test.expectedMaxAmmunition ||
			speed != test.expectedSpeed ||
			fireSpeed != test.expectedFireSpeed ||
			hidden != test.expectedHidden {
			t.Errorf("For unitType %c and tileType %c, expected (%d, %d, %d, %d, %d, %d, %d, %v), but got (%d, %d, %d, %d, %d, %d, %d, %v)",
				test.unitType, test.tileType,
				test.expectedView, test.expectedCloseView, test.expectedArmour, test.expectedFireRange, test.expectedMaxAmmunition, test.expectedSpeed, test.expectedFireSpeed, test.expectedHidden,
				view, closeView, armour, fireRange, maxAmmunition, speed, fireSpeed, hidden)
		}
	}
}
