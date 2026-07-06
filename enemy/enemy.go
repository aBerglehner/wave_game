// Package enemy all about enemies
package enemy

import (
	"bytes"
	"embed"
	"image"
	"math/rand"
	"path/filepath"
	"sort"
	"time"

	"github.com/alex/ebiten_tutorial/constants"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	EnemiesCount = 30
)

// 0 indexd -> can all be looked up via -> enemies lvl - 1 -> lvl 1 = index 0
var (
	enemyDmgLookup    [constants.LvlMax]int = [...]int{1, 10, 100, 500, 1000, 10_000, 20_000, 50_000, 100_000, 200_000}
	EnemyHealthLookup [constants.LvlMax]int = [...]int{10, 100, 1_000, 5_000, 10_000, 50_000, 100_000, 500_000, 900_000, 2_000_000}
	enemyExpLookup    [constants.LvlMax]int = [...]int{1, 10, 1_000, 2_000, 4_000, 8_000, 16_000, 32_000, 64_000, 128_000}
	// slower on lower lvl
	enemyAttackSpeedLookup [constants.LvlMax]int64 = [...]int64{1_500, 1_000, 850, 700, 600, 550, 500, 400, 350, 200}
	// pixels per second
	enemyProjectileSpeedLookup [constants.LvlMax]float64 = [...]float64{70, 80, 90, 100, 110, 120, 130, 140, 150, 160}
)

// TODO: spwan enemies
type Enemy struct {
	// this is just a counter so it will always walk a given ticks into one direction
	// when positiv the walk will be positiv and vice versa
	PosXMovingDirection int
	PosYMovingDirection int
	PosX                float64
	PosY                float64
	Alive               bool
	Lvl                 int
	Dmg                 int
	Health              int
	Exp                 int
	// ms
	AttackSpeed int64
	LastAttack  time.Time
	// pixels per second
	ProjectileSpeed float64
}

func newEnemy(randomWidth float64, randomHeight float64, aroundLvl int) Enemy {
	return Enemy{
		PosXMovingDirection: 0,
		PosYMovingDirection: 0,
		PosX:                randomWidth,
		PosY:                randomHeight,
		Alive:               true,
		Lvl:                 aroundLvl,
		Dmg:                 enemyDmgLookup[aroundLvl-1],
		Health:              EnemyHealthLookup[aroundLvl-1],
		Exp:                 enemyExpLookup[aroundLvl-1],
		// ms
		AttackSpeed:     enemyAttackSpeedLookup[aroundLvl-1],
		LastAttack:      time.Now(),
		ProjectileSpeed: enemyProjectileSpeedLookup[aroundLvl-1],
	}
}

func (e *Enemy) Patrol(maxWidth float64, maxHeight float64, moveDistance float64, fps int) {
	// fps*Seconds to reset the direction
	var directionResetTime int = fps * 5
	// don't let the enemies move as fast as the player
	moveDistance = moveDistance / 4

	// x part
	var xRanomizer float64 = 1
	if e.PosXMovingDirection == 0 {
		e.PosXMovingDirection = 1
		if rand.Intn(2) == 0 {
			e.PosXMovingDirection = -1
		}
	} else if e.PosXMovingDirection < 0 {
		xRanomizer = -1
		e.PosXMovingDirection -= 1
	} else {
		e.PosXMovingDirection += 1
	}

	newPosX := e.PosX + (rand.Float64()*moveDistance)*xRanomizer
	if newPosX < maxWidth && newPosX > 0 {
		e.PosX = newPosX
	}

	// reset after walked given time in certain direction
	if e.PosXMovingDirection == directionResetTime || e.PosXMovingDirection == -directionResetTime {
		e.PosXMovingDirection = 0
	}

	// y part
	var yRanomizer float64 = 1
	if e.PosYMovingDirection == 0 {
		e.PosYMovingDirection = 1
		if rand.Intn(2) == 0 {
			e.PosYMovingDirection = -1
		}
	} else if e.PosYMovingDirection < 0 {
		yRanomizer = -1
		e.PosYMovingDirection -= 1
	} else {
		e.PosYMovingDirection += 1
	}

	newPosY := e.PosY + (rand.Float64()*moveDistance)*yRanomizer
	if newPosY < maxHeight && newPosY > 0 {
		e.PosY = newPosY
	}

	// reset after walked given time in certain direction
	if e.PosYMovingDirection == directionResetTime || e.PosYMovingDirection == -directionResetTime {
		e.PosYMovingDirection = 0
	}
}

func EnemyCreateInit(maxWidth float64, maxHeight float64) []Enemy {
	var enemies []Enemy
	for i := 0; i < EnemiesCount; i++ {
		aroundLvl := 1
		if i%2 == 0 {
			aroundLvl = 2
		}

		randomWidth := rand.Float64()*(maxWidth-1) + 1
		randomHeight := rand.Float64()*(maxHeight-1) + 1
		enemies = append(enemies, newEnemy(randomWidth, randomHeight, aroundLvl))
	}
	return enemies
}

func LoadEnemyImages(fsys embed.FS, dir string) ([]*ebiten.Image, error) {
	files, err := fsys.ReadDir(dir)
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

		if filepath.Ext(file.Name()) != ".png" {
			continue
		}

		data, err := fsys.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}

		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}

		images = append(images, ebiten.NewImageFromImage(img))
	}

	return images, nil
}

type Pos struct {
	X float64
	Y float64
}

type EnemyProjectile struct {
	OldPos Pos
	CurPos Pos
	// divide / fps -> to get real number
	Velocity Pos
	Dmg      int
	Alive    bool
}

func NewEnemyProjectile(pos Pos, velocity Pos, dmg int) EnemyProjectile {
	return EnemyProjectile{OldPos: pos, CurPos: pos, Velocity: velocity, Dmg: dmg, Alive: true}
}

// TODO: idea will be have a pool of them and if really all of them are alive double the pool!!
func EnemyProjectilesInit(size int) []EnemyProjectile {
	var result []EnemyProjectile
	for i := 0; i < size; i++ {
		ep := EnemyProjectile{}
		result = append(result, ep)
	}
	return result
}
