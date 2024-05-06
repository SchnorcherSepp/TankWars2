package core

/*
  This file primarily handles the mechanics of a hexagonal game world simulation. It defines
  the World structure to represent the game's grid, its dimensions, and iteration count.
  The core tasks addressed here involve tile management, tile interaction queries, and command
  execution for unit movement and firing. Additionally, functions for retrieving tiles,
  filtering tile lists, identifying neighboring tiles, and initiating movement and firing
  commands provide essential components for building a comprehensive simulation of unit-based
  gameplay in a hexagonal world setting.
*/

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

//--------  Struct  --------------------------------------------------------------------------------------------------//

// World represents the game world with its tiles and dimensions.
type World struct {
	lock sync.Mutex // Mutex for synchronization of concurrent access to the world.

	Tiles         [][]*Tile       // Two-dimensional slice of tile pointers representing the world's tiles.
	XWidth        int             // The width of the world in tiles.
	YHeight       int             // The height of the world in tiles.
	Reinforcement map[uint64]byte // reinforcement for all players with a base. key is iteration, value is unit type.

	Iteration uint64 // Current iteration (game time) of the world.
	Freeze    bool   // if true, the update function has no effect and the world remains frozen
}

// NewWorld creates a new game world with the specified dimensions and initializes its tiles.
// It takes the width (xWidth) and height (yHeight) of the world grid as parameters and
// returns a pointer to the newly created World struct. The function populates the grid with
// Tile objects using the NewTile function, effectively setting up a game world ready for
// further simulation and interaction.
func NewWorld(xWidth, yHeight int) *World {

	// build world
	world := &World{
		Tiles:   make([][]*Tile, xWidth),
		XWidth:  xWidth,
		YHeight: yHeight,
	}

	// init tiles
	for x := range world.Tiles {
		world.Tiles[x] = make([]*Tile, yHeight)
		for y := range world.Tiles[x] {
			world.Tiles[x][y] = NewTile(0, x, y)
		}
	}

	return world
}

//--------  Getter  --------------------------------------------------------------------------------------------------//

// Tile returns the Tile at the specified coordinates within the game world.
// It takes the x and y coordinates as parameters and returns a pointer to the Tile
// object located at the given position. If the coordinates are outside the valid
// range of the world grid, the function returns nil, indicating that no Tile
// exists at that position.
func (w *World) Tile(x, y int) *Tile {

	// check x and y
	if x < 0 || x >= len(w.Tiles) {
		return nil
	}
	if y < 0 || y >= len(w.Tiles[x]) {
		return nil
	}

	// return
	return w.Tiles[x][y]
}

// TileList returns a list of tiles that match the specified type within the game world.
// It takes a filter byte as a parameter to indicate the type of tiles to be included in
// the list. If the filter byte is set to 0, all tiles are included in the list.
func (w *World) TileList(typeFilter byte) []*Tile {
	ret := make([]*Tile, 0, w.XWidth*w.YHeight)

	// process all tiles
	for _, xList := range w.Tiles {
		for _, t := range xList {
			if t != nil {
				// use filter
				if typeFilter == 0 || t.Type == typeFilter {
					ret = append(ret, t)
				}
			}
		}
	}
	return ret
}

// Units returns a list of tiles that contain units belonging to the specified player
// within the game world. It takes a filter byte as a parameter to indicate the player
// whose units should be included in the list. If the filter byte is set to 0, units
// from all players are included in the list.
func (w *World) Units(playerFilter byte) []*Tile {
	ret := make([]*Tile, 0, 30)

	// process all tiles
	for _, t := range w.TileList(0) {
		if t != nil {
			unit := t.Unit
			if unit != nil {
				// use filter
				if playerFilter == 0 || unit.Player == playerFilter {
					ret = append(ret, t)
				}
			}
		}
	}
	return ret
}

// Neighbors returns a list of neighboring tiles for the given tile within the game world.
// It takes a Tile pointer as input and calculates the neighboring tiles based on hexagonal
// grid coordinates. The function checks the adjacent tiles in six directions: top left,
// top right, right, bottom right, bottom left, and left. It calculates the coordinates of
// these neighboring tiles and retrieves them using the Tile method of the World struct.
func (w *World) Neighbors(tile *Tile) []*Tile {
	var neighbors = make([]*Tile, 0, 6)

	// check input
	if nil == tile {
		return neighbors
	}

	// get position
	x := tile.XCol
	y := tile.YRow

	// set neighbors
	cor := y % 2 // shift every 2nd line
	topLeft := w.Tile(x-1+cor, y-1)
	topRight := w.Tile(x+cor, y-1)
	right := w.Tile(x+1, y)
	bottomRight := w.Tile(x+cor, y+1)
	bottomLeft := w.Tile(x-1+cor, y+1)
	left := w.Tile(x-1, y)

	for _, t := range []*Tile{topLeft, topRight, right, bottomRight, bottomLeft, left} {
		if t != nil {
			neighbors = append(neighbors, t)
		}
	}
	return neighbors
}

// ExtNeighbors returns a 2D slice of tiles representing neighboring tiles with an extended
// radius from the given tile within the game world. The function takes a Tile pointer and
// an integer radius as input and calculates a set of neighboring tiles up to the specified
// radius. It utilizes a breadth-first search algorithm to find all tiles within the given
// radius while avoiding duplicates. The calculated tiles are stored in a 2D slice, where
// each sub-slice represents tiles at a specific distance from the source tile.
func (w *World) ExtNeighbors(tile *Tile, radius int) [][]*Tile {
	var known = make(map[string]*Tile)
	var distance = make(map[string]int)
	var open = make([]*Tile, 0, 24)

	// init
	for _, add := range w.Neighbors(tile) {
		open = append(open, add)
	}

	// radius
	for n := 0; n < radius; n++ {

		// find all new neighbors
		var tmp = make([]*Tile, 0, 24)
		for _, t := range open {

			// skip nil tile
			if t == nil {
				continue
			}

			// skip known neighbor
			key := fmt.Sprintf("%d,%d", t.XCol, t.YRow)
			_, ok := known[key]
			if ok || (t.XCol == tile.XCol && t.YRow == tile.YRow) {
				continue
			}

			// save tile as known
			known[key] = t    // set tile
			distance[key] = n // set radius

			// add all neighbors from new tile
			for _, add := range w.Neighbors(t) {
				if add != nil {
					tmp = append(tmp, add)
				}
			}
		}

		// set new open list
		open = tmp

		// end?
		if len(open) == 0 {
			break
		}
	}

	// return values
	var ret = make([][]*Tile, radius)

	for key, t := range known {
		r := distance[key]
		if ret[r] == nil {
			ret[r] = make([]*Tile, 0)
		}
		ret[r] = append(ret[r], t)
	}
	return ret
}

// Clone creates a deep copy of a World struct using JSON serialization and deserialization.
// If any error occurs during the cloning process, nil is returned.
func (w *World) Clone() *World {
	w.lock.Lock()         // Acquire the lock to ensure thread safety
	defer w.lock.Unlock() // Release the lock when the function exits

	// Serialize the original Unit struct into JSON data
	origJSON, err := json.Marshal(w)
	if err != nil {
		return nil
	}

	// Create an empty Unit struct to hold the cloned data
	clone := World{}

	// Deserialize the JSON data into the clone struct
	if err = json.Unmarshal(origJSON, &clone); err != nil {
		return nil
	}

	// Return a pointer to the cloned Unit struct
	return &clone
}

// Json converts the World object to a JSON-formatted string.
func (w *World) Json() string {
	w.lock.Lock()         // Acquire the lock to ensure thread safety
	defer w.lock.Unlock() // Release the lock when the function exits

	// Marshal the World object to JSON format.
	b, err := json.Marshal(w)
	if err != nil {
		// Return the error message if serialization fails.
		return err.Error()
	} else {
		// Return the JSON string representation of the World object.
		return string(b)
	}
}

// PlayerCount returns the number of players in this world.
func (w *World) PlayerCount() int {
	playerCount := make(map[uint8]bool)

	for _, tile := range w.Units(0) {
		if tile != nil && tile.Unit != nil {
			playerCount[tile.Unit.Player] = true
		}
	}

	return len(playerCount)
}

//--------  Setter  --------------------------------------------------------------------------------------------------//

// Move initiates a movement command for a unit from one tile to another within the game world.
// The function takes three parameters: the starting tile ('from'), the target tile ('to'),
// and an optional player filter ('playerFilter') represented as a player ID (uint8).
// The 'playerFilter' parameter is used to restrict the move commands to be accepted only from
// the specified player while rejecting moves for other players.
//
// It returns the new target tile and an error if any issue arises during the process.
// The new target tile parameter corresponds to the 'to' input parameter, unless pathfinding is
// applied, in which case it is updated to reflect the first tile reached through pathfinding.
//
// The function performs the following steps:
//  1. Acquire the lock on the world using a mutex to ensure thread safety and release
//     the lock when the function exits using the 'defer' keyword.
//  2. Check the validity of the 'from' and 'to' input tiles and the player's eligibility.
//  3. Check if the unit is already performing an activity.
//  4. Check if the 'to' tile is a neighbor of the 'from' tile; if not, use pathfinding.
//  5. Check the target tile's validity based on unit and tile types.
//  6. Create a movement activity command ('Activity') for the unit and mark it as 'busy'.
//
// The function returns the new 'to' tile and nil error if the command is successfully executed.
// Otherwise, it returns nil and an error indicating the reason for failure.
func (w *World) Move(from, to *Tile, playerFilter uint8) (newTo *Tile, err error) {
	w.lock.Lock()         // Acquire the lock to ensure thread safety
	defer w.lock.Unlock() // Release the lock when the function exits

	// check input
	if from == nil || to == nil {
		return nil, errors.New("input is nil")
	}
	unit := from.Unit
	if unit == nil || (playerFilter != 0 && unit.Player != playerFilter) {
		return nil, errors.New("no player unit found")
	}

	// check activity
	if unit.Activity != nil {
		return nil, errors.New("unit is already processing a command")
	}

	// check neighbors
	ok := false
	for _, t := range w.Neighbors(from) {
		if t == to {
			ok = true // 'to' is a neighbor
			break
		}
	}
	if !ok {
		// target is not a neighbor
		// -> use path finding to find a new target
		//
		// Note: Since Pathfinding has all the information at this point, the positioning of enemy invisible
		//       units in choke points could be leaked. However, this triggers an irrevocable move and is
		//       only one square wide, so it's negligible.
		way := FindPath(w, unit.Type, from, to)

		// is there a path?
		if way != nil && len(way) > 1 {
			to = way[1]
		} else {
			return nil, errors.New("target is not a neighbor and no path was found")
		}
	}

	// check target tile
	// ATTENTION: If these game rules are changed, the pathfinding must also be adjusted!
	// (see pathfinder.canPass())
	if unit.Type != SOLDIER { // TANK and ARTILLERY
		if to.Type == MOUNTAIN || to.Type == STRUCTURE || to.Type == WATER {
			return nil, errors.New("invalid target for this unit")
		}
	}

	// set command
	unit.Activity = &Activity{
		Name:  MOVE,
		From:  [2]int{from.XCol, from.YRow},
		To:    [2]int{to.XCol, to.YRow},
		Start: w.Iteration,
		End:   w.Iteration + unit.Speed,
	}
	return to, nil
}

// Fire initiates a firing command for a unit from one tile to another within the game world.
// The function takes two Tile pointers, 'from' and 'to', as input and processes the following steps:
// - It locks the world to prevent concurrent access.
// - It checks the validity of 'from' and 'to' inputs.
// - It retrieves the unit associated with the 'from' tile.
// - It verifies that the unit is not already processing a command.
// - It checks if the 'to' tile is within the firing range of the 'from' tile based on the unit's attributes.
// - It checks if the unit has ammunition available for firing.
// - It decrements the unit's ammunition count by one.
// - It sets a firing command for the unit, specifying the start and end iterations.
// - It returns an error if any of the checks fail or if there's an issue with the input.
func (w *World) Fire(from, to *Tile, playerFilter uint8) error {
	w.lock.Lock()         // Acquire the lock to ensure thread safety
	defer w.lock.Unlock() // Release the lock when the function exits

	// check input
	if from == nil || to == nil {
		return errors.New("input is nil")
	}
	unit := from.Unit
	if unit == nil || (playerFilter != 0 && unit.Player != playerFilter) {
		return errors.New("no player unit found")
	}

	// check activity
	if unit.Activity != nil {
		return errors.New("unit is already processing a command")
	}

	// check neighbors
	ok := false
	neighbors := w.ExtNeighbors(from, unit.FireRange)
	for _, tmp := range neighbors {
		for _, t := range tmp {
			if t == to {
				ok = true // 'to' is a neighbor
				break
			}
		}
	}
	if !ok {
		return errors.New("target is not in range")
	}

	// ammunition
	if unit.Ammunition < 1 {
		return errors.New("no ammunition")
	}
	unit.Ammunition -= 1 // fire one ammunition

	// set command
	unit.Activity = &Activity{
		Name:  FIRE,
		From:  [2]int{from.XCol, from.YRow},
		To:    [2]int{to.XCol, to.YRow},
		Start: w.Iteration,
		End:   w.Iteration + unit.FireSpeed,
	}
	return nil
}

//--------  Helper  --------------------------------------------------------------------------------------------------//

// Censorship applies censorship or information restriction to a given game world.
// The function takes a world and a player ID (represented as an uint8) as parameters and
// returns a modified copy of the game world.
//
// The function performs the following steps:
//  1. The original game world is cloned using the Clone method to obtain an independent
//     copy of the world for editing. If cloning fails for any reason, nil is returned.
//
// 2. The function iterates through all tiles in the copied world and processes them accordingly.
//
// 3. For each tile, the following actions are taken:
//   - Delivery data for other players is removed, retaining data only for the specified player.
//   - Visibility data for other players is removed, retaining data only for the specified player.
//   - For tiles that are in Fog of War visibility mode for the specified player and have no visibility,
//     the owner is reset (if it doesn't belong to the specified player), and any hidden units are removed.
//   - For tiles in normal view mode that have hidden units, these hidden units are removed.
//
// 4. The edited copied world returned.
//
// Overall, the function enforces a form of visibility restriction and information withholding
// for the specified player in the game world.
func Censorship(world *World, player uint8) *World {

	// Clone the original game world to work on an independent copy.
	world = world.Clone()
	if world == nil {
		return nil
	}

	// Iterate through all tiles and apply information restriction.
	for _, t := range world.TileList(0) {

		// Remove delivery data for other players and retain data only for the specified player.
		supply := make(map[uint8]int)
		supply[player] = t.Supply[player]
		t.Supply = supply

		// Remove visibility data for other players and retain data only for the specified player.
		Visibility := make(map[uint8]int)
		Visibility[player] = t.Visibility[player]
		t.Visibility = Visibility

		// Handle tiles in Fog of War visibility mode for the specified player.
		if vis := t.Visibility[player]; vis == FogOfWar {
			// hide other base owner
			if t.Owner != player {
				t.Owner = 0
			}
			// hide unit in fog of war
			t.Unit = nil

		} else if vis == NormalView && t.Unit != nil && t.Unit.Hidden == true {
			// hide hidden unit
			t.Unit = nil
		}
	}

	// Return the edited game world.
	return world
}
