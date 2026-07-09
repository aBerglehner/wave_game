package projectile

type Pos struct {
	X float64
	Y float64
}

type Projectile struct {
	OldPos Pos
	CurPos Pos
	// divide / fps -> to get real number
	Velocity Pos
	Dmg      int
	Alive    bool
	Radius   float32
}

func NewProjectile(pos Pos, velocity Pos, dmg int) Projectile {
	// TODO: maybe projectile size constant or lvl lookup
	return Projectile{OldPos: pos, CurPos: pos, Velocity: velocity, Dmg: dmg, Alive: true, Radius: 5}
}

func ProjectilesInit(size int) []Projectile {
	var result []Projectile
	for i := 0; i < size; i++ {
		ep := Projectile{}
		result = append(result, ep)
	}
	return result
}
