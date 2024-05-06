package ai2000

import (
	"TankWars2_AI/core"
	"math/rand"
	"sort"
)

/*
Shadow Units: Aktionen von zuletzt noch sichtbaren Einheiten werden auch im Schatten fortgeführt und damit bleibt die feindlochen Einheit eine kurze Zeit für die AI-Entscheidungen sichtbar.

EVADE: Einheiten weichen eingehendem Beschuss aus, sofern sie ihn sehen und die Zeit zum Ausweichen noch reicht.

CAPTURE: Benachbarte, unbesetzte, feindliche Basen werden erobert.

NO AMMO: Bei einem schweren Gefecht fallen Einheiten zurück, um sich neue Munition zu holen.

FIRE: Feindliche Einheiten in Schussreichweite werden priorisiert unter Beschuss genommen. Dabei werden auch Bewegungen mit einbezogen.

RETREAT: Demoralisierte Einheiten reparieren sich in einer Basis.

ATTACK: Feindliche Einheiten in Sichtweite werden konzentriert angegriffen. Verbündete Einheiten kommen zu Hilfe.

CAPTURE 2: Unbewachte feindliche Basen in mittlerer Nähe werden einfach eingenommen.

TARGET: Es werden strategische Ziele (Basen,  Engstellen, …)  auf der Karte definiert und erobert.
*/
func unitLoop() {
	targets := WORLD.EnemyBases(PLAYER)

UnitLoop:
	for _, tile := range WORLD.Units(PLAYER) {
		if tile == nil || tile.Unit == nil {
			continue // ERROR!!!
		}
		unit := tile.Unit

		// Skip units with existing commands, nil units, or units with ongoing activities.
		if tile.Unit.Activity != nil {
			unit.TEXT = MEMORYxUnitText[unit.ID]
			continue UnitLoop // NEXT UNIT
		}

		//___ TARGET: select random base ________________________________________________________________________

		// Check or set the unit's target memory.
		target := MEMORYxTarget[unit.ID] // get target
		{
			if target != nil {
				// update target (status changes)
				target = WORLD.Tile(target.XCol, target.YRow)
			}

			if target == nil || target.Owner == PLAYER {
				// No target found or target is captured by me.
				rand.Shuffle(len(targets), func(i, j int) {
					targets[i], targets[j] = targets[j], targets[i]
				})

				if len(targets) > 0 {
					// set new random target
					target = targets[0]

				} else {
					// no targets left -> go to next base
					// is there no free oder enemy base, the target is nil
					target = unit.NearestAllBase(WORLD, true) // set next base
				}
			}

			MEMORYxTarget[unit.ID] = target // set target
		}

		//___ EVADE: send move command and exit __________________________________________________________________

		// EVADE
		incomingFire := unit.IncomingFire(WORLD)
		if len(incomingFire) > 0 {

			// can the unit evade the fire?
			yes := 0
			no := 0
			for _, incoming := range incomingFire {
				if incoming.End > WORLD.Iteration+unit.Speed/2 {
					yes++
				} else {
					no++
				}
			}

			// evade
			if yes > 0 {
				list := unit.PossibleNeighborMoves(WORLD)
				if len(list) > 0 {
					_ = CLIENT.Move(tile.XCol, tile.YRow, list[0].Neighbor.XCol, list[0].Neighbor.YRow)
					unit.TEXT = "EVADE"
					MEMORYxUnitText[unit.ID] = unit.TEXT
					continue UnitLoop // NEXT UNIT
				}
			}
		}

		//___ CAPTURE (base next to me) __________________________________________________________________________________

		// tile next to me is an empty base
		for _, t := range WORLD.Neighbors(tile) {
			if t.Type == core.BASE && t.Unit == nil && t.Owner != PLAYER &&
				0 == len(WORLD.ActivitiesToTile(t.XCol, t.YRow, false)) {

				_ = CLIENT.Move(tile.XCol, tile.YRow, t.XCol, t.YRow)
				unit.TEXT = "CAPTURE"
				MEMORYxUnitText[unit.ID] = unit.TEXT
				continue UnitLoop // NEXT UNIT
			}
		}

		//___ NO AMMO __________________________________________________________________________________

		// no ammo? go to next base
		if unit.Ammunition < 1 {
			if nearest := unit.NearestBase(WORLD, PLAYER, true, false, false); nearest != nil {
				target = nearest
			} else if nearest = unit.NearestBase(WORLD, PLAYER, false, false, true); nearest != nil {
				target = nearest
			}

			if target != nil {
				_ = CLIENT.Move(tile.XCol, tile.YRow, target.XCol, target.YRow)
				unit.TEXT = "NO AMMO"
				MEMORYxUnitText[unit.ID] = unit.TEXT
				continue UnitLoop // NEXT UNIT
			}

		}

		//___ FIRE: send fire command and exit ___________________________________________________________

		// FIRE
		if unit.Ammunition >= 1 {

			// Check for enemies within firing range and initiate 'Fire' command if found.
			possible := make([]*core.Tile, 0, 12)
			for _, tmp := range WORLD.ExtNeighbors(tile, unit.FireRange) {
				for _, t := range tmp {

					// is enemy unit on tile
					if t != nil && t.Unit != nil && t.Unit.Player != PLAYER {

						// check is unit is moving and too fast to hit
						isTooFast := false
						if t.Unit.Activity != nil && t.Unit.Activity.Name == core.MOVE {
							act := t.Unit.Activity
							// moving from this tile to another
							if act.From[0] == t.XCol && act.From[1] == t.YRow {
								// too fast?
								if act.SwitchPoint() < WORLD.Iteration+unit.FireSpeed {
									// is the "TO" tile free?
									toTile := WORLD.Tile(act.To[0], act.To[1])
									if toTile != nil && toTile.Unit == nil {
										isTooFast = true
									}
								}
							}
						}

						if !isTooFast {
							possible = append(possible, t)
						}
					}

					// is enemy unit move to tile (PRIOR = nil)
					if t != nil && t.Unit == nil { // free tile
						for _, act := range WORLD.ActivitiesToTile(t.XCol, t.YRow, false) { // check activities
							if act.Name == core.MOVE && act.To[0] == t.XCol && act.To[1] == t.YRow { // move to tile
								if act.SwitchPoint() < WORLD.Iteration+unit.FireSpeed { // is fast enough
									// is enemy?
									isEnemy := WORLD.Tile(act.From[0], act.From[1])
									if isEnemy != nil && isEnemy.Unit != nil && isEnemy.Unit.Player != PLAYER {
										possible = append(possible, t) // incoming target
									}
								}
							}
						}
					}

					// add enemy bases for ARTILLERY
					if unit.Type == core.ARTILLERY {
						if t != nil && t.Type == core.BASE && t.Owner != PLAYER && t.Owner != 0 {
							// check own incoming unit
							for _, at := range WORLD.ActivitiesToTile(t.XCol, t.YRow, false) {
								if at.Name == core.MOVE {
									att := WORLD.Tile(at.From[0], at.From[1])
									if att != nil && att.Unit != nil && att.Unit.Player == PLAYER {
										continue // abort !!!!!
									}
								}
							}
							// else: add target
							possible = append(possible, t) // enemy base
						}
					}

				}
			}

			// sort targets (low armour & low health FIRST)
			sort.Slice(possible, func(i, j int) bool {
				a := possible[i].Unit
				b := possible[j].Unit
				if a == nil || b == nil {
					return false
				}
				scoreA := 10*(4-a.Armour) + (100 - a.Health)
				scoreB := 10*(4-b.Armour) + (100 - b.Health)
				return scoreA > scoreB
			})

			// select target
			if len(possible) > 0 {
				t := possible[0]

				_ = CLIENT.Fire(tile.XCol, tile.YRow, t.XCol, t.YRow)
				unit.TEXT = "FIRE"
				MEMORYxUnitText[unit.ID] = unit.TEXT
				continue UnitLoop // NEXT UNIT
			}
		}

		//___ RETREAT __________________________________________________________________________________

		if unit.Demoralized {
			if nearest := unit.NearestBase(WORLD, PLAYER, true, false, false); nearest != nil {
				target = nearest
			} else if nearest = unit.NearestBase(WORLD, PLAYER, false, false, true); nearest != nil {
				target = nearest
			}

			if target != nil {
				_ = CLIENT.Move(tile.XCol, tile.YRow, target.XCol, target.YRow)
				unit.TEXT = "RETREAT"
				MEMORYxUnitText[unit.ID] = unit.TEXT
				continue UnitLoop // NEXT UNIT
			}

		}

		//___ ATTACK (move) __________________________________________________________________________________

		// ATTACK
		if unit.Ammunition >= 1 {

			// Check for visible enemies within extended view range and initiate
			possible := make([]*core.Tile, 0, 12)
			for _, tmp := range WORLD.ExtNeighbors(tile, unit.View+2) {
				for _, t := range tmp {
					// all tiles in view range
					if t != nil && t.Unit != nil && t.Unit.Player != PLAYER {
						possible = append(possible, t)
					}
				}
			}

			// no options -> skip
			goToTileNearTarget := false
			if len(possible) > 0 { //________________________________________________________

				// sort
				sort.SliceStable(possible, func(p, q int) bool {
					a := possible[p].Supply[PLAYER]
					if a <= 0 {
						a = 99999
					}
					b := possible[q].Supply[PLAYER]
					if b <= 0 {
						b = 99999
					}
					return a < b
				})

				// DEBUG: move to target (possible[0]) with 0 fire range
				if unit.FireRange == 0 && possible[0] != nil && possible[0].Unit != nil {

					// is target (possible[0]) next to me?
					yesNextToMe := false
					for _, n := range WORLD.Neighbors(possible[0]) {
						if n.Unit != nil && n.Unit.ID == unit.ID {
							yesNextToMe = true
							break
						}
					}

					// go to another tile
					if yesNextToMe {
						for _, n := range WORLD.Neighbors(possible[0]) {
							if unit.CanMoveToTileType(n) && n.Unit == nil && n.Type != core.WATER && n.Type != core.BASE {
								// override target (possible[0]) !!!!!
								possible[0] = n
								goToTileNearTarget = true
								break
							}
						}
					}
				}

				// 'Move' command towards them, overriding the base target.
				for _, t := range possible {
					_ = CLIENT.Move(tile.XCol, tile.YRow, t.XCol, t.YRow)
					if goToTileNearTarget {
						unit.TEXT = "ATTACK (near)"
					} else {
						unit.TEXT = "ATTACK"
					}
					MEMORYxUnitText[unit.ID] = unit.TEXT
					continue UnitLoop // NEXT UNIT
				}
			} //____________________________________________________
		}

		//___ CAPTURE 2 (base 'near' to me) __________________________________________________________________________________

		// tile near to me is an empty base
		const capture2radius = 3
		for _, tt := range WORLD.ExtNeighbors(tile, capture2radius) {
			for _, t := range tt {
				// base is empty and not mine
				if t.Type == core.BASE && t.Unit == nil && t.Owner != PLAYER &&
					0 == len(WORLD.ActivitiesToTile(t.XCol, t.YRow, false)) {

					// check path distance
					path := core.FindPath(WORLD, unit.Type, tile, t)
					distance := len(path)

					if distance <= capture2radius+1 {
						_ = CLIENT.Move(tile.XCol, tile.YRow, t.XCol, t.YRow)
						unit.TEXT = "CAPTURE (near)"
						MEMORYxUnitText[unit.ID] = unit.TEXT
						continue UnitLoop // NEXT UNIT
					}
				}
			}
		}

		//___ TARGET (random Base) __________________________________________________________________________________

		// is target (base) next to me?
		if target != nil && target.Unit != nil {
			for _, n := range WORLD.Neighbors(target) {
				if n.Unit != nil && n.Unit.ID == unit.ID {
					// unit is next to target and the target is not empty
					if target.Unit.Player == PLAYER {
						println("target is next to me and not empty (my unit)")
					} else {
						println("target is next to me and not empty (enemy unit)")
					}
					MEMORYxTarget[unit.ID] = nil // clear target and select another one
				}
			}
		}

		// Move towards the chosen target (enemy base or AI-selected target).
		if target != nil {
			_ = CLIENT.Move(tile.XCol, tile.YRow, target.XCol, target.YRow)
			unit.TEXT = "TARGET"
			MEMORYxUnitText[unit.ID] = unit.TEXT
		}
	}
}
