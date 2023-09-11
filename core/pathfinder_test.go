package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestFindPath tests the findPath function of the pathfinder.
func TestFindPath(t *testing.T) {
	world := NewWorld(10, 10)
	world.Tile(1, 0).Type = MOUNTAIN
	world.Tile(1, 1).Type = WATER
	world.Tile(1, 2).Type = STRUCTURE

	startTile := world.Tile(0, 0)
	goalTile := world.Tile(2, 0)

	path := FindPath(world, TANK, startTile, goalTile)
	path2 := FindPath(world, SOLDIER, startTile, goalTile)

	// tests
	assert.NotNil(t, path)
	assert.NotNil(t, path2)
	assert.Equal(t, 8, len(path))
	assert.Equal(t, 3, len(path2))
}

// TestHeuristic tests the heuristic function of the pathfinder.
func TestHeuristic(t *testing.T) {
	pathfinder := &pathfinder{
		world:    nil,
		unitType: SOLDIER,
	}
	currentTile := &Tile{XCol: 0, YRow: 0}
	goalTile := &Tile{XCol: 3, YRow: 4}

	heuristicValue := pathfinder.heuristic(currentTile, goalTile)

	// tests
	assert.Equal(t, 7.0, heuristicValue)
}

// TestCanPass tests the canPass function of the pathfinder.
func TestCanPass(t *testing.T) {
	pathfinder := &pathfinder{
		world:    nil,
		unitType: TANK,
	}
	passableTile := &Tile{Type: GRASS}
	blockingTile := &Tile{Type: MOUNTAIN}

	passable := pathfinder.canPass(passableTile)
	blocking := pathfinder.canPass(blockingTile)

	// tests
	assert.True(t, passable)
	assert.False(t, blocking)
}

// TestGetNodeFromList tests the getNodeFromList function of the pathfinder.
func TestGetNodeFromList(t *testing.T) {
	node1 := &node{tile: &Tile{}}
	node2 := &node{tile: &Tile{}}
	node3 := &node{tile: &Tile{}}
	nodeList := openList{node1, node2, node3}

	pathfinder := &pathfinder{
		world:    nil,
		unitType: SOLDIER,
	}
	tileToFind := node2.tile

	foundNode := pathfinder.getNodeFromList(nodeList, tileToFind)

	// tests
	assert.NotNil(t, foundNode)
	assert.Equal(t, node2, foundNode)
}
