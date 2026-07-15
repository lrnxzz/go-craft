package agent

import (
	"context"
	"fmt"
	"math"
	"net"
	"strconv"
	"sync"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	v765 "github.com/lrnxzz/go-craft/codec/v765"
	"github.com/lrnxzz/go-craft/codec/v765/blocks"
)

const (
	tickRate     = 50 * time.Millisecond
	arriveRadius = 0.6
)

type Agent struct {
	client  *gocraft.Client
	session *v765.Session
	physics *gocraft.Physics

	mu       sync.Mutex
	controls gocraft.Controls
	yaw      float32
	pitch    float32
	look     bool
	goal     *gocraft.Vec3d

	onSpawn      func()
	spawnedFired bool
	ticks        uint64
	snapshot     Snapshot
}

type Snapshot struct {
	Tick     uint64
	Position gocraft.Vec3d
	Yaw      float32
	Pitch    float32
	OnGround bool
	Health   float32
}

func (a *Agent) Snapshot() Snapshot {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.snapshot
}

func Join(ctx context.Context, address, username string) (*Agent, error) {
	host, port, err := splitAddress(address)
	if err != nil {
		return nil, err
	}

	conn, err := gocraft.Dial(ctx, net.JoinHostPort(host, strconv.Itoa(int(port))))
	if err != nil {
		return nil, err
	}

	client := gocraft.NewClient(conn, v765.Protocol())
	a := &Agent{client: client, physics: gocraft.NewPhysics(blocks.Collision)}

	session, err := v765.Join(client, host, port, username, nil)
	if err != nil {
		client.Close()

		return nil, err
	}
	a.session = session

	client.Tick(tickRate, a.tick)

	return a, nil
}

func (a *Agent) Run(ctx context.Context) error {
	return a.client.Run(ctx)
}

func (a *Agent) World() *gocraft.World {
	return a.session.World()
}

func (a *Agent) Player() *gocraft.Player {
	return a.session.Player()
}

func (a *Agent) SetControls(controls gocraft.Controls) {
	a.mu.Lock()
	a.controls = controls
	a.mu.Unlock()
}

func (a *Agent) Look(yaw, pitch float32) {
	a.mu.Lock()
	a.yaw, a.pitch, a.look = yaw, pitch, true
	a.mu.Unlock()
}

func (a *Agent) LookAt(target gocraft.Vec3d) {
	player := a.session.Player()
	yaw, pitch := gocraft.LookAngles(player.Eye(), target)

	a.Look(yaw, pitch)
}

func (a *Agent) Target(reach float64) (gocraft.RayHit, bool) {
	player := a.session.Player()

	return a.session.World().Raycast(player.Eye(), player.LookDirection(), reach, blocks.Solid)
}

func (a *Agent) OnSpawn(fn func()) {
	a.onSpawn = fn
}

func (a *Agent) MoveTo(target gocraft.Vec3d) {
	a.mu.Lock()
	a.goal = &target
	a.mu.Unlock()
}

func (a *Agent) Stop() {
	a.mu.Lock()
	a.goal = nil
	a.controls = gocraft.Controls{}
	a.mu.Unlock()
}

func (a *Agent) tick() {
	if !a.session.Spawned() {
		return
	}
	if a.onSpawn != nil && !a.spawnedFired {
		a.spawnedFired = true
		a.onSpawn()
	}

	player := a.session.Player()
	a.pursue(player)

	a.mu.Lock()
	controls := a.controls
	yaw, pitch, look := a.yaw, a.pitch, a.look
	a.mu.Unlock()

	if look {
		player.Yaw = yaw
		player.Pitch = pitch
	}

	a.physics.Tick(a.session.World(), player, controls)
	_ = a.session.SendPosition()

	a.mu.Lock()
	a.ticks++
	a.snapshot = Snapshot{
		Tick:     a.ticks,
		Position: player.Position,
		Yaw:      player.Yaw,
		Pitch:    player.Pitch,
		OnGround: player.OnGround,
		Health:   player.Health,
	}
	a.mu.Unlock()
}

func (a *Agent) pursue(player *gocraft.Player) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.goal == nil {
		return
	}

	dx := a.goal.X - player.Position.X
	dz := a.goal.Z - player.Position.Z
	if dx*dx+dz*dz < arriveRadius*arriveRadius {
		a.goal = nil
		a.controls.Forward = false

		return
	}

	a.yaw = float32(math.Atan2(-dx, dz) * 180 / math.Pi)
	a.pitch = 0
	a.look = true
	a.controls.Forward = true
}

func splitAddress(address string) (string, uint16, error) {
	host, raw, err := net.SplitHostPort(address)
	if err != nil {
		return address, 25565, nil
	}

	port, err := strconv.ParseUint(raw, 10, 16)
	if err != nil {
		return "", 0, fmt.Errorf("agent: invalid port in %q", address)
	}

	return host, uint16(port), nil
}
