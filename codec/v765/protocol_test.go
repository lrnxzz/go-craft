package v765_test

import (
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
	v765 "github.com/lrnxzz/go-craft/codec/v765"
)

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
