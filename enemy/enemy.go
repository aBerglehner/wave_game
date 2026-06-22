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
	EnemiesCount = 10
)

// 0 indexd -> can all be looked up via -> enemies lvl - 1 -> lvl 1 = index 0
var (
	enemyDmgLookup    [constants.LvlMax]int = [...]int{1, 10, 100, 500, 1000, 10_000, 20_000, 50_000, 100_000, 200_000}
	enemyHealthLookup [constants.LvlMax]int = [...]int{10, 100, 1_000, 5_000, 10_000, 50_000, 100_000, 500_000, 900_000, 2_000_000}
	enemyExpLookup    [constants.LvlMax]int = [...]int{1, 10, 1_000, 2_000, 4_000, 8_000, 16_000, 32_000, 64_000, 128_000}
	// slower on lower lvl
	enemyAttackSpeedLookup [constants.LvlMax]int = [...]int{1_500, 1_000, 850, 700, 600, 550, 500, 400, 350, 200}
)

// TODO: spwan enemies
type Enemy struct {
	PosX   float32
	PosY   float32
	Alive  bool
	Lvl    int
	Dmg    int
	Health int
	Exp    int
	// ms
	AttackSpeed int
	LastAttack  time.Time
}

func CreateInit(maxWidth float32, maxHeight float32) []Enemy {
	var enemies []Enemy
	for i := 0; i < EnemiesCount; i++ {
		aroundLvl := 1
		if i%2 == 0 {
			aroundLvl = 2
		}

		randomWidth := rand.Float32()*(maxWidth-1) + 1
		randomHeight := rand.Float32()*(maxHeight-1) + 1
		enemies = append(enemies, Enemy{
			PosX:   randomWidth,
			PosY:   randomHeight,
			Alive:  true,
			Lvl:    aroundLvl,
			Dmg:    enemyDmgLookup[0],
			Health: enemyHealthLookup[0],
			Exp:    enemyExpLookup[0],
			// ms
			AttackSpeed: enemyAttackSpeedLookup[0],
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
