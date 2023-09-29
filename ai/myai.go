package ai

import (
	"github.com/SchnorcherSepp/TankWars2/core"
	"github.com/SchnorcherSepp/TankWars2/remote"
	"math/rand"
	"time"
)

// unitMemory is a structure used within the AI package to store target coordinates for units.
// It contains targetX and targetY, representing the X and Y coordinates of the unit's assigned target location.
// This memory mechanism enables AI-controlled units to retain their objectives and make informed decisions
// during AI simulations. The RunAI function utilizes this memory to determine appropriate actions such as
// selecting new targets, issuing firing commands, and executing movement toward the chosen target.
type unitMemory struct {
	targetX int
	targetY int
}

// RunAI simulates an AI-controlled player by continuously making decisions for units.
// The function takes a 'client' object representing the remote client of the game as a parameter.
//
// The function performs the following steps in a loop:
//  1. Retrieves the current state of the game world using the 'client.Status()' method.
//  2. Identifies all enemy bases on the map and populates the 'targets' slice with them.
//  3. If no enemy bases are left, the AI loop continues to the next iteration.
//  4. Iterates through all units belonging to the AI-controlled player.
//  5. Skips units with existing commands or units with ongoing activities.
//  6. Checks the memory for the current unit's target. If no target base is set or the base owner is now
//     the AI player, it selects a new target from the 'targets' slice and updates the memory accordingly.
//  7. Checks for enemies within the unit's firing range and initiates a 'Fire' command if found.
//  8. Checks for visible enemies within the unit's extended view range and initiates a 'Move' command
//     towards them, overriding the base target.
//  9. If no enemies are found in the extended view range, the unit moves towards its original target base.
//
// The function simulates AI decision-making by considering firing at enemies within range,
// moving towards visible enemies, and finally moving towards the chosen target.
func RunAI(client *remote.Client) {

	// Create a memory to store target coordinates for units.
	aiMemory := make(map[int]unitMemory)
	player := client.Player() // Get the AI player's ID.

	// Main AI loop
	for {
		time.Sleep(50 * time.Millisecond) // Prevent server denial of service (DoS) by pacing requests.
		world := client.Status()          // Get the current state of the game world from the server.

		// Get all enemy bases on the map.
		targets := make([]*core.Tile, 0, 8)
		for _, t := range world.TileList(core.BASE) {
			if t != nil && t.Owner != player {
				targets = append(targets, t)
			}
		}

		// If there are no enemy bases left, skip to the next AI loop iteration.
		if len(targets) == 0 {
			continue // NEXT AI LOOP
		}

		// Iterate through all AI-controlled units.
	UnitLoop:
		for _, tile := range world.Units(player) {

			// Skip units with existing commands, nil units, or units with ongoing activities.
			if tile == nil || tile.Unit == nil || tile.Unit.Activity != nil {
				continue UnitLoop // NEXT UNIT
			}
			unit := tile.Unit

			// Check or set the unit's target memory.
			um, ok := aiMemory[unit.ID]
			if !ok || world.Tile(um.targetX, um.targetY).Owner == player {
				// No target found or target is captured by the AI player.
				rand.Shuffle(len(targets), func(i, j int) {
					targets[i], targets[j] = targets[j], targets[i]
				})
				um.targetX = targets[0].XCol
				um.targetY = targets[0].YRow
				aiMemory[unit.ID] = um
			}
			target := world.Tile(um.targetX, um.targetY)

			if unit.Ammunition >= 0.8 {

				// Check for enemies within firing range and initiate 'Fire' command if found.
				for _, tmp := range world.ExtNeighbors(tile, unit.FireRange) {
					for _, t := range tmp {
						if t != nil && t.Unit != nil && t.Unit.Player != player {
							_ = client.Fire(tile.XCol, tile.YRow, t.XCol, t.YRow)
							continue UnitLoop // NEXT UNIT
						}
					}
				}

				// Check for visible enemies within extended view range and initiate
				// 'Move' command towards them, overriding the base target.
				for _, tmp := range world.ExtNeighbors(tile, unit.View+2) {
					for _, t := range tmp {
						// all tiles in view range
						if t != nil && t.Unit != nil && t.Unit.Player != player {
							_ = client.Move(tile.XCol, tile.YRow, t.XCol, t.YRow)
							continue UnitLoop // NEXT UNIT
						}
					}
				}
			}

			// Move towards the chosen target (enemy base or AI-selected target).
			_ = client.Move(tile.XCol, tile.YRow, target.XCol, target.YRow)
		}
	}
}
