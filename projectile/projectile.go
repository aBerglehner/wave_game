package projectile

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
	Radius   float32
}

func NewEnemyProjectile(pos Pos, velocity Pos, dmg int) EnemyProjectile {
	// TODO: maybe projectile size constant or lvl lookup
	return EnemyProjectile{OldPos: pos, CurPos: pos, Velocity: velocity, Dmg: dmg, Alive: true, Radius: 5}
}

func EnemyProjectilesInit(size int) []EnemyProjectile {
	var result []EnemyProjectile
	for i := 0; i < size; i++ {
		ep := EnemyProjectile{}
		result = append(result, ep)
	}
	return result
}
