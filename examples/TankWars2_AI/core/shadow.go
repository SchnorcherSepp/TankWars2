package core

import (
	"fmt"
	"sort"
)

var ShadowUnits map[int]*ShadowUnit
var ShadowUnitsSortList []*ShadowUnit

func init() {
	ShadowUnits = make(map[int]*ShadowUnit)
	ShadowUnitsSortList = make([]*ShadowUnit, 0)
}

type ShadowUnit struct {
	Unit      *Unit
	LastSeen  uint64
	Visible   bool
	Lost      bool
	Destroyed bool
}

//--------------------------------------------------------------------------------------------------------------------//

// CalculateShadowUnits creates phantom entities in the world.
// This means that even invisible units can be seen walking if you know they are there.
func (w *World) CalculateShadowUnits(player uint8) {

	// reset visible flag
	//---------------------------------------------------------------------------
	for _, v := range ShadowUnits {
		v.Visible = false
	}

	// update all visible units
	//---------------------------------------------------------------------------
	for _, tile := range w.Units(0) {
		// new unit -> create shadow unit
		su, ok := ShadowUnits[tile.Unit.ID]
		if !ok {
			su = new(ShadowUnit)
		}
		// skip phantoms
		if su != nil && su.Unit != nil && su.Unit.PHANTOM == w.Iteration {
			continue // -> skip
		}
		// undead warning
		if su.Destroyed && !su.Lost {
			fmt.Printf("WARNUNG: undead unit: %s (%d) with %d health on %s by %v\n", string(su.Unit.Type), su.Unit.ID, su.Unit.Health, string(tile.Type), su.Unit.Activity)
		}
		// update values for visible units
		su.Unit = tile.Unit
		su.LastSeen = w.Iteration
		su.Visible = true
		su.Lost = false
		su.Destroyed = false
		// set shadow unit
		ShadowUnits[tile.Unit.ID] = su
	}

	// confirmed destroyed flag
	//  GUI: Destroyed=true && Lost=false
	//---------------------------------------------------------------------------
	for _, su := range ShadowUnits {
		if !su.Visible && !su.Destroyed && !su.Lost && su.LastSeen+5 > w.Iteration {
			unit := su.Unit
			tile := w.Tile(unit.tile.XCol, unit.tile.YRow)

			// maybe moved away?
			if unit.Activity != nil && unit.Activity.Name == MOVE {
				if unit.Activity.SwitchPoint() < w.Iteration {
					// change tile
					tile = w.Tile(unit.Activity.To[0], unit.Activity.To[1])
				}
			}

			// can we see the tile
			if tile != nil {
				// CloseView
				if tile.Visibility[player] == CloseView {
					su.Destroyed = true
					fmt.Printf("kill CloseView: %s (%d) with %d health on %s by %v\n", string(su.Unit.Type), su.Unit.ID, su.Unit.Health, string(tile.Type), su.Unit.Activity)
				}
				// NormalView and not hidden
				if tile.Visibility[player] == NormalView {
					switch unit.Type {
					case TANK, ARTILLERY:
						if tile.Type != FOREST {
							su.Destroyed = true
							fmt.Printf("kill NormalView1: %s (%d) with %d health on %s by %v\n", string(su.Unit.Type), su.Unit.ID, su.Unit.Health, string(tile.Type), su.Unit.Activity)

						}
					case SOLDIER:
						if tile.Type != FOREST && tile.Type != GRASS && tile.Type != STRUCTURE {
							su.Destroyed = true
							fmt.Printf("kill NormalView2: %s (%d) with %d health on %s by %v\n", string(su.Unit.Type), su.Unit.ID, su.Unit.Health, string(tile.Type), su.Unit.Activity)

						}
					}
				}
			}

		}
	}

	// set lost flag (timeout)
	//---------------------------------------------------------------------------
	for _, v := range ShadowUnits {
		lostTime := float64(v.Unit.Speed) * 1.5
		if !v.Destroyed && v.LastSeen+uint64(lostTime) < w.Iteration {
			v.Lost = true
		}
	}

	// set destroyed flag (timeout)
	//---------------------------------------------------------------------------
	for _, v := range ShadowUnits {
		const destroyedTie = 650
		if !v.Destroyed && v.LastSeen+destroyedTie < w.Iteration {
			v.Destroyed = true
		}
	}

	// sort list
	//---------------------------------------------------------------------------
	{
		sortList := make([]*ShadowUnit, 0, len(ShadowUnits))
		for _, v := range ShadowUnits {
			sortList = append(sortList, v)
		}
		sort.Slice(sortList, func(i, j int) bool {
			a := sortList[i]
			b := sortList[j]
			if a.Unit.Player != b.Unit.Player {
				// sort by player
				return a.Unit.Player < b.Unit.Player
			} else {
				if a.Unit.Score() != b.Unit.Score() {
					// sort by score
					return a.Unit.Score() > b.Unit.Score()
				} else {
					// sort by ID
					return a.Unit.ID < b.Unit.ID
				}
			}
		})
		ShadowUnitsSortList = sortList
	}

	// generate phantom from last seen units (and set to world)
	//---------------------------------------------------------------------------
	for _, su := range ShadowUnits {

		// ---- not visible and alive and not my unit ------------------
		if !su.Visible && !su.Destroyed && su.Unit.Player != player {

			// unit don't move (and can evade)
			if !su.Unit.IsBusy() {
				endTime := su.Unit.Speed/2 + su.LastSeen
				if endTime > w.Iteration {
					// flag as phantom
					su.Unit.PHANTOM = w.Iteration
					// creat phantom in world
					tile := w.Tile(su.Unit.XCol(), su.Unit.YRow())
					if tile != nil && tile.Unit == nil {
						tile.Unit = su.Unit
					}
				}
				// fin
				continue // -> NEXT
			}

			// unit is fire -> and can't evade
			if su.Unit.IsFiring() {
				endTime := su.Unit.Activity.End + su.Unit.Speed/2
				if endTime > w.Iteration {
					// flag as phantom
					su.Unit.PHANTOM = w.Iteration
					// creat phantom in world
					tile := w.Tile(su.Unit.XCol(), su.Unit.YRow())
					if tile != nil && tile.Unit == nil {
						tile.Unit = su.Unit
					}
				}
				// fin
				continue // -> NEXT
			}

			// unit is running
			if su.Unit.IsMoving() {
				endTime := su.Unit.Activity.End + su.Unit.Speed/2
				if endTime > w.Iteration {
					// update position
					if su.Unit.Activity.SwitchPoint() > w.Iteration {
						su.Unit.tile.XCol = su.Unit.Activity.From[0]
						su.Unit.tile.YRow = su.Unit.Activity.From[1]
					} else {
						su.Unit.tile.XCol = su.Unit.Activity.To[0]
						su.Unit.tile.YRow = su.Unit.Activity.To[1]
					}
					// flag as phantom
					su.Unit.PHANTOM = w.Iteration
					// creat phantom in world
					tile := w.Tile(su.Unit.XCol(), su.Unit.YRow())
					if tile != nil && tile.Unit == nil {
						tile.Unit = su.Unit
					}
				}
			}
		}
	}
}
