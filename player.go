package gocraft

type GameMode uint8

const (
	Survival GameMode = iota
	Creative
	Adventure
	Spectator
)

type Player struct {
	Position Vec3d
	Yaw      float32
	Pitch    float32
	OnGround bool
	Health   float32
	Food     int32
	GameMode GameMode
}
