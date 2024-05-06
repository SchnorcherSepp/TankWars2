package ai2000

import (
	"TankWars2_AI/core"
	"TankWars2_AI/remote"
	"time"
)

func RunAI(c *remote.Client, disable bool) {
	CLIENT = c
	PLAYER = CLIENT.Player()
	WORLD = core.NewWorld(0, 0) // dummy world

	for {
		time.Sleep(50 * time.Millisecond)
		WORLD = CLIENT.Status()
		WORLD.CalculateShadowUnits(PLAYER)
		if !disable {
			unitLoop()
		}
		debug()
	}
}

func debug() {
	for _, tile := range WORLD.Units(0) {
		if tile != nil && tile.Unit != nil {
			for _, activity := range tile.Unit.IncomingFire(WORLD) {
				if activity != nil {
					//activity.MARK = true
					tile.Unit.MARK = true
				}
			}
		}
	}
}
