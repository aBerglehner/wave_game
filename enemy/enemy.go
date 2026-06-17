// Package enemy all about enemies
package enemy

import (
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
	enemyAttackSpeedLookup [constants.LvlMax]int = [...]int{1_000, 800, 750, 650, 600, 550, 500, 450, 400, 300}
)

// TODO: spwan enemies
type Enemy struct {
	posX   float32
	posY   float32
	alive  bool
	lvl    int
	dmg    int
	health int
	exp    int
	// ms
	attackSpeed int
	lastAttack  time.Time
}

func CreateInit() []Enemy {
	var enemies []Enemy
	for i := 0; i < EnemiesCount; i++ {
		aroundLvl := 1
		if i%2 == 0 {
			aroundLvl = 2
		}

		enemies = append(enemies, Enemy{
			posX:   0, // todo in range
			posY:   0, // todo in rage
			alive:  true,
			lvl:    aroundLvl,
			dmg:    enemyDmgLookup[0],
			health: enemyHealthLookup[0],
			exp:    enemyExpLookup[0],
			// ms
			attackSpeed: enemyAttackSpeedLookup[0],
			lastAttack:  time.Now(),
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
