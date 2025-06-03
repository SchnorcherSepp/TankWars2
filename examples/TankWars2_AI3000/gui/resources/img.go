package resources

import (
	"embed"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	_ "image/jpeg" // needed for ebitenutil.NewImageFromReader()
	_ "image/png"  // needed for ebitenutil.NewImageFromReader()
	"log"
)

// Imgs is the global variable that holds all image resources
var Imgs *ImgResources

// ImgResources is a collection of all images
type ImgResources struct {
	Danger    *ebiten.Image // 17 x 17
	EmptyTile *ebiten.Image // 120 x 140
	Error     *ebiten.Image // 64 x 64
	ErrorTile *ebiten.Image // 120 x 140
	Explosion *ebiten.Image // 64 x 64
	Eye       *ebiten.Image // 26 x 15
	Logo      *ebiten.Image // 64 x 64

	Artillery *ebiten.Image // 50 x 50
	Soldier   *ebiten.Image // 50 x 50
	Tank      *ebiten.Image // 50 x 50

	TilesBase      [2]*ebiten.Image  // 120 x 140
	TilesDirt      [1]*ebiten.Image  // 120 x 140
	TilesForest    [4]*ebiten.Image  // 120 x 140
	TilesGrass     [1]*ebiten.Image  // 120 x 140
	TilesHill      [1]*ebiten.Image  // 120 x 140
	TilesHole      [3]*ebiten.Image  // 120 x 140
	TilesMountain  [3]*ebiten.Image  // 120 x 140
	TilesStructure [13]*ebiten.Image // 120 x 140
	TilesWater     [1]*ebiten.Image  // 120 x 140
}

func init() {
	Imgs = &ImgResources{
		Danger:    loadGameImg("img/danger.png"),
		EmptyTile: loadGameImg("img/empty-tile.png"),
		Error:     loadGameImg("img/error.png"),
		ErrorTile: loadGameImg("img/error-tile.png"),
		Explosion: loadGameImg("img/explosion.png"),
		Eye:       loadGameImg("img/eye.png"),
		Logo:      loadGameImg("img/logo.png"),

		Artillery: loadGameImg("img/artillery.png"),
		Soldier:   loadGameImg("img/soldier.png"),
		Tank:      loadGameImg("img/tank.png"),
	}

	for i := 0; i < len(Imgs.TilesBase); i++ {
		Imgs.TilesBase[i] = loadGameImg(fmt.Sprintf("img/base/%d.png", i+1))
	}
	for i := 0; i < len(Imgs.TilesDirt); i++ {
		Imgs.TilesDirt[i] = loadGameImg(fmt.Sprintf("img/dirt/%d.png", i+1))
	}
	for i := 0; i < len(Imgs.TilesForest); i++ {
		Imgs.TilesForest[i] = loadGameImg(fmt.Sprintf("img/forest/%d.png", i+1))
	}
	for i := 0; i < len(Imgs.TilesGrass); i++ {
		Imgs.TilesGrass[i] = loadGameImg(fmt.Sprintf("img/grass/%d.png", i+1))
	}
	for i := 0; i < len(Imgs.TilesHill); i++ {
		Imgs.TilesHill[i] = loadGameImg(fmt.Sprintf("img/hill/%d.png", i+1))
	}
	for i := 0; i < len(Imgs.TilesHole); i++ {
		Imgs.TilesHole[i] = loadGameImg(fmt.Sprintf("img/hole/%d.png", i+1))
	}
	for i := 0; i < len(Imgs.TilesMountain); i++ {
		Imgs.TilesMountain[i] = loadGameImg(fmt.Sprintf("img/mountain/%d.png", i+1))
	}
	for i := 0; i < len(Imgs.TilesStructure); i++ {
		Imgs.TilesStructure[i] = loadGameImg(fmt.Sprintf("img/structure/%d.png", i+1))
	}
	for i := 0; i < len(Imgs.TilesWater); i++ {
		Imgs.TilesWater[i] = loadGameImg(fmt.Sprintf("img/water/%d.png", i+1))
	}
}

//go:embed img
var gFS embed.FS

func loadGameImg(name string) *ebiten.Image {
	// open reader
	r, err := gFS.Open(name)
	if err != nil {
		log.Fatalf("err: loadGameImg: %v\n", err)
	}
	// get image
	eim, _, err := ebitenutil.NewImageFromReader(r)
	if err != nil {
		log.Fatalf("err: loadGameImg: %v\n", err)
	}
	// return
	return eim
}
