package core

/*
  This file offers a collection of functions tailored for simulating attacks,
  calculating inflicted damage, and resolving combat scenarios within the
  game world, enhancing the strategic and tactical elements of the gameplay.
  The functions contained in this file are called by the Update() function.
*/

import (
	"slices"
	"sort"
)

// processFire handles the attacking of units within the game world. It checks
// all units on the world map to determine if they are engaged in a firing activity.
// If an attacker is firing, the function calculates the outcome of the attack,
// including potential damage to the target tile or unit. An attack always hits the
// terrain itself and all units on the tile.
func processFire(world *World) {
	if world == nil {
		return
	}

	// Iterate through all units on the world map
	for _, tile := range world.Units(0) {
		attacker := tile.Unit
		if attacker == nil {
			continue // no attacker unit found -> skip
		}
		activity := attacker.Activity
		iteration := world.Iteration

		// Skip: no activity or wrong activity
		if activity == nil || activity.Name != FIRE {
			continue // nothing to do -> skip
		}

		// Disable old activity if it has ended
		if activity.End < iteration {
			attacker.Activity = nil // Disable attacker's activity
			continue                // my job is done -> skip
		}

		// Calculate damage and apply attack effects
		if iteration == activity.End-1 {
			target := world.Tile(activity.To[0], activity.To[1])

			// Attack target tile or structure
			switch target.Type {
			case BASE:
				if rnd.Intn(5) == 0 {
					target.Owner = 0 // disable base
				}
			case STRUCTURE:
				if rnd.Intn(10) == 0 {
					target.Type = FOREST
				}
			case FOREST:
				if rnd.Intn(10) == 0 {
					target.Type = GRASS
				}
			case GRASS:
				if rnd.Intn(15) == 0 {
					target.Type = DIRT
				}
			case DIRT:
				if rnd.Intn(25) == 0 {
					target.Type = HOLE
				}
			}

			// Attack target unit
			targetUnit := target.Unit
			if targetUnit != nil {

				// calc and add damage to target unit
				damage, critical := calcDamage(attacker.Demoralized, targetUnit.Armour)
				targetUnit.Health -= damage
				if critical {
					targetUnit.Demoralized = critical
				}

				// Eliminate target unit if health is zero or negative
				if targetUnit.Health <= 0 {
					target.Unit = nil // Remove unit from tile
				}
			}
		}
	}
}

// calcDamage calculates the damage inflicted during an attack based on the attacker's
// and target's attributes. It takes into account whether the attacker is demoralized
// and the target's armor. The function returns the calculated damage value and a
// boolean indicating whether a critical hit occurred. The min. damage is 3.
func calcDamage(demoralized bool, armour int) (int, bool) {

	// Configuration for dice rolling
	const sides = 20 // Number of sides on each dice
	const dices = 3  // Number of dice to roll for both attacker and target
	diceDiff := 2    // Bonus number of dice for the attacker, if not demoralized

	// Adjust dice rolling if attacker is demoralized
	if demoralized {
		diceDiff = 0
	}

	// Roll dice for attacker and target
	attacker := rollDice(sides, dices+diceDiff, 0, 0)
	target := rollDice(sides, dices, armour, armour) // armor gives a static bonus and dice re-rolls

	// Calculate damage
	damage := attacker - target
	if damage < 3 {
		damage = 3 // min. damage
	}

	// Check for critical hit:
	// This code snippet calculates whether a critical hit occurs
	// during an attack based on the randomly generated 'flip'
	// value and the inflicted damage.
	//
	// A critical hit is determined by any of these conditions:
	// - A 5% chance if damage > 10 hp
	// - A 60% chance if damage > 30 hp
	// - A 100% chance if damage > 50 hp
	flip := rnd.Intn(100)
	critical := (flip < 5 && damage > 10) || (flip < 60 && damage > 30) || damage > 50

	// Return calculated damage and critical hit status
	return damage, critical
}

// rollDice simulates rolling a set of dice and calculating the total value. It generates
// a specified number of dice rolls (numDice + reRolls), each with a specified number of
// sides (sides). The rolls are then sorted in descending order, and the sum of the best
// rolls (specified by numDice) is calculated. A bonus value is added to the sum, and the
// final total is returned.
func rollDice(sides, numDice, bonus, reRolls int) int {
	dices := make([]int, numDice+reRolls)

	// Roll the dice (numDice + reRolls)
	for i := range dices {
		dices[i] = rnd.Intn(sides) + 1
	}

	// Sort the dice rolls in descending order (the best first)
	sort.Ints(dices)
	slices.Reverse(dices)

	// Calculate the sum of the best dice rolls (numDice)
	total := 0
	for i := 0; i < numDice; i++ {
		total += dices[i]
	}

	// Add the bonus value and return the total
	total += bonus
	return total
}
