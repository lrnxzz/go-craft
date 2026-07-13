package v765_test

import (
	"reflect"
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765"
)

func encodeAndDecode(t *testing.T, state gocraft.State, dir gocraft.Direction, packet gocraft.Packet) gocraft.Packet {
	t.Helper()

	proto := v765.Protocol()
	frame := gocraft.EncodeFrame(packet)

	decoded, ok, err := proto.Decode(state, dir, frame)
	if err != nil {
		t.Fatalf("decode 0x%02x: %v", packet.ID(), err)
	}
	if !ok {
		t.Fatalf("packet 0x%02x not registered for %s/%s", packet.ID(), state, dir)
	}

	return decoded
}

func TestHandshakeCarriesConnectionParameters(t *testing.T) {
	original := &v765.Handshake{
		ProtocolVersion: v765.ProtocolVersion,
		ServerAddress:   "mc.local",
		ServerPort:      25565,
		NextState:       gocraft.VarInt(gocraft.StateLogin),
	}

	decoded := encodeAndDecode(t, gocraft.StateHandshaking, gocraft.Serverbound, original)

	if got := decoded.(*v765.Handshake); *got != *original {
		t.Errorf("got %+v, want %+v", got, original)
	}
}

func TestLoginStartCarriesUsernameAndUUID(t *testing.T) {
	original := &v765.LoginStart{
		Username: "gocraft",
		UUID:     gocraft.UUID{0x11, 0x22, 0x33},
	}

	decoded := encodeAndDecode(t, gocraft.StateLogin, gocraft.Serverbound, original)

	if got := decoded.(*v765.LoginStart); *got != *original {
		t.Errorf("got %+v, want %+v", got, original)
	}
}

func TestLoginSuccessCarriesProfileProperties(t *testing.T) {
	original := &v765.LoginSuccess{
		UUID:     gocraft.UUID{0xAB, 0xCD},
		Username: "gocraft",
		Properties: gocraft.Slice[v765.Property]{
			{Name: "textures", Value: "base64", Signature: gocraft.Some(gocraft.String("sig"))},
			{Name: "plain", Value: "value"},
		},
	}

	decoded := encodeAndDecode(t, gocraft.StateLogin, gocraft.Clientbound, original)

	if got := decoded.(*v765.LoginSuccess); !reflect.DeepEqual(got, original) {
		t.Errorf("got %+v, want %+v", got, original)
	}
}

func TestLoginAcknowledgedIsEmpty(t *testing.T) {
	frame := gocraft.EncodeFrame(&v765.LoginAcknowledged{})

	if len(frame.Payload) != 0 {
		t.Errorf("login acknowledged payload = %d bytes, want 0", len(frame.Payload))
	}

	encodeAndDecode(t, gocraft.StateLogin, gocraft.Serverbound, &v765.LoginAcknowledged{})
}

func TestUnknownPacketIsSkipped(t *testing.T) {
	proto := v765.Protocol()

	_, ok, err := proto.Decode(gocraft.StateLogin, gocraft.Clientbound, gocraft.Frame{
		ID: 0x7F,
	})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("unregistered packet reported as known")
	}
}
