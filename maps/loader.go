package maps

import (
	"TankWars2/core"
	"encoding/json"
	"os"
)

// Loader is a function that loads a game world from a JSON file located at the specified path.
// It reads the JSON data, parses it into a core.World structure, and creates a new world with
// the specified dimensions. The function populates the new world's tiles and their attributes,
// including tile types and associated unit information.
func Loader(path string) (*core.World, error) {

	// Read JSON data from the file
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse the JSON data into a loadable world structure
	load := core.World{}
	if err := json.Unmarshal(b, &load); err != nil {
		return nil, err
	}

	// Create a new world based on the loaded dimensions
	world := core.NewWorld(load.XWidth, load.YHeight)

	// Read and populate relevant attributes for each tile in the loaded world
	for x := range load.Tiles {
		for y := range load.Tiles[x] {

			// Get tiles from the loaded and new worlds
			lTile := load.Tile(x, y)  // Tile from the loaded world
			wTile := world.Tile(x, y) // Tile from the new world
			if lTile == nil || wTile == nil {
				continue // Skip if either tile is invalid
			}

			// Set tile attributes
			wTile.Type = lTile.Type       // Tile type
			wTile.ImageID = lTile.ImageID // Tile ID for GUI images

			// Set unit if present in the loaded tile
			lUnit := lTile.Unit
			if lUnit != nil {
				// Create a new unit for the new world
				wTile.Unit = core.NewUnit(lUnit.Player, lUnit.Type)
			}
		}
	}

	// Return the loaded world, the count of unique players, and any error
	return world, nil
}
