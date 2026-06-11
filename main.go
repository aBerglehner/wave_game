package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const (
	FPS_TARGET                = 144
	SCREEN_WIDTH              = 800
	SCREEN_HEIGHT             = 600
	PLAYER_RECT_SIZE          = 10
	STATS_BOTTOM_SIZE float32 = 35
	LVL_MAX           int     = 6
)

var (
	FPS_AVG      = make([]float64, 0, FPS_TARGET+10)
	TIME_CURRENT = time.Now()
)

// can be looked up via -> lvl indexed lvl 1 = index 1
var EXP_LVL_LOOKUP [LVL_MAX]int = [...]int{
	0,
	100,
	1000,
	10_000,
	50_000,
	100_0000,
}

// can all be looked up via -> enemies lvl-> lvl 1 = index 1
var (
	enemy_dmg_lookup    [LVL_MAX]int = [...]int{0, 1, 0, 0, 0, 0}
	enemy_health_lookup [LVL_MAX]int = [...]int{0, 1, 0, 0, 0, 0}
	enemy_exp_lookup    [LVL_MAX]int = [...]int{0, 1, 0, 0, 0, 0}
	// slower on lower lvl
	enemy_attack_speed_lookup [LVL_MAX]int = [...]int{0, 5_000, 4_000, 2_000, 1_000, 500}
)

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
	enemies      [10]Enemy
}

type Enemy struct {
	posX   float32
	posY   float32
	lvl    int
	dmg    int
	health int
	exp    int
	// ms
	attackSpeed int
	lastAttack  time.Time
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// Write your game's logical update.
	movementController(g)

	go logFpsAvg()
	return nil
}

func movementController(g *Game) {
	minDiffToCorner := float32(PLAYER_RECT_SIZE + 1)
	// up
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		if g.posY > minDiffToCorner {
			g.posY -= 1
		}
		fmt.Println("s key pressed")
	}
	// down
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		if g.posY < SCREEN_HEIGHT-STATS_BOTTOM_SIZE-minDiffToCorner {
			g.posY += 1
		}
		fmt.Println("d key pressed")
	}
	// left
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		if g.posX > minDiffToCorner {
			g.posX -= 1
		}
		fmt.Println("a key pressed")
	}
	// right
	if ebiten.IsKeyPressed(ebiten.KeyF) {
		if g.posX < SCREEN_WIDTH-minDiffToCorner {
			g.posX += 1
		}
		fmt.Println("f key pressed")
	}
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	// Write your game's rendering.
	screen.Fill(color.RGBA{30, 100, 50, 1})

	vector.FillRect(
		screen,
		g.posX-4, g.posY-4,
		PLAYER_RECT_SIZE,
		PLAYER_RECT_SIZE,
		color.White,
		true,
	)

	statsBottom(g, screen)
}

func statsBottom(g *Game, screen *ebiten.Image) {
	// bottom line
	vector.StrokeLine(screen, 0, SCREEN_HEIGHT-STATS_BOTTOM_SIZE,
		SCREEN_WIDTH, SCREEN_HEIGHT-STATS_BOTTOM_SIZE,
		10, color.Black, true)

	// stats
	var bottomDistance int = 10
	text.Draw(
		screen,
		fmt.Sprintf("health: %d | dmg: %0.2f | health absorb: %d%% | lvl: %v | exp: %d/%d",
			g.health, g.dmg, int(g.healthAbsorb*100), g.level, g.exp, g.expNeeded),
		basicfont.Face7x13,
		10,
		SCREEN_HEIGHT-bottomDistance,
		color.White,
	)
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return SCREEN_WIDTH, SCREEN_HEIGHT
	// return 1024, 768
}

func main() {
	game := &Game{posX: 10, posY: 10, health: 99, dmg: 1, healthAbsorb: 0.01, level: 1, exp: 0, expNeeded: EXP_LVL_LOOKUP[1]}
	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetTPS(FPS_TARGET)
	ebiten.SetWindowSize(1024, 768)
	// ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Your game's title")

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
	FPS_AVG = append(FPS_AVG, fps)

	if time.Since(TIME_CURRENT) >= time.Second {
		var sum float64 = 0
		for _, v := range FPS_AVG {
			sum += v
		}
		log.Printf("Avg FPS last second: %0.2f", (sum / float64(len(FPS_AVG))))

		TIME_CURRENT = time.Now()
		FPS_AVG = make([]float64, 0, FPS_TARGET+10)
	}
}
