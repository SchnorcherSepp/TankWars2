package core

/*
  This file contains constants and settings related to the game mechanics.
  These definitions provide the foundational elements for the game's core functionality and interactions.
*/

import (
	"math/rand"
	"time"
)

// game settings
const (
	GameSpeed   = 30  // Number of iterations per second
	MaxSupply   = 15  // Maximum supply distance
	SupplySpeed = 1.0 // This factor affecting the rate of ammunition regeneration
)

// tile types
const (
	BASE      = 'B'
	DIRT      = 'D'
	FOREST    = 'F'
	GRASS     = 'G'
	HILL      = 'H'
	HOLE      = 'O'
	MOUNTAIN  = 'M'
	STRUCTURE = 'S'
	WATER     = 'W'
)

// TILES is a list of all tiles
var TILES = []byte{BASE, DIRT, FOREST, GRASS, HILL, HOLE, MOUNTAIN, STRUCTURE, WATER}

// unit types
const (
	ARTILLERY = 'A'
	SOLDIER   = 'U'
	TANK      = 'T'
)

// UNITS is a list of all units
var UNITS = []byte{ARTILLERY, SOLDIER, TANK}

// player
const (
	RED = iota + 1
	BLUE
	GREEN
	YELLOW
	WHITE
	BLACK
)

// PLAYERS is a list of all players
var PLAYERS = []byte{RED, BLUE, GREEN, YELLOW, WHITE, BLACK}

// visibility
const (
	FogOfWar   = 0 // no sight of enemy units
	NormalView = 1 // normal enemy units are visible
	CloseView  = 2 // all units visible, including cloaked ones
)

// activity
const (
	MOVE = "MOVE" // Move activity command name
	FIRE = "FIRE" // Fire activity command name
)

// inti random
var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
