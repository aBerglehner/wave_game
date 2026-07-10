package main

import (
	"bytes"
	"embed"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"sync"
	"time"

	"github.com/alex/ebiten_tutorial/constants"
	enemyI "github.com/alex/ebiten_tutorial/enemy"
	"github.com/alex/ebiten_tutorial/projectile"
	"github.com/alex/ebiten_tutorial/utils"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

//go:embed assets/font.otf
var fontBytes []byte
var fontFace *text.GoTextFace

// this will load it into the binary(no assets folder needed than)
//
//go:embed assets/protagonist.png
var playerPNG []byte

//go:embed assets/monsters/*
var monsterVirtualFileSystem embed.FS

const (
	FpsTarget                     = 144
	ScreenWidth                   = 1400
	ScreenWidthMaxSpawn           = float64(ScreenWidth - 50)
	ScreenHeight                  = 1050
	ScreenHeightMaxSpawn          = float64(MaxUsableScreenHeight - 50)
	MaxUsableScreenHeight         = ScreenHeight - int(StatsBottomSize)
	StatsBottomSize       float64 = 50
)

var (
	fpsAvg                   = make([]float64, 0, FpsTarget+10)
	fpsTime                  = time.Now()
	movementTimePrev         = time.Now()
	movementSpeed    float64 = 100
)

// player
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
	playerProjectiles []projectile.Projectile
)

// enemies
var (
	// 0 indexed
	enemy_images     []*ebiten.Image
	enemyProjectiles []projectile.Projectile
)

// 0 indexd -> can be looked up via -> lvl - 1 indexed lvl 1 = index 0
var playerExpLvlLookup [constants.LvlMax]int = [...]int{
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

// ms
var playerAttackSpeedLookup [constants.LvlMax]int = [...]int{
	800, // 1
	700, // 2
	600, // 3
	500, // 4
	400, // 5
	300, // 6
	200, // 7
	100, // 8
	80,  // 9
	60,  // 10
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
	// ms
	attackSpeed int
	lastAttack  time.Time
	// this is for the animation of dmg taken
	damageTakenTime time.Time
	enemies         []enemyI.Enemy
}

func initGame() *Game {
	enemies := enemyI.EnemyCreateInit(ScreenWidthMaxSpawn, ScreenHeightMaxSpawn)
	return &Game{
		posX:         10,
		posY:         10,
		health:       100,
		dmg:          1,
		healthAbsorb: 0.01,
		level:        1,
		exp:          0,
		expNeeded:    playerExpLvlLookup[0],
		attackSpeed:  playerAttackSpeedLookup[0],
		enemies:      enemies,
	}
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// Write your game's logical update.
	// start := time.Now()
	moveDistance := movementController(g)

	// TODO:load check which monster is alive -> otherwise spawn a new one
	// TODO:load it to random postion that is valid
	// TODO:load only +1 -1 to own level monsters

	playerPosX := g.posX
	playerPosY := g.posY

	enemiesThatWantToAttackCh := make(chan *enemyI.Enemy)
	var wg sync.WaitGroup

	workers := 6 // runtime.GOMAXPROCS(0)
	tasks := utils.SplitTasks(g.enemies, workers)

	for _, enemies := range tasks {
		wg.Add(1)
		go func(enemies []enemyI.Enemy) {
			defer wg.Done()
			updateEnemies(enemies, moveDistance, playerPosX, playerPosY, enemiesThatWantToAttackCh)
		}(enemies)
	}

	go func() {
		wg.Wait()
		close(enemiesThatWantToAttackCh)
	}()

	for c := range enemiesThatWantToAttackCh {
		createEnemyProjectile(c, g)
	}

	updateEnemyProjectiles()
	handleEnemyProjectilesCollisions(g)

	// go logFpsAvg()
	// log.Printf("update took %v", time.Since(start))
	return nil
}

// createPlayerProjectile -> will create new projectiles every n time
func createPlayerProjectile(enemy *enemyI.Enemy, g *Game) {
	playerX := g.posX + playerImageSize/2
	playerY := g.posY + playerImageSize/2
	dx := playerX - enemy.PosX
	dy := playerY - enemy.PosY
	length := math.Sqrt(dx*dx + dy*dy)
	dir := projectile.Pos{X: dx / length, Y: dy / length}
	velocity := projectile.Pos{X: dir.X * enemy.ProjectileSpeed, Y: dir.Y * enemy.ProjectileSpeed}
	// find not alive enemyProjectiles to use
	doublePoolNeeded := true
	for i := range enemyProjectiles {
		if !enemyProjectiles[i].Alive {
			enemyProjectiles[i] = projectile.NewProjectile(projectile.Pos{X: enemy.PosX, Y: enemy.PosY}, velocity, enemy.Dmg)
			doublePoolNeeded = false
			break
		}
	}

	// yes this will skip one attack -> but this(bug) is ok to have 1 attack out of a shit ton not happening
	if doublePoolNeeded {
		fmt.Printf("double pool needed current len(enemyProjectiles): %v\n", len(enemyProjectiles))
		enemyProjectiles = append(enemyProjectiles, projectile.ProjectilesInit(len(enemyProjectiles))...)
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

func updateEnemies(enemies []enemyI.Enemy, moveDistance float64, playerPosX float64, playerPosY float64, enemiesThatWantToAttackCh chan<- *enemyI.Enemy) {
	for i := range enemies {
		enemy := &enemies[i]
		enemy.Patrol(ScreenWidthMaxSpawn, ScreenHeightMaxSpawn, moveDistance, FpsTarget)

		// TODO: let player attack
		attackFromEnemy(enemy, playerPosX, playerPosY, enemiesThatWantToAttackCh)

	}
}

func attackFromEnemy(enemy *enemyI.Enemy, playerPosX float64, playerPosY float64, enemiesThatWantToAttackCh chan<- *enemyI.Enemy) {
	attackRange2 := enemy.AttackRange * enemy.AttackRange
	posXDiff := enemy.PosX - playerPosX
	posYDiff := enemy.PosY - playerPosY
	if posXDiff*posXDiff+posYDiff*posYDiff <= attackRange2 {
		timeNow := time.Now()
		deltaLastAttackTime := timeNow.Sub(enemy.LastAttack)
		if enemy.AttackSpeed < deltaLastAttackTime.Milliseconds() {
			enemy.LastAttack = timeNow

			// just send them and handle them afterwards all togehter
			enemiesThatWantToAttackCh <- enemy
		}
	}
}

// createEnemyProjectile: a projectile in the direction to the player is created. It is added to the global var []enemyProjectiles
func createEnemyProjectile(enemy *enemyI.Enemy, g *Game) {
	playerX := g.posX + playerImageSize/2
	playerY := g.posY + playerImageSize/2
	dx := playerX - enemy.PosX
	dy := playerY - enemy.PosY
	length := math.Sqrt(dx*dx + dy*dy)
	dir := projectile.Pos{X: dx / length, Y: dy / length}
	velocity := projectile.Pos{X: dir.X * enemy.ProjectileSpeed, Y: dir.Y * enemy.ProjectileSpeed}
	// find not alive enemyProjectiles to use
	doublePoolNeeded := true
	for i := range enemyProjectiles {
		if !enemyProjectiles[i].Alive {
			enemyProjectiles[i] = projectile.NewProjectile(projectile.Pos{X: enemy.PosX, Y: enemy.PosY}, velocity, enemy.Dmg)
			doublePoolNeeded = false
			break
		}
	}

	// yes this will skip one attack -> but this(bug) is ok to have 1 attack out of a shit ton not happening
	if doublePoolNeeded {
		fmt.Printf("double pool needed current len(enemyProjectiles): %v\n", len(enemyProjectiles))
		enemyProjectiles = append(enemyProjectiles, projectile.ProjectilesInit(len(enemyProjectiles))...)
	}
}

func updateEnemyProjectiles() {
	// count is never 0 as len(enemyProjectiles) == min enemies on display
	count := len(enemyProjectiles)
	workers := 6 // runtime.GOMAXPROCS(0)
	// workForEach will never be 0 -> as len(enemyProjectiles) -> never shrinks only alive = false
	workForEach := count / workers

	var wg sync.WaitGroup
	i := 0
	for i = 0; i < count; i += workForEach {
		// this handles the left overs
		max := min(i+workForEach, count)

		wg.Add(1)
		go func(start int, end int) {
			defer wg.Done()
			updatePartOfEnemyProjectiles(start, end)
		}(i, max)
	}
	wg.Wait()
}

func updatePartOfEnemyProjectiles(start int, end int) {
	for i := start; i < end; i += 1 {
		if enemyProjectiles[i].Alive {
			pos := enemyProjectiles[i].CurPos
			vel := enemyProjectiles[i].Velocity
			newPosX := pos.X + (vel.X / FpsTarget)
			newPosY := pos.Y + (vel.Y / FpsTarget)
			enemyProjectiles[i].OldPos = projectile.Pos{X: pos.X, Y: pos.Y}
			enemyProjectiles[i].CurPos = projectile.Pos{X: newPosX, Y: newPosY}

			if newPosX < 0 || newPosX > ScreenWidth {
				enemyProjectiles[i].Alive = false
			}
			if newPosY < 0 || newPosY > float64(MaxUsableScreenHeight) {
				enemyProjectiles[i].Alive = false
			}
		}
	}
}

func handleEnemyProjectilesCollisions(g *Game) {
	playerXCenter := g.posX + playerImageSize/2
	playerYCenter := g.posY + playerImageSize/2
	playerSizeRadius := float32(math.Sqrt(playerImageSize))

	// count is never 0 as len(enemyProjectiles) == min enemies on display
	count := len(enemyProjectiles)
	workers := 6 // runtime.GOMAXPROCS(0)
	// workForEach will never be 0 -> as len(enemyProjectiles) -> never shrinks only alive = false
	workForEach := count / workers

	var wg sync.WaitGroup
	dmgTakenProjectilesCh := make(chan *projectile.Projectile)
	i := 0
	for i = 0; i < count; i += workForEach {
		// this handles the left overs
		max := min(i+workForEach, count)

		wg.Add(1)
		go func(start int, end int) {
			defer wg.Done()

			for i := start; i < end; i += 1 {
				if enemyProjectiles[i].Alive {
					project := enemyProjectiles[i]
					hitBoxRadius := float64(project.Radius + playerSizeRadius)
					if projectileCollision(project.OldPos, project.CurPos, projectile.Pos{X: playerXCenter, Y: playerYCenter}, hitBoxRadius) {
						dmgTakenProjectilesCh <- &enemyProjectiles[i]
					}
				}
			}
		}(i, max)
	}
	go func() {
		wg.Wait()
		close(dmgTakenProjectilesCh)
	}()

	for v := range dmgTakenProjectilesCh {
		g.damageTakenTime = time.Now()
		g.health -= v.Dmg
		// otherwise it will make dmg every tick
		v.Alive = false
	}
}

// TODO: hitbox head ball comming from the left does not work
func projectileCollision(oldPos projectile.Pos, newPos projectile.Pos, testPos projectile.Pos, hitBoxRadius float64) bool {
	dx := newPos.X - oldPos.X
	dy := newPos.Y - oldPos.Y

	// position along the projectile's path
	// 0 = start
	// 0.5 = halfway
	// 1 = end
	// <0 = before the start
	// >1 = after the end
	t := ((testPos.X-oldPos.X)*dx + (testPos.Y-oldPos.Y)*dy) / (dx*dx + dy*dy)

	// Clamp to segment
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	closestX := oldPos.X + t*dx
	closestY := oldPos.Y + t*dy

	distX := testPos.X - closestX
	distY := testPos.Y - closestY

	return distX*distX+distY*distY <= hitBoxRadius*hitBoxRadius
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
// The stuff that you draw last is on top
func (g *Game) Draw(screen *ebiten.Image) {
	// start := time.Now()
	drawBackground(screen)

	drawEnemies(g, screen)
	drawEnemyProjectiles(screen)
	// we draw player after enemies so the image is on top
	drawPlayer(g, screen)

	statsBottom(g, screen)

	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
	// log.Printf("draw took %v", time.Since(start))
}

func drawEnemyProjectiles(screen *ebiten.Image) {
	for i := range enemyProjectiles {
		if enemyProjectiles[i].Alive {
			var cx float32 = float32(enemyProjectiles[i].CurPos.X)
			var cy float32 = float32(enemyProjectiles[i].CurPos.Y)
			var r float32 = enemyProjectiles[i].Radius
			vector.FillCircle(screen, cx, cy, r, color.RGBA{150, 0, 0, 150}, false)
		}
	}
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
		vector.FillRect(screen, healthBarPosX, healthBarPosY, life, lifeHeight, color.RGBA{150, 0, 0, 150}, false)
	}
}

type StatsInfo struct {
	t          string
	sizeNeeded int
}

func statsBottom(g *Game, screen *ebiten.Image) {
	// TODO: health show max health also (currently not available)

	var bottomDistance int = 10
	yPos := float64(ScreenHeight - int(fontFace.Size) - bottomDistance)

	var statsText []StatsInfo
	statsText = append(statsText, StatsInfo{fmt.Sprintf("hp: %d", g.health), 200})
	statsText = append(statsText, StatsInfo{fmt.Sprintf("dmg: %0.2f", g.dmg), 250})
	statsText = append(statsText, StatsInfo{fmt.Sprintf("hp absorb: %d%%", int(g.healthAbsorb*100)), 200})
	statsText = append(statsText, StatsInfo{fmt.Sprintf("lvl: %v", g.level), 200})
	statsText = append(statsText, StatsInfo{fmt.Sprintf("exp: %d%%", g.exp*100/g.expNeeded), 100})

	var curPosX float64 = 10
	for i := range statsText {
		op := &text.DrawOptions{}
		op.GeoM.Translate(curPosX, yPos)
		curText := statsText[i].t

		text.Draw(
			screen,
			curText,
			// TODO: maybe another font
			fontFace,
			op,
		)
		curPosX += float64(statsText[i].sizeNeeded)
	}
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
	// return 1024, 768
}

func init() {
	// load player
	img, _, err := image.Decode(bytes.NewReader(playerPNG))
	if err != nil {
		panic(err)
	}

	playerSheet = ebiten.NewImageFromImage(img)

	// load enemies
	monsters, err := enemyI.LoadEnemyImages(monsterVirtualFileSystem, "assets/monsters")
	if err != nil {
		panic(err)
	}

	enemy_images = monsters

	// load font
	source, err := text.NewGoTextFaceSource(bytes.NewReader(fontBytes))
	if err != nil {
		panic(err)
	}

	fontFace = &text.GoTextFace{
		Source: source,
		Size:   24,
	}

	// create the init pool of enemyProjectiles
	enemyProjectiles = projectile.ProjectilesInit(enemyI.EnemiesCount)
	// create the init pool of playerProjectiles
	playerProjectiles = projectile.ProjectilesInit(enemyI.EnemiesCount)
}

func gameInit() *Game {
	return initGame()
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
