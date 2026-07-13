package gocraft_test

import (
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
)

func TestPositionRoundTrip(t *testing.T) {
	positions := []gocraft.Position{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 2, Z: 3},
		{X: -1, Y: -1, Z: -1},
		{X: 33554431, Y: 2047, Z: 33554431},
		{X: -33554432, Y: -2048, Z: -33554432},
		{X: 33554431, Y: -2048, Z: -33554432},
	}

	for _, want := range positions {
		var got gocraft.Position

		if err := gocraft.Unmarshal(want.Append(nil), &got); err != nil {
			t.Errorf("decode %s: %v", want, err)
			continue
		}
		if got != want {
			t.Errorf("round trip of %s yielded %s", want, got)
		}
	}
}

func TestAngleDegrees(t *testing.T) {
	tests := []struct {
		angle   gocraft.Angle
		degrees float64
	}{
		{
			angle:   0,
			degrees: 0,
		},
		{
			angle:   64,
			degrees: 90,
		},
		{
			angle:   128,
			degrees: 180,
		},
		{
			angle:   192,
			degrees: 270,
		},
	}

	for _, tt := range tests {
		if got := tt.angle.Degrees(); got != tt.degrees {
			t.Errorf("Angle(%d).Degrees() = %g, want %g", tt.angle, got, tt.degrees)
		}
		if got := gocraft.AngleFromDegrees(tt.degrees); got != tt.angle {
			t.Errorf("AngleFromDegrees(%g) = %d, want %d", tt.degrees, got, tt.angle)
		}
	}
}

func TestAngleRoundTrip(t *testing.T) {
	for _, want := range []gocraft.Angle{0, 1, 64, 128, 200, 255} {
		var got gocraft.Angle

		if err := gocraft.Unmarshal(want.Append(nil), &got); err != nil {
			t.Errorf("decode %d: %v", want, err)
			continue
		}
		if got != want {
			t.Errorf("round trip of %d yielded %d", want, got)
		}
	}
}
