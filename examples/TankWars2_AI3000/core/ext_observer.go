package core

import (
	"fmt"
)

//--------  Struct  --------------------------------------------------------------------------------------------------//

type Observer struct {
	PLAYER    uint8
	lastWorld *World
	nextWorld *World

	myUnits           map[int]*Unit
	unknownEnemyUnits map[int]*Unit
	enemyUnits        map[int]*Unit
}

func NewObserver(PLAYER uint8, world *World) *Observer {
	observer := &Observer{
		PLAYER:            PLAYER,
		lastWorld:         world,
		nextWorld:         world,
		myUnits:           make(map[int]*Unit),
		unknownEnemyUnits: make(map[int]*Unit),
		enemyUnits:        make(map[int]*Unit),
	}
	observer.InspectNextWorld(world)
	fmt.Printf("Start Observer for player %d\n", PLAYER)
	return observer
}

//--------  Setter  --------------------------------------------------------------------------------------------------//

func (o *Observer) InspectNextWorld(nextWorld *World) {

	// skip calculations
	if nextWorld == nil {
		return // skip invalid world
	}
	if nextWorld.Iteration < 5 {
		return // skip first world
	}
	if nextWorld == o.nextWorld { // check by ref !!!!
		return // skip same world
	}

	// switch worlds
	o.lastWorld = o.nextWorld
	o.nextWorld = nextWorld

	//fmt.Printf("Observer iteration: %d\n", it)
	//##################################################################################

	// spawn detection && unit inventory
	//----------------------------------------------------------------------------------
	o.UpdateUnitInventory()

	// generate phantom from last seen units (and set to world)
	//----------------------------------------------------------------------------------
	o.AddPhantoms()

	// calc enemy supply
	// code from updateSupply(nextWorld)
	//----------------------------------------------------------------------------------
	updateSupply(nextWorld) // todo: einheiten beobahten und supply daras beechnen und basen setzen  ANACH supply berechnen

	// calc enemy visibility
	// code from updateVisibility(nextWorld)
	//----------------------------------------------------------------------------------
	updateVisibility(nextWorld)

	// abnormal movement
	//----------------------------------------------------------------------------------
	// TODO: soll/IST bewegung prüfen

	// hidden attack
	//----------------------------------------------------------------------------------
	// TODO: hp verloren, basis verloren, tale ändert sich
}

//______________________________________________________________________________________________________________________

func (o *Observer) UpdateUnitInventory() {

	// SET LAST SEEN                                               [all units]
	//------------------------------------------------------------------------
	world := o.nextWorld
	for _, t := range world.Units(0) {
		// i can see this unit
		u := t.Unit
		u.ObserverVisible = true
		u.ObserverLastSeen = world.Iteration
		u.ObserverPhantom = 0
	}

	// SPAWN: find own new units  (add mirror to unknown enemy)      [myUnits]
	//------------------------------------------------------------------------
	for _, t := range world.Units(o.PLAYER) {
		u := t.Unit
		if _, ok := o.myUnits[u.ID]; !ok {
			// a new unit is found
			o.myUnits[u.ID] = u
			// create an enemy unit (mirror)
			o.unknownEnemyUnits[u.ID] = u
			fmt.Printf("SPAWN UNIT:\t\t %c with id  %d\n", u.Type, u.ID)
		}
		// update my unit stats
		o.myUnits[u.ID] = u
	}

	// UNIT LOST: remove own killed units                            [myUnits]
	//------------------------------------------------------------------------
	for key, u := range o.myUnits {
		if u.ObserverLastSeen < world.Iteration {
			delete(o.myUnits, key)
			fmt.Printf("UNIT LOST:\t\t %c with id  %d\t\t\t[my=%d|unk=%d|en=%d]\n", u.Type, u.ID, len(o.myUnits), len(o.unknownEnemyUnits), len(o.enemyUnits))
		}
	}

	// SEARCH: spott enemy unit                                   [enemyUnits]
	//------------------------------------------------------------------------
	for _, t := range world.Units(0) {
		u := t.Unit
		if u.Player != o.PLAYER {
			// is enemy
			if _, ok := o.enemyUnits[u.ID]; !ok {
				// a new enemy unit spotted
				// take unit from unknown and move to known
				spottError := true
				for key, unk := range o.unknownEnemyUnits {
					if unk.Type == u.Type {
						// add new enemy unit
						o.enemyUnits[u.ID] = u
						// remove from unknown units
						delete(o.unknownEnemyUnits, key)
						fmt.Printf("ENEMY SPOTTED:\t %c with id  %d\t\t\t[my=%d|unk=%d|en=%d]\n", u.Type, u.ID, len(o.myUnits), len(o.unknownEnemyUnits), len(o.enemyUnits))
						spottError = false
						break
					}
				}
				// not enough unknown to spott this unit!
				if spottError {
					fmt.Printf("SPOTT ERROR!!:\t %c with id  %d\t\t\t[my=%d|unk=%d|en=%d]\n", u.Type, u.ID, len(o.myUnits), len(o.unknownEnemyUnits), len(o.enemyUnits))
				}
			}
			// update enemy unit stats
			o.enemyUnits[u.ID] = u
		}
	}

	// UPDATE ObserverVisible (enemy units) (just away)           [enemyUnits]
	//------------------------------------------------------------------------
	for _, u := range o.enemyUnits {
		if u.ObserverVisible && u.ObserverLastSeen < world.Iteration {
			// this enemy unit is not visible now!
			u.ObserverVisible = false
			//fmt.Printf("HIDE UNIT:\t\t %c with id  %d\n", u.Type, u.ID)  // DEBUG

			// confirmed destroyed flag
			//  GUI: Destroyed=true && Lost=false
			//-------------------------------------
			// - enemy unit is not visible
			// - this happened just now
			tile := world.Tile(u.tile.XCol, u.tile.YRow)

			// maybe moved away?
			if u.Activity != nil && u.Activity.Name == MOVE {
				if u.Activity.SwitchPoint() < world.Iteration {
					// change tile
					tile = world.Tile(u.Activity.To[0], u.Activity.To[1])
				}
			}

			// can we see the tile (old or new one from moving)
			if tile != nil {
				// CloseView
				if tile.Visibility[o.PLAYER] == CloseView {
					// the unit is disappeared, but i can see the tile
					delete(o.enemyUnits, u.ID)
					fmt.Printf("KILL 1 UNIT:\t %c with id  %d\t\t\t[my=%d|unk=%d|en=%d]\n", u.Type, u.ID, len(o.myUnits), len(o.unknownEnemyUnits), len(o.enemyUnits))
				}
				// NormalView and not hidden
				if tile.Visibility[o.PLAYER] == NormalView {
					switch u.Type {
					case TANK, ARTILLERY:
						if tile.Type != FOREST {
							// the unit is disappeared, but i can see the tile
							delete(o.enemyUnits, u.ID)
							fmt.Printf("KILL 2 UNIT:\t %c with id  %d\t\t\t[my=%d|unk=%d|en=%d]\n", u.Type, u.ID, len(o.myUnits), len(o.unknownEnemyUnits), len(o.enemyUnits))
						}
					case SOLDIER:
						if tile.Type != FOREST && tile.Type != GRASS && tile.Type != STRUCTURE {
							// the unit is disappeared, but i can see the tile
							delete(o.enemyUnits, u.ID)
							fmt.Printf("KILL 3 UNIT:\t %c with id  %d\t\t\t[my=%d|unk=%d|en=%d]\n", u.Type, u.ID, len(o.myUnits), len(o.unknownEnemyUnits), len(o.enemyUnits))
						}
					}
				}
			}

		}
	}
}

func (o *Observer) AddPhantoms() {
	// process all invisible, enemy units ==> can be phantoms
	w := o.nextWorld
	for _, su := range o.enemyUnits {
		if !su.ObserverVisible {
			//----------------------------------------------------------------------------------------------------------

			/*
			 *  Enemy, invisible units, killed out of sight,
			 *  can become phantoms.
			 *
			 *  Phantoms must add to the world constantly,
			 *  or the next world update will remove them!
			 */

			// there is a hard limit: 2* speed
			if su.ObserverLastSeen+2*su.Speed <= w.Iteration {
				continue // --> skip this unit!!
			}

			/*  Units that have recently disappeared are processed  */

			// Mark as phantom and add to world if the tile is empty
			if su.ObserverPhantom == 0 {
				// in the first phase we known the position exactly
				su.ObserverPhantom = 100
			}

			// remove old activity if it has ended
			if su.Activity != nil && su.Activity.End < w.Iteration {
				su.Activity = nil
			}

			// If a unit is in the process of moving, it calculates the mid-point of the movement
			// activity and determines whether the unit should be moved from its source tile to
			// the destination tile.
			if su.IsMoving() {
				if su.Activity.SwitchPoint() > w.Iteration {
					su.tile.XCol = su.Activity.From[0]
					su.tile.YRow = su.Activity.From[1]
				} else {
					su.tile.XCol = su.Activity.To[0]
					su.tile.YRow = su.Activity.To[1]
				}
			}

			// is the phantom has no action, add a new one
			if !su.IsBusy() {
				su.ObserverPhantom = 99 // now the position is not safe to say
				addPhantomAiAction(su, w)
			}

			// update hidden status
			tile := w.Tile(su.XCol(), su.YRow())
			_, _, _, _, _, _, _, hidden := stats(su.Type, tile.Type)
			su.Hidden = hidden

			// creat phantom in the world
			if tile.Unit == nil {

				// phantoms cant exist in the "light"
				visibility := tile.Visibility[o.PLAYER]
				if visibility == FogOfWar || (visibility == NormalView && hidden) {
					tile.Unit = su // ADD ObserverPhantom!!
				}
			}

			//----------------------------------------------------------------------------------------------------------
		}
	}
}

func addPhantomAiAction(unit *Unit, world *World) {
	tile := unit.tile

	if unit.Ammunition >= 0.8 {

		// Check for enemies within firing range and initiate 'Fire' command if found.
		for _, tmp := range world.ExtNeighbors(tile, unit.FireRange) {
			for _, t := range tmp {
				if t != nil && t.Unit != nil && t.Unit.Player != unit.Player {
					_ = world.Fire(tile, t, 0)
					return // --> EXIT
				}
			}
		}

		// Check for visible enemies within extended view range and initiate
		// 'Move' command towards them, overriding the base target.
		for _, tmp := range world.ExtNeighbors(tile, unit.View+2) {
			for _, t := range tmp {
				// all tiles in view range
				if t != nil && t.Unit != nil && t.Unit.Player != unit.Player {
					_, _ = world.Move(tile, t, 0)
					return // --> EXIT
				}
			}
		}
	}

	// Move towards the chosen target (enemy base or AI-selected target).
	t := unit.NearestAllBase(world, false)
	_, _ = world.Move(tile, t, 0)
}
