package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenW = 320
	screenH = 240
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
	cursorX, cursorY := ebiten.CursorPosition()
	g.x = float32(cursorX)
	g.y = float32(cursorY)
	fmt.Printf("cursorX: %v\n", cursorX)
	fmt.Printf("cursorY: %v\n", cursorY)

	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	// Write your game's rendering.
	screen.Fill(color.RGBA{30, 100, 50, 1})

	vector.FillRect(
		screen,
		g.x-4, g.y-4,
		10,
		10,
		color.White,
		true,
	)
	fps := ebiten.ActualFPS()
	log.Printf("FPS: %0.2f", fps)
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return screenW, screenH
	// return 1024, 768
}

func main() {
	game := &Game{x: 10, y: 10}
	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetTPS(60)
	ebiten.SetWindowSize(1024, 768)
	// ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Your game's title")

	// only for development
	ebiten.SetWindowPosition(2200, 0)

	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
