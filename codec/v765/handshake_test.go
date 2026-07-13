package v765_test

import (
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765"
)

func TestHandshakeCarriesConnectionParameters(t *testing.T) {
	original := &v765.Handshake{
		ProtocolVersion: v765.ProtocolVersion,
		ServerAddress:   "mc.local",
		ServerPort:      25565,
		NextState:       gocraft.VarInt(gocraft.StateLogin),
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(gocraft.StateHandshaking, gocraft.Serverbound, gocraft.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("handshake not registered")
	}

	if got := decoded.(*v765.Handshake); *got != *original {
		t.Errorf("got %+v, want %+v", got, original)
	}
}
