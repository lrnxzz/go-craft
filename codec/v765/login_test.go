package v765_test

import (
	"reflect"
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765"
)

func TestLoginStartCarriesUsernameAndUUID(t *testing.T) {
	original := &v765.LoginStart{
		Username: "gocraft",
		UUID:     gocraft.UUID{0x11, 0x22, 0x33},
	}

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(gocraft.StateLogin, gocraft.Serverbound, gocraft.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("login start not registered")
	}

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

	proto := v765.Protocol()
	decoded, ok, err := proto.Decode(gocraft.StateLogin, gocraft.Clientbound, gocraft.EncodeFrame(original))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("login success not registered")
	}

	if got := decoded.(*v765.LoginSuccess); !reflect.DeepEqual(got, original) {
		t.Errorf("got %+v, want %+v", got, original)
	}
}

func TestLoginAcknowledgedIsEmpty(t *testing.T) {
	frame := gocraft.EncodeFrame(&v765.LoginAcknowledged{})

	if len(frame.Payload) != 0 {
		t.Errorf("login acknowledged payload = %d bytes, want 0", len(frame.Payload))
	}

	proto := v765.Protocol()
	_, ok, err := proto.Decode(gocraft.StateLogin, gocraft.Serverbound, frame)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("login acknowledged not registered")
	}
}
