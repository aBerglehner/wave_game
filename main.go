package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	SCREEN_WIDTH  = 800
	SCREEN_HEIGHT = 600
	FPS_TARGET    = 144
	RECT_SIZE     = 10
)

var (
	FPS_AVG      = make([]float64, 0, FPS_TARGET+10)
	TIME_CURRENT = time.Now()
)

// Game implements ebiten.Game interface.
type Game struct {
	x float32
	y float32
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// Write your game's logical update.
	movementController(g)

	logFpsAvg()
	return nil
}

func movementController(g *Game) {
	minDiffToCorner := float32(RECT_SIZE + 1)
	// up
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		if g.y > minDiffToCorner {
			g.y -= 1
		}
		fmt.Println("s key pressed")
	}
	// down
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		if g.y < SCREEN_HEIGHT-minDiffToCorner {
			g.y += 1
		}
		fmt.Println("d key pressed")
	}
	// left
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		if g.x > minDiffToCorner {
			g.x -= 1
		}
		fmt.Println("a key pressed")
	}
	// right
	if ebiten.IsKeyPressed(ebiten.KeyF) {
		if g.x < SCREEN_WIDTH-minDiffToCorner {
			g.x += 1
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
		g.x-4, g.y-4,
		RECT_SIZE,
		RECT_SIZE,
		color.White,
		true,
	)
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return SCREEN_WIDTH, SCREEN_HEIGHT
	// return 1024, 768
}

func main() {
	game := &Game{x: 10, y: 10}
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
