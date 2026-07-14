package gocraft

const (
	gravity      = 0.08
	verticalDrag = 0.98
	playerWidth  = 0.6
	playerHeight = 1.8
)

type Physics struct {
	Velocity Vec3d
}

func (p *Physics) Tick(world *World, player *Player) {
	p.Velocity.Y -= gravity

	box := BoxAround(player.Position, playerWidth, playerHeight)
	moved := collide(world, box, p.Velocity)

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

	p.Velocity.Y *= verticalDrag
}

func collide(world *World, box AABB, velocity Vec3d) Vec3d {
	obstacles := solids(world, box.Stretch(velocity.X, velocity.Y, velocity.Z))

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

	return Vec3d{dx, dy, dz}
}

func solids(world *World, region AABB) []AABB {
	lo, hi := region.Min.Floor(), region.Max.Floor()

	var boxes []AABB
	for x := lo.X; x <= hi.X; x++ {
		for y := lo.Y; y <= hi.Y; y++ {
			for z := lo.Z; z <= hi.Z; z++ {
				state, ok := world.Block(x, y, z)
				if !ok || state == 0 {
					continue
				}
				boxes = append(boxes, cube(x, y, z))
			}
		}
	}

	return boxes
}

func cube(x, y, z int) AABB {
	corner := Vec3d{float64(x), float64(y), float64(z)}

	return AABB{Min: corner, Max: corner.Offset(1, 1, 1)}
}
