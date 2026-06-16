// Package enemy all about enemies
package enemy

import (
	"time"

	"github.com/alex/ebiten_tutorial/constants"
)

// can all be looked up via -> enemies lvl +1 -> lvl 1 = index 0
var (
	// TODO: fill stats
	enemyDmgLookup    [constants.LvlMax]int = [...]int{0, 1, 0, 0, 0, 0}
	enemyHealthLookup [constants.LvlMax]int = [...]int{0, 1, 0, 0, 0, 0}
	enemyExpLookup    [constants.LvlMax]int = [...]int{0, 1, 0, 0, 0, 0}
	// slower on lower lvl
	enemyAttackSpeedLookup [constants.LvlMax]int = [...]int{0, 5_000, 4_000, 2_000, 1_000, 500}
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

func CreateInit(count int, aroundLvl int) [10]Enemy {
	enemies := [10]Enemy{}
	for i := 0; i < count; i++ {
		enemies[i] = Enemy{
			posX:   0, // todo in range
			posY:   0, // todo in rage
			alive:  true,
			lvl:    aroundLvl, // todo around
			dmg:    1,         // todo
			health: 100,       // todo
			exp:    1,         // todo
			// ms
			attackSpeed: 5000, // todo
			lastAttack:  time.Now(),
		}
	}
	return enemies
}
