package ai2000

import (
	"TankWars2_AI/core"
	"TankWars2_AI/remote"
)

var CLIENT *remote.Client

var PLAYER uint8

var WORLD *core.World

var MEMORYxTarget map[int]*core.Tile
var MEMORYxUnitText map[int]string

// -----------------------------------------------------

func init() {
	MEMORYxTarget = make(map[int]*core.Tile)
	MEMORYxUnitText = make(map[int]string)
}
