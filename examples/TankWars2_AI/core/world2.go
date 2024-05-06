package core

//--------  AI2000  --------------------------------------------------------------------------------------------------//

// EnemyBases returns all enemy and neutral bases.
func (w *World) EnemyBases(player uint8) []*Tile {
	enemyBases := make([]*Tile, 0, 8)

	for _, base := range w.TileList(BASE) {
		if base != nil && base.Owner != player {
			enemyBases = append(enemyBases, base)
		}
	}

	return enemyBases
}

// AllBases returns all bases.
func (w *World) AllBases() []*Tile {
	return w.TileList(BASE)
}

// ActivitiesToTile returns all activities of units that have the specified position as their target.
func (w *World) ActivitiesToTile(x, y int, fireOnly bool) []*Activity {
	list := make([]*Activity, 0, 6)

	for _, tile := range w.Units(0) {
		if tile != nil && tile.Unit != nil && tile.Unit.Activity != nil {
			activity := tile.Unit.Activity
			toX := activity.To[0]
			toY := activity.To[1]

			if x == toX && y == toY {
				if !fireOnly || activity.Name == FIRE {
					list = append(list, activity)
				}
			}
		}
	}

	return list
}
