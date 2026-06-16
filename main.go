package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/alex/ebiten_tutorial/constants"
	"github.com/alex/ebiten_tutorial/enemy"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const (
	FpsTarget               = 144
	ScreenWidth             = 1400
	ScreenHeight            = 1050
	StatsBottomSize float32 = 35
)

var (
	fpsAvg      = make([]float64, 0, FpsTarget+10)
	timeCurrent = time.Now()
)

// player -> protagonist
const playerImageSize = 64 // pixels

var (
	playerCurrentFrame = 1
	playerSheet        *ebiten.Image
	playerFrames       = []image.Rectangle{
		image.Rect(0, 0, 64, 64),     // top-left
		image.Rect(64, 0, 128, 64),   // top-right
		image.Rect(0, 64, 64, 128),   // bottom-left
		image.Rect(64, 64, 128, 128), // bottom-right
	}
)

// enemies
var (
	// 0 indexed
	MONSTERS []*ebiten.Image
)

// can be looked up via -> lvl +1 indexed lvl 1 = index 0
var expLvlLookup [constants.LvlMax]int = [...]int{
	100,      // 1
	1000,     // 2
	10_000,   // 3
	50_000,   // 4
	100_0000, // 5
	500_0000, // 6
}

// Game implements ebiten.Game interface.
type Game struct {
	posX         float32
	posY         float32
	health       int
	dmg          float32
	healthAbsorb float32
	level        int
	exp          int
	expNeeded    int
	enemies      [10]enemy.Enemy
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// Write your game's logical update.
	movementController(g)

	// TODO:load check which monster is alive -> otherwise spawn a new one
	// TODO:load it to random postion that is valid
	// TODO:load only +1 -1 to own level monsters

	go logFpsAvg()
	return nil
}

func movementController(g *Game) {
	minDiffToCorner := float32(playerImageSize)
	// up
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		if g.posY > 0 {
			g.posY -= 1
			// OPTIMIZE:sure not needed every frame ->
			// idea either if check
			// or time based check and just set it every n frames to correct position so we get maybe also some transition
			playerCurrentFrame = 1
		}
		// fmt.Println("s key pressed")
	}
	// down
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		if g.posY < ScreenHeight-StatsBottomSize-minDiffToCorner {
			g.posY += 1
			playerCurrentFrame = 0
		}
		// fmt.Println("d key pressed")
	}
	// left
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		if g.posX > 0 {
			g.posX -= 1
			playerCurrentFrame = 2
		}
		// fmt.Println("a key pressed")
	}
	// right
	if ebiten.IsKeyPressed(ebiten.KeyF) {
		if g.posX <= ScreenWidth-minDiffToCorner {
			g.posX += 1
			playerCurrentFrame = 3
		}
		// fmt.Println("f key pressed")
	}
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	// Write your game's rendering.
	screen.Fill(color.RGBA{30, 100, 50, 1})

	// draw player
	sprite := playerSheet.SubImage(playerFrames[playerCurrentFrame]).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(g.posX), float64(g.posY))
	screen.DrawImage(sprite, op)

	// enemies
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Scale(0.35, 0.35)
	op2.GeoM.Translate(100, 100)
	screen.DrawImage(MONSTERS[1], op2)
	// TODO:draw all g.enemies

	statsBottom(g, screen)
}

func statsBottom(g *Game, screen *ebiten.Image) {
	// bottom line
	vector.StrokeLine(screen, 0, ScreenHeight-StatsBottomSize,
		ScreenWidth, ScreenHeight-StatsBottomSize,
		10, color.Black, true)

	// stats
	var bottomDistance int = 10
	text.Draw(
		screen,
		fmt.Sprintf("health: %d | dmg: %0.2f | health absorb: %d%% | lvl: %v | exp: %d/%d",
			g.health, g.dmg, int(g.healthAbsorb*100), g.level, g.exp, g.expNeeded),
		basicfont.Face7x13,
		10,
		ScreenHeight-bottomDistance,
		color.White,
	)
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
	// return 1024, 768
}

func init() {
	img, _, err := ebitenutil.NewImageFromFile("assets/protagonist.png")
	if err != nil {
		panic(err)
	}
	playerSheet = img

	// enemies
	monsters, err := loadMonsterImages("assets/monsters")
	if err != nil {
		panic(err)
	}
	MONSTERS = monsters
}

func main() {
	// TODO: create enemies
	game := &Game{posX: 10, posY: 10, health: 99, dmg: 1, healthAbsorb: 0.01, level: 1, exp: 0, expNeeded: expLvlLookup[1]}
	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetTPS(FpsTarget)
	ebiten.SetWindowSize(1400, 1050)
	// ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("simple mmo")

	// only for development
	ebiten.SetWindowPosition(2200, 0)

	ebiten.SetCursorMode(ebiten.CursorModeHidden)

	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func logFpsAvg() {
	fps := ebiten.ActualTPS()
	fpsAvg = append(fpsAvg, fps)

	if time.Since(timeCurrent) >= time.Second {
		var sum float64 = 0
		for _, v := range fpsAvg {
			sum += v
		}
		log.Printf("Avg FPS last second: %0.2f", (sum / float64(len(fpsAvg))))

		timeCurrent = time.Now()
		fpsAvg = make([]float64, 0, FpsTarget+10)
	}
}

func loadMonsterImages(dir string) ([]*ebiten.Image, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	var images []*ebiten.Image

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := filepath.Ext(file.Name())
		if ext != ".png" {
			continue
		}

		path := filepath.Join(dir, file.Name())

		img, _, err := ebitenutil.NewImageFromFile(path)
		if err != nil {
			return nil, err
		}

		images = append(images, img)
	}

	return images, nil
}
