package core

import (
	"math/rand"
	"sort"
)

//--------  Activity  ------------------------------------------------------------------------------------------------//

// SwitchPoint calculates the mid-point of the movement activity (MOVE: switch to next tile)
func (a *Activity) SwitchPoint() uint64 {
	diff := a.End - a.Start
	switchPoint := a.Start + diff/2

	return switchPoint
}

//--------  Tile  ----------------------------------------------------------------------------------------------------//

// SetTile set a tile ref.
func (u *Unit) SetTile(tile *Tile) {
	u.tile = tile
}

// TileType returns the tile type (see TILES)
func (u *Unit) TileType() byte {
	if u == nil || u.tile == nil {
		println("err: TileType: tile is nil")
		return 0
	}
	return u.tile.Type
}

// XCol returns the column of the grid
func (u *Unit) XCol() int {
	if u == nil || u.tile == nil {
		println("err: XCol: tile is nil")
		return 0
	}
	return u.tile.XCol
}

// YRow returns the row of the grid
func (u *Unit) YRow() int {
	if u == nil || u.tile == nil {
		println("err: YRow: tile is nil")
		return 0
	}
	return u.tile.YRow
}

// Supply returns my supply level on this tile (1-15)
func (u *Unit) Supply() int {
	if u == nil || u.tile == nil {
		println("err: Supply: tile is nil")
		return 0
	}
	for _, v := range u.tile.Supply {
		return v // return first value
	}
	return 0
}

//--------  Unit  ----------------------------------------------------------------------------------------------------//

// Score returns the score points (value) of this unit
func (u *Unit) Score() int {
	var score = 0

	// add health
	score += u.Health

	// add unit type
	switch u.Type {
	case TANK:
		score += 20
	case SOLDIER:
		score += 10
	case ARTILLERY:
		score += 1
	}

	// add damage
	if u.Demoralized {
		score += 45 // demoralized !!
	} else {
		score += 78 // full damage
	}

	return score
}

// IsBusy checks if a unit activity exist
func (u *Unit) IsBusy() bool {
	return u.Activity != nil
}

// IsMoving checks if the unit is moving
func (u *Unit) IsMoving() bool {
	return u.Activity != nil && u.Activity.Name == MOVE
}

// IsFiring checks if the unit is firing
func (u *Unit) IsFiring() bool {
	return u.Activity != nil && u.Activity.Name == FIRE
}

// NearestBase return the nearest base.
func (u *Unit) NearestBase(world *World, player uint8, skipEnemy, skipOwn, skipOccupied bool) *Tile {
	if u.tile != nil && u.tile.Type == BASE {
		return u.tile // unit is in a base already
	}

	// find nearest base
	var bestBase *Tile
	var bestDistance = 999999999999

	for _, base := range world.AllBases() {
		if skipOccupied && base.Unit != nil {
			continue // ignore bases with units
		}
		if skipEnemy && base.Owner != player {
			continue // ignore enemy bases
		}
		if skipOwn && base.Owner == player {
			continue // ignore my own bases
		}

		path := FindPath(world, u.Type, u.tile, base)
		if bestDistance > len(path) {
			bestDistance = len(path)
			bestBase = base
		}
	}

	return bestBase
}

// NearestAllBase return the nearest base.
func (u *Unit) NearestAllBase(world *World, skipOccupied bool) *Tile {
	return u.NearestBase(world, 0, false, false, skipOccupied)
}

// IncomingFire returns a list of incoming shots that will hit the unit.
// This takes into account that the unit can move.
func (u *Unit) IncomingFire(world *World) []*Activity {
	var incomingList = make([]*Activity, 0)

	// collect all possible incomingList activities
	if u.IsMoving() {
		// is moving -> check two tiles
		from := world.ActivitiesToTile(u.Activity.From[0], u.Activity.From[1], true)
		to := world.ActivitiesToTile(u.Activity.To[0], u.Activity.To[1], true)
		incomingList = append(incomingList, from...)
		incomingList = append(incomingList, to...)

	} else {
		// unit stand still
		incomingList = world.ActivitiesToTile(u.XCol(), u.YRow(), true) // unit position

		// If the unit does not move, all incoming fire will hit the unit
		// ==>  FIN !
		return incomingList // -> FIN
	}

	// no activities
	if len(incomingList) == 0 {
		return incomingList // -> FIN
	}

	// Otherwise, all incoming projectiles must now be checked
	// to ensure that the unit does not dodge
	for i, incoming := range incomingList {
		// calculate iteration and position
		impactPos := incoming.To
		impactIter := incoming.End
		switchIter := u.Activity.SwitchPoint()

		// where is the unit ?
		unitPos := u.Activity.From // default is start position
		if impactIter > switchIter {
			unitPos = u.Activity.To // the impact is after the switch point
		}

		// check impact and unit position
		if impactPos[0] != unitPos[0] || impactPos[1] != unitPos[1] {
			incomingList[i] = nil // no hit
		}
	}

	// delete nil activities
	{
		tmp := make([]*Activity, 0, len(incomingList))
		for _, v := range incomingList {
			if v != nil {
				tmp = append(tmp, v)
			}
		}
		incomingList = tmp
	}

	// return
	return incomingList // -> FIN
}

func (u *Unit) CanMoveToTileType(to *Tile) bool {
	if u == nil || to == nil {
		return false
	}

	if u.Type != SOLDIER { // TANK and ARTILLERY
		if to.Type == MOUNTAIN || to.Type == STRUCTURE || to.Type == WATER {
			return false
		}
	}
	return true
}

type PossibleNeighbor struct {
	Neighbor *Tile
	Status   string
	Score    int // 0 blocked, 1-9 fire, 10 ok
}

// PossibleNeighborMoves returns the list of all neighboring tiles.
// Each tile is evaluated:
//   - A score of 0 means it is blocked.
//   - A score of 10 or more means it is good.
//   - A score between 1 and 9 means that it is passable but is under fire.
//
// The list is returned sorted by score (10, 10, 9, 0).
func (u *Unit) PossibleNeighborMoves(world *World) []*PossibleNeighbor {
	list := make([]*PossibleNeighbor, 0, 8)

	for _, n := range world.Neighbors(u.tile) {

		// can't move to this tile type
		if !u.CanMoveToTileType(n) {
			list = append(list, &PossibleNeighbor{
				Neighbor: n,
				Status:   "invalid",
			})
			continue
		}

		// other unit is on tile
		if n.Unit != nil {

			// stand still or does not move
			if n.Unit.Activity == nil || n.Unit.Activity.Name != MOVE {
				list = append(list, &PossibleNeighbor{
					Neighbor: n,
					Status:   "occupied",
				})
				continue
			}

			// unit move ...
			if n.Unit.Activity != nil && n.Unit.Activity.Name == MOVE {
				act := n.Unit.Activity
				if act.To[0] == n.XCol && act.To[1] == n.YRow {
					// ... to tile already
					list = append(list, &PossibleNeighbor{
						Neighbor: n,
						Status:   "occup.in",
					})
					continue
				} else {
					// ... not fast enough away
					if act.SwitchPoint() >= world.Iteration+u.Speed/2 {
						list = append(list, &PossibleNeighbor{
							Neighbor: n,
							Status:   "occup.out",
						})
						continue
					}
				}
			}
		}

		// incoming moves
		activitiesToTile := world.ActivitiesToTile(n.XCol, n.YRow, false)
		neighborScore := 0 // 0 good, >1 bad, >10 blocked
		for _, act := range activitiesToTile {
			if act.Name == MOVE {
				if act.SwitchPoint() <= world.Iteration+u.Speed/2 {
					neighborScore += 11 // blocked
				}
			}
			if act.Name == FIRE {
				if act.End >= world.Iteration+u.Speed/2 {
					neighborScore += 1 // fire
				}
			}
		}
		if neighborScore > 10 {
			list = append(list, &PossibleNeighbor{
				Neighbor: n,
				Status:   "move.in",
			})
			continue
		}
		if neighborScore > 0 {
			list = append(list, &PossibleNeighbor{
				Neighbor: n,
				Status:   "fire",
				Score:    10 - neighborScore,
			})
			continue
		}

		// FIN: neighbor is ok -> add to list
		list = append(list, &PossibleNeighbor{
			Neighbor: n,
			Status:   "ok",
			Score:    10,
		})
	}

	// add boni to score
	for _, pn := range list {
		if pn.Score <= 0 {
			continue // 0 is blocked
		}

		switch pn.Neighbor.Type {
		case BASE: // protection, vision, heath
			pn.Score += 2
		case STRUCTURE: // protection, hidden
			pn.Score += 1
		case GRASS: // open field, no boni (for tanks)
			if u.Type != SOLDIER {
				pn.Score += -1
			}
		case DIRT: // open field, no boni
			pn.Score += -1
		case WATER: // open field, no boni, no range
			pn.Score += -2
		}
	}

	// shuffle and sort list
	rand.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
	sort.Slice(list, func(i, j int) bool {
		return list[i].Score > list[j].Score
	})
	return list
}
