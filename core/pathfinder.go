package core

/*
  This file provides the implementation of the A* algorithm for pathfinding on the hexagonal game world grid.
*/

import (
	"container/heap"
	"fmt"
	"math"
)

//--------  Struct  --------------------------------------------------------------------------------------------------//

// pathfinder implements the A* algorithm for pathfinding on the hexagonal game world grid.
type pathfinder struct {
	world    *World
	unitType byte
}

// node represents a node in the A* algorithm.
type node struct {
	tile   *Tile   // The tile associated with the node.
	parent *node   // The parent node in the path.
	g      float64 // The cost of the path from the start node to this node.
	h      float64 // The estimated cost from this node to the goal node (heuristic).
	f      float64 // The total cost of the path (f = g + h).
	opened bool    // Indicates if the node is in the open list.
	closed bool    // Indicates if the node is in the closed list.
	index  int     // Index of the node in the open list.
}

// FindPath performs the A* algorithm to find the best path between the two specified tiles.
func FindPath(world *World, unitType byte, startTile, goalTile *Tile) []*Tile {
	pf := &pathfinder{
		world:    world,
		unitType: unitType,
	}
	return pf.findPath(startTile, goalTile)
}

//--------  Getter  --------------------------------------------------------------------------------------------------//

// findPath performs the A* algorithm to find the best path between the two specified tiles.
func (pf *pathfinder) findPath(startTile, goalTile *Tile) []*Tile {

	// Create the start node with initial values.
	startNode := &node{
		tile:   startTile,
		parent: nil,
		g:      0,
		h:      pf.heuristic(startTile, goalTile),
		f:      0,
		opened: true,
		closed: false,
		index:  0,
	}

	// Initialize the priority queue (Open List) and add the start node.
	openList := make(openList, 0)
	closeList := make(map[string]bool)
	heap.Init(&openList)
	heap.Push(&openList, startNode)

	// Execute the A* algorithm while the Open List is not empty.
	for len(openList) > 0 {
		// Retrieve the node with the lowest F-score from the Open List.
		currentNode := heap.Pop(&openList).(*node)
		currentNode.opened = false
		currentNode.closed = true
		key := fmt.Sprintf("%d,%d", currentNode.tile.XCol, currentNode.tile.YRow)
		closeList[key] = true

		// Check if the current node is the goal tile.
		if currentNode.tile == goalTile {
			return pf.reconstructPath(currentNode) // Reconstruct the path back to the start node.
		}

		// Get the neighboring tiles of the current node.
		neighbors := pf.world.Neighbors(currentNode.tile)
		for _, neighborTile := range neighbors {
			// Check if the neighboring tile is not nil and can be traversed.
			if neighborTile != nil && (goalTile == neighborTile || pf.canPass(neighborTile)) { // canPass don't count for the goalTile
				// skip closed tiles
				key := fmt.Sprintf("%d,%d", neighborTile.XCol, neighborTile.YRow)
				_, ok := closeList[key]
				if ok {
					continue // closed
				}

				// Calculate the new G-score for the neighboring tile.
				g := currentNode.g + 1

				// Find the neighboring node in the Open List.
				neighborNode := pf.getNodeFromList(openList, neighborTile)
				if neighborNode == nil || g < neighborNode.g {
					// If the neighbor node is not in the Open List, add it.
					if neighborNode == nil {
						neighborNode = &node{
							tile:   neighborTile,
							parent: currentNode,
						}
					}

					// Update the values of the neighboring node.
					neighborNode.g = g
					neighborNode.h = pf.heuristic(neighborTile, goalTile)
					neighborNode.f = neighborNode.g + neighborNode.h

					if !neighborNode.opened {
						// If the neighbor node is not open, add it to the Open List.
						heap.Push(&openList, neighborNode)
						neighborNode.opened = true
					} else {
						// Otherwise, update the position of the neighbor node in the Open List.
						heap.Fix(&openList, neighborNode.index)
					}
				}
			}
		}
	}

	return nil // No path found.
}

//--------  Helper  --------------------------------------------------------------------------------------------------//

// heuristic calculates the heuristic value (estimated cost) between two tiles.
func (pf *pathfinder) heuristic(currentTile, goalTile *Tile) float64 {
	dx := math.Abs(float64(currentTile.XCol - goalTile.XCol))
	dy := math.Abs(float64(currentTile.YRow - goalTile.YRow))
	return dx + dy
}

// canPass determines if a unit can pass from the current tile to the neighbor tile.
// Return true if passable, false otherwise.
func (pf *pathfinder) canPass(neighborTile *Tile) bool {
	if neighborTile == nil {
		return false // error
	}

	// check tile type
	if pf.unitType != SOLDIER { // TANK and ARTILLERY
		if neighborTile.Type == MOUNTAIN || neighborTile.Type == STRUCTURE || neighborTile.Type == WATER {
			return false
		}
	}

	// check other units
	if neighborTile.Unit != nil {
		return false
	}

	// OK
	return true
}

// getNodeFromList retrieves a node from the open list based on the associated tile.
func (pf *pathfinder) getNodeFromList(list openList, tile *Tile) *node {
	for _, node := range list {
		if node.tile == tile {
			return node
		}
	}
	return nil
}

// reconstructPath reconstructs the path from the goal node to the start node.
func (pf *pathfinder) reconstructPath(node *node) []*Tile {
	path := make([]*Tile, 0)
	for node != nil {
		path = append([]*Tile{node.tile}, path...)
		node = node.parent
	}
	return path
}

//--------  OpenList  ------------------------------------------------------------------------------------------------//

// openList is a priority queue for the open nodes in the A* algorithm.
type openList []*node

// Len returns the number of elements in the priority queue.
func (list openList) Len() int { return len(list) }

// Less reports whether the element with index i should sort before the element with index j.
func (list openList) Less(i, j int) bool { return list[i].f < list[j].f }

// Swap swaps the elements with indexes i and j.
func (list openList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
	list[i].index = i
	list[j].index = j
}

// Push adds an element to the priority queue.
func (list *openList) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	n := len(*list)
	node := x.(*node)
	node.index = n
	*list = append(*list, node)
}

// Pop removes and returns the smallest element (according to Less) from the priority queue.
func (list *openList) Pop() interface{} {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	old := *list
	n := len(old)
	node := old[n-1]
	node.index = -1
	*list = old[0 : n-1]
	return node
}
