package gocraft

import (
	"fmt"
	"math"
)

type Position struct {
	X int32
	Y int32
	Z int32
}

func (p Position) Append(dst []byte) []byte {
	packed := int64(p.X&0x3FFFFFF)<<38 | int64(p.Z&0x3FFFFFF)<<12 | int64(p.Y&0xFFF)

	return Long(packed).Append(dst)
}

func (p *Position) Decode(r *Reader) error {
	var packed Long
	if err := packed.Decode(r); err != nil {
		return err
	}

	p.X = int32(packed >> 38)
	p.Y = int32(packed << 52 >> 52)
	p.Z = int32(packed << 26 >> 38)

	return nil
}

func (p Position) Add(dx, dy, dz int32) Position {
	return Position{X: p.X + dx, Y: p.Y + dy, Z: p.Z + dz}
}

func (p Position) String() string {
	return fmt.Sprintf("(%d, %d, %d)", p.X, p.Y, p.Z)
}

type Angle uint8

func (a Angle) Append(dst []byte) []byte {
	return UByte(a).Append(dst)
}

func (a *Angle) Decode(r *Reader) error {
	var raw UByte
	if err := raw.Decode(r); err != nil {
		return err
	}

	*a = Angle(raw)

	return nil
}

func (a Angle) Degrees() float64 {
	return float64(a) * 360 / 256
}

func (a Angle) Radians() float64 {
	return float64(a) * 2 * math.Pi / 256
}

func AngleFromDegrees(degrees float64) Angle {
	return Angle(int64(math.Round(degrees / 360 * 256)))
}

func AngleFromRadians(radians float64) Angle {
	return Angle(int64(math.Round(radians / (2 * math.Pi) * 256)))
}

var (
	_ Field    = Position{}
	_ FieldPtr = (*Position)(nil)
	_ Field    = Angle(0)
	_ FieldPtr = (*Angle)(nil)
)
