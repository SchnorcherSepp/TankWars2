package ai

import (
	"TankWars2_AI/core"
	"TankWars2_AI/remote"
	"time"
)

func RunAI(c *remote.Client) {
	CLIENT = c
	PLAYER = CLIENT.Player()
	WORLD = core.NewWorld(0, 0) // dummy world
	OBSERVER = core.NewObserver(PLAYER, WORLD)

	for {
		time.Sleep(1 * time.Millisecond)
		WORLD, _ = CLIENT.Status()
		OBSERVER.InspectNextWorld(WORLD) // magic!!!!
		CLIENT.SetGuiWorld(WORLD)
		unitLoop()
		debug()
	}
}

func debug() {
	for _, tile := range WORLD.Units(0) {
		if tile != nil && tile.Unit != nil {
			for _, activity := range tile.Unit.IncomingFire(WORLD) {
				if activity != nil {
					activity.MARK = true
				}
			}
		}
	}
}
