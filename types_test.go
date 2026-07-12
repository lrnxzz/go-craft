package gocraft_test

import (
	"bytes"
	"encoding/binary"
	"math"
	"reflect"
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
)

type profileEntry struct {
	Name gocraft.String
	ID   gocraft.UUID
}

func (p profileEntry) Append(dst []byte) []byte {
	dst = p.Name.Append(dst)

	return p.ID.Append(dst)
}

func (p *profileEntry) Decode(r *gocraft.Reader) error {
	if err := p.Name.Decode(r); err != nil {
		return err
	}

	return p.ID.Decode(r)
}

func TestFieldRoundTrip(t *testing.T) {
	var id gocraft.UUID
	for i := range id {
		id[i] = byte(i)
	}

	fields := []gocraft.Field{
		gocraft.Bool(true),
		gocraft.Bool(false),
		gocraft.Byte(math.MinInt8),
		gocraft.UByte(math.MaxUint8),
		gocraft.Short(math.MinInt16),
		gocraft.UShort(math.MaxUint16),
		gocraft.Int(math.MinInt32),
		gocraft.Long(math.MinInt64),
		gocraft.Float(math.Pi),
		gocraft.Double(-math.MaxFloat64),
		gocraft.VarInt(-1),
		gocraft.VarLong(math.MinInt64),
		gocraft.String("gocraft ⛏"),
		id,
		gocraft.Slice[gocraft.VarInt]{0, -1, 25565, math.MaxInt32},
		gocraft.Slice[profileEntry]{
			{Name: "steve", ID: id},
			{Name: "alex"},
		},
		gocraft.Some(gocraft.String("skin-data")),
		gocraft.None[gocraft.String](),
	}

	for _, field := range fields {
		decoded := reflect.New(reflect.TypeOf(field))

		if err := gocraft.Unmarshal(field.Append(nil), decoded.Interface().(gocraft.FieldPtr)); err != nil {
			t.Errorf("decode %#v: %v", field, err)
			continue
		}

		if got := decoded.Elem().Interface(); !reflect.DeepEqual(got, field) {
			t.Errorf("round trip of %#v yielded %#v", field, got)
		}
	}
}

func TestFixedEncodingAgainstStdlib(t *testing.T) {
	var expected bytes.Buffer

	err := binary.Write(&expected, binary.BigEndian, struct {
		Flag   bool
		Kind   int8
		Level  uint8
		Delta  int16
		Port   uint16
		Block  int32
		Seed   int64
		Angle  float32
		Health float64
	}{true, math.MinInt8, math.MaxUint8, math.MinInt16, 25565, math.MinInt32, math.MinInt64, math.Pi, math.Pi})
	if err != nil {
		t.Fatal(err)
	}

	payload := gocraft.Marshal(
		gocraft.Bool(true),
		gocraft.Byte(math.MinInt8),
		gocraft.UByte(math.MaxUint8),
		gocraft.Short(math.MinInt16),
		gocraft.UShort(25565),
		gocraft.Int(math.MinInt32),
		gocraft.Long(math.MinInt64),
		gocraft.Float(math.Pi),
		gocraft.Double(math.Pi),
	)

	if !bytes.Equal(payload, expected.Bytes()) {
		t.Errorf("payload = %x, want %x", payload, expected.Bytes())
	}
}

func TestMarshalUnmarshalHandshake(t *testing.T) {
	payload := gocraft.Marshal(
		gocraft.VarInt(765),
		gocraft.String("mc.local"),
		gocraft.UShort(25565),
		gocraft.VarInt(1),
	)

	var (
		protocolVersion gocraft.VarInt
		serverAddress   gocraft.String
		serverPort      gocraft.UShort
		nextState       gocraft.VarInt
	)

	if err := gocraft.Unmarshal(payload, &protocolVersion, &serverAddress, &serverPort, &nextState); err != nil {
		t.Fatal(err)
	}

	if protocolVersion != 765 || serverAddress != "mc.local" || serverPort != 25565 || nextState != 1 {
		t.Errorf("decoded (%d, %q, %d, %d), want (765, \"mc.local\", 25565, 1)",
			protocolVersion, serverAddress, serverPort, nextState)
	}
}

func TestStringRejectsMalformedPayload(t *testing.T) {
	truncated := gocraft.String("gocraft").Append(nil)

	tests := []struct {
		input []byte
	}{
		{
			input: gocraft.AppendVar(nil, int32(-1)),
		},
		{
			input: truncated[:len(truncated)-1],
		},
	}

	for _, tt := range tests {
		var s gocraft.String

		if err := gocraft.Unmarshal(tt.input, &s); err == nil {
			t.Errorf("Unmarshal(%x): expected an error, got nil", tt.input)
		}
	}
}

func TestSliceRejectsOverclaimedCount(t *testing.T) {
	var s gocraft.Slice[gocraft.VarInt]

	if err := gocraft.Unmarshal(gocraft.AppendVar(nil, int32(math.MaxInt32)), &s); err == nil {
		t.Error("expected an error, got nil")
	}
}

func TestOptionGet(t *testing.T) {
	value, ok := gocraft.Some(gocraft.VarInt(7)).Get()

	if !ok || value != 7 {
		t.Errorf("Some(7).Get() = (%d, %t), want (7, true)", value, ok)
	}

	if _, ok := gocraft.None[gocraft.VarInt]().Get(); ok {
		t.Error("None().Get() reported presence")
	}
}

func TestNoneEncodesAsSingleByte(t *testing.T) {
	raw := gocraft.None[gocraft.VarInt]().Append(nil)

	if len(raw) != 1 {
		t.Errorf("None() encoded as %d bytes, want 1", len(raw))
	}
}
