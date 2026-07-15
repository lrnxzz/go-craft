package gocraft

const (
	gravity      = 0.08
	verticalDrag = 0.98
	playerWidth  = 0.6
	playerHeight = 1.8
)

type Collider func(BlockState) []AABB

type Physics struct {
	Velocity Vec3d
	collider Collider
}

func NewPhysics(collider Collider) *Physics {
	return &Physics{
		collider: collider,
	}
}

func (p *Physics) Tick(world *World, player *Player, controls Controls) {
	heading := controls.heading(player.Yaw)
	speed := walkSpeed
	if controls.Sprint {
		speed = sprintSpeed
	}
	p.Velocity.X = heading.X * speed
	p.Velocity.Z = heading.Z * speed

	if controls.Jump && player.OnGround {
		p.Velocity.Y = jumpVelocity
	}

	box := BoxAround(player.Position, playerWidth, playerHeight)
	moved := p.collide(world, box, p.Velocity)

	player.OnGround = p.Velocity.Y < 0 && moved.Y != p.Velocity.Y
	player.Position = player.Position.Add(moved)

	if moved.X != p.Velocity.X {
		p.Velocity.X = 0
	}
	if moved.Y != p.Velocity.Y {
		p.Velocity.Y = 0
	}
	if moved.Z != p.Velocity.Z {
		p.Velocity.Z = 0
	}

	p.Velocity.Y -= gravity
	p.Velocity.Y *= verticalDrag
}

func (p *Physics) collide(world *World, box AABB, velocity Vec3d) Vec3d {
	obstacles := p.obstacles(world, box.Stretch(velocity.X, velocity.Y, velocity.Z))

	dy := velocity.Y
	for _, o := range obstacles {
		dy = o.ClampY(box, dy)
	}
	box = box.OffsetXYZ(0, dy, 0)

	dx := velocity.X
	for _, o := range obstacles {
		dx = o.ClampX(box, dx)
	}
	box = box.OffsetXYZ(dx, 0, 0)

	dz := velocity.Z
	for _, o := range obstacles {
		dz = o.ClampZ(box, dz)
	}

	return Vec3(dx, dy, dz)
}

func (p *Physics) obstacles(world *World, region AABB) []AABB {
	lo, hi := region.Min.Floor(), region.Max.Floor()

	var boxes []AABB
	for x := lo.X; x <= hi.X; x++ {
		for y := lo.Y; y <= hi.Y; y++ {
			for z := lo.Z; z <= hi.Z; z++ {
				corner := Vec3(float64(x), float64(y), float64(z))

				state, ok := world.Block(x, y, z)
				if !ok {
					boxes = append(boxes, Box(corner, corner.Offset(1, 1, 1)))

					continue
				}

				for _, shape := range p.collider(state) {
					boxes = append(boxes, shape.Offset(corner))
				}
			}
		}
	}

	return boxes
}
