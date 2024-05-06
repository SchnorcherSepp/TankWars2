package gui

// Update updates an img by one tick. The given argument represents a screen image.
//
// Update updates only the img logic and Draw draws the screen.
//
// In the first frame, it is ensured that Update is called at least once before Draw. You can use Update
// to initialize the img state.
//
// After the first frame, Update might not be called or might be called once
// or more for one frame. The frequency is determined by the current TPS (tick-per-second).
func (g *Game) Update() error {

	if g.remote != nil {
		// use remote world status and override local world
		g.world = g.remote.Status()

	} else {
		// call local world update
		if g.world != nil {
			g.world.Update()
		}
	}
	return nil
}
