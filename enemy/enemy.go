// Package enemy all about enemies
package enemy

import (
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/alex/ebiten_tutorial/constants"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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

func CreateInit(maxWidth float64, maxHeight float64) []Enemy {
	var enemies []Enemy
	for i := 0; i < EnemiesCount; i++ {
		aroundLvl := 1
		if i%2 == 0 {
			aroundLvl = 2
		}

		randomWidth := rand.Float64()*(maxWidth-1) + 1
		randomHeight := rand.Float64()*(maxHeight-1) + 1
		enemies = append(enemies, Enemy{
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
			AttackSpeed: enemyAttackSpeedLookup[aroundLvl-1],
			LastAttack:  time.Now(),
		})
	}
	return enemies
}

func LoadEnemyImages(dir string) ([]*ebiten.Image, error) {
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
