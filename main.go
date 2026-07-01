package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"path/filepath"
	"time"

	"github.com/alex/ebiten_tutorial/constants"
	"github.com/alex/ebiten_tutorial/enemy"
	enemyI "github.com/alex/ebiten_tutorial/enemy"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const (
	FpsTarget                     = 144
	ScreenWidth                   = 1400
	ScreenWidthMaxSpawn           = float64(ScreenWidth - 50)
	ScreenHeight                  = 1050
	ScreenHeightMaxSpawn          = float64(MaxUsableScreenHeight - 50)
	MaxUsableScreenHeight         = ScreenHeight - int(StatsBottomSize) - StatsLineBottomSize
	StatsBottomSize       float64 = 40
	StatsLineBottomSize           = 10
)

var (
	fpsAvg                   = make([]float64, 0, FpsTarget+10)
	fpsTime                  = time.Now()
	movementTimePrev         = time.Now()
	movementSpeed    float64 = 100
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
	enemy_images     []*ebiten.Image
	enemyProjectiles []enemyI.EnemyProjectile
)

// 0 indexd -> can be looked up via -> lvl - 1 indexed lvl 1 = index 0
var expLvlLookup [constants.LvlMax]int = [...]int{
	100,         // 1
	1000,        // 2
	10_000,      // 3
	50_000,      // 4
	100_0000,    // 5
	500_0000,    // 6
	1_000_0000,  // 7
	2_000_0000,  // 8
	5_000_0000,  // 9
	10_000_0000, // 10
}

// Game implements ebiten.Game interface.
type Game struct {
	posX         float64
	posY         float64
	health       int
	dmg          float32
	healthAbsorb float32
	level        int
	exp          int
	expNeeded    int
	enemies      []enemyI.Enemy
	// this is for the animation of dmg taken
	damageTakenTime time.Time
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// Write your game's logical update.
	moveDistance := movementController(g)

	// TODO:load check which monster is alive -> otherwise spawn a new one
	// TODO:load it to random postion that is valid
	// TODO:load only +1 -1 to own level monsters

	attackRange2 := constants.AttackRange * constants.AttackRange
	playerPosX := g.posX
	playerPosY := g.posY
	for i := range g.enemies {
		enemy := &g.enemies[i]
		enemy.Patrol(ScreenWidthMaxSpawn, ScreenHeightMaxSpawn, moveDistance, FpsTarget)

		// TODO: let player attack
		attackFromEnemy(enemy, g, playerPosX, playerPosY, attackRange2)

	}

	// go logFpsAvg()
	return nil
}

func attackFromEnemy(enemy *enemy.Enemy, g *Game, playerPosX float64, playerPosY float64, attackRange2 float64) {
	posXDiff := enemy.PosX - playerPosX
	posYDiff := enemy.PosY - playerPosY
	if posXDiff*posXDiff+posYDiff*posYDiff <= attackRange2 {
		timeNow := time.Now()
		deltaLastAttackTime := timeNow.Sub(enemy.LastAttack)
		if enemy.AttackSpeed < deltaLastAttackTime.Milliseconds() {
			enemy.LastAttack = timeNow

			g.damageTakenTime = time.Now()
			g.health -= enemy.Dmg
		}
	}
}

func movementController(g *Game) (moveDistance float64) {
	cur := time.Now()
	deltaTime := cur.Sub(movementTimePrev)
	movementTimePrev = cur
	moveDistance = movementSpeed * float64(deltaTime.Seconds())

	minDiffToCorner := float64(playerImageSize)
	// up
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		if g.posY > 0 {
			g.posY -= moveDistance
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
			g.posY += moveDistance
			playerCurrentFrame = 0
		}
		// fmt.Println("d key pressed")
	}
	// left
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		if g.posX > 0 {
			g.posX -= moveDistance
			playerCurrentFrame = 2
		}
		// fmt.Println("a key pressed")
	}
	// right
	if ebiten.IsKeyPressed(ebiten.KeyF) {
		if g.posX <= ScreenWidth-minDiffToCorner {
			g.posX += moveDistance
			playerCurrentFrame = 3
		}
		// fmt.Println("f key pressed")
	}
	return moveDistance
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	// the stuff that you draw last is on top
	drawBackground(screen)

	drawEnemies(g, screen)
	// we draw player after enemies so the image is on top
	drawPlayer(g, screen)

	statsBottom(g, screen)

	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
}

func drawBackground(screen *ebiten.Image) {
	// background color
	screen.Fill(color.RGBA{30, 100, 50, 1})

	// grid lines -> left to right
	var gridSpace float32 = 100
	var strokeWidth float32 = 0.6
	var x0 float32 = 0
	var x1 float32 = ScreenWidth
	var y0 float32 = gridSpace
	var y1 float32 = gridSpace
	var i float32
	for i = gridSpace; i <= float32(MaxUsableScreenHeight); i += gridSpace {
		vector.StrokeLine(screen, x0, y0, x1, y1, strokeWidth, color.RGBA{255, 255, 255, 1}, true)
		y0 += gridSpace
		y1 += gridSpace
	}

	// grid lines -> top to bottom
	x0 = gridSpace
	x1 = gridSpace
	y0 = 0
	y1 = float32(MaxUsableScreenHeight)
	for i = gridSpace; i < ScreenWidth; i += gridSpace {
		vector.StrokeLine(screen, x0, y0, x1, y1, strokeWidth, color.RGBA{255, 255, 255, 1}, true)
		x0 += gridSpace
		x1 += gridSpace
	}
}

func drawPlayer(g *Game, screen *ebiten.Image) {
	sprite := playerSheet.SubImage(playerFrames[playerCurrentFrame]).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(g.posX, g.posY)
	screen.DrawImage(sprite, op)

	drawPlayerDmgTaken(g, screen)
}

// TODO: idea draw the same for enemies???
func drawPlayerDmgTaken(g *Game, screen *ebiten.Image) {
	var r float32 = 21
	var strokeWidth float32 = 0

	var cx float32 = 0
	var cy float32 = 0
	if playerCurrentFrame == 3 { // right
		cx = float32(g.posX) + 25
		cy = float32(g.posY) + 25
	} else if playerCurrentFrame == 2 { // left
		cx = float32(g.posX) + 35
		cy = float32(g.posY) + 25
	} else if playerCurrentFrame == 1 { // up
		cx = float32(g.posX) + 26
		cy = float32(g.posY) + 40
	} else if playerCurrentFrame == 0 { // down
		cx = float32(g.posX) + 35
		cy = float32(g.posY) + 40
	}

	timeNow := time.Now()
	deltaDamageTakenTime := timeNow.Sub(g.damageTakenTime).Milliseconds()
	if deltaDamageTakenTime < 45 {
		strokeWidth = 1
	} else if deltaDamageTakenTime < 90 {
		strokeWidth = 3
		r += strokeWidth
	} else if deltaDamageTakenTime < 135 {
		strokeWidth = 6
		r += strokeWidth
	} else if deltaDamageTakenTime < 180 {
		strokeWidth = 9
		r += strokeWidth
	}

	if deltaDamageTakenTime < 180 {
		vector.StrokeCircle(screen, cx, cy, r, strokeWidth, color.RGBA{150, 0, 0, 150}, true)
	}
}

func drawEnemies(g *Game, screen *ebiten.Image) {
	for i := range g.enemies {
		enemy := g.enemies[i]
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(0.35, 0.35)
		op.GeoM.Translate(enemy.PosX, enemy.PosY)
		screen.DrawImage(enemy_images[enemy.Lvl], op)

		// health bar
		// should be a percentage of 40. 40 == 100%
		enemyMaxHealth := enemyI.EnemyHealthLookup[enemy.Lvl-1]
		enemyHealthPercentage := enemy.Health * 100 / enemyMaxHealth
		var lifeVectorMaxWidth float32 = 40
		var life float32 = lifeVectorMaxWidth * float32(enemyHealthPercentage) / 100

		var lifeHeight float32 = 6
		var healthBarPosY float32 = float32(enemy.PosY) - (lifeHeight + 1)
		var healthBarPosX float32 = float32(enemy.PosX)
		vector.FillRect(screen, healthBarPosX, healthBarPosY, life, lifeHeight, color.RGBA{150, 0, 0, 150}, true)
	}
}

func statsBottom(g *Game, screen *ebiten.Image) {
	// bottom line
	vector.StrokeLine(screen, 0, float32(ScreenHeight-StatsBottomSize),
		ScreenWidth, float32(ScreenHeight-StatsBottomSize),
		StatsLineBottomSize, color.Black, true)

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
	img, _, err := ebitenutil.NewImageFromFile(filepath.Join("assets", "protagonist.png"))
	if err != nil {
		panic(err)
	}
	playerSheet = img

	// enemies
	monsters, err := enemy.LoadEnemyImages(filepath.Join("assets", "monsters"))
	if err != nil {
		panic(err)
	}
	enemy_images = monsters

	// create the init pool of enemyProjectiles
	enemyProjectiles = enemyI.EnemyProjectilesInit(enemyI.EnemiesCount)
}

func gameInit() *Game {
	enemies := enemy.CreateInit(ScreenWidthMaxSpawn, ScreenHeightMaxSpawn)
	return &Game{posX: 10, posY: 10, health: 100, dmg: 1, healthAbsorb: 0.01, level: 1, exp: 0, expNeeded: expLvlLookup[1], enemies: enemies}
}

func main() {
	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetTPS(FpsTarget)
	ebiten.SetWindowSize(1400, 1050)
	// ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("wave game")

	// only for development
	ebiten.SetWindowPosition(2200, 0)

	ebiten.SetCursorMode(ebiten.CursorModeHidden)

	game := gameInit()
	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func logFpsAvg() {
	fps := ebiten.ActualTPS()
	fpsAvg = append(fpsAvg, fps)

	if time.Since(fpsTime) >= time.Second {
		var sum float64 = 0
		for _, v := range fpsAvg {
			sum += v
		}
		log.Printf("Avg FPS last second: %0.2f", (sum / float64(len(fpsAvg))))

		fpsTime = time.Now()
		fpsAvg = make([]float64, 0, FpsTarget+10)
	}
}
