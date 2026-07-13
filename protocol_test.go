package gocraft_test

import (
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
)

type keepAlivePacket struct {
	Nonce gocraft.Long
	Label gocraft.String
}

func (*keepAlivePacket) ID() int32 {
	return 0x2A
}

func (p keepAlivePacket) Append(dst []byte) []byte {
	return gocraft.Marshal(p.Nonce, p.Label)
}

func (p *keepAlivePacket) Decode(r *gocraft.Reader) error {
	if err := p.Nonce.Decode(r); err != nil {
		return err
	}

	return p.Label.Decode(r)
}

func TestProtocolDecodesRegisteredPacket(t *testing.T) {
	proto := gocraft.NewProtocol()
	gocraft.Bind[keepAlivePacket](proto, gocraft.StatePlay, gocraft.Clientbound)

	original := &keepAlivePacket{
		Nonce: 99,
		Label: "alive",
	}
	frame := gocraft.EncodeFrame(original)

	packet, ok, err := proto.Decode(gocraft.StatePlay, gocraft.Clientbound, frame)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("packet not registered in protocol")
	}

	got, isKeepAlive := packet.(*keepAlivePacket)
	if !isKeepAlive {
		t.Fatalf("decoded %T, want *keepAlivePacket", packet)
	}
	if *got != *original {
		t.Errorf("round trip got %+v, want %+v", got, original)
	}
}

func TestProtocolUnknownPacket(t *testing.T) {
	proto := gocraft.NewProtocol()

	_, ok, err := proto.Decode(gocraft.StatePlay, gocraft.Clientbound, gocraft.Frame{
		ID: 0x99,
	})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("unknown packet reported as registered")
	}
}

func TestProtocolIsolatesStateAndDirection(t *testing.T) {
	proto := gocraft.NewProtocol()
	gocraft.Bind[keepAlivePacket](proto, gocraft.StatePlay, gocraft.Clientbound)

	if _, ok := proto.New(gocraft.StateLogin, gocraft.Clientbound, 0x2A); ok {
		t.Error("packet leaked across states")
	}
	if _, ok := proto.New(gocraft.StatePlay, gocraft.Serverbound, 0x2A); ok {
		t.Error("packet leaked across directions")
	}
}

func TestBindPanicsOnDuplicateRegistration(t *testing.T) {
	proto := gocraft.NewProtocol()
	gocraft.Bind[keepAlivePacket](proto, gocraft.StatePlay, gocraft.Clientbound)

	defer func() {
		if recover() == nil {
			t.Error("expected a panic registering the same key twice, got none")
		}
	}()

	gocraft.Bind[keepAlivePacket](proto, gocraft.StatePlay, gocraft.Clientbound)
}

func TestStateAndDirectionString(t *testing.T) {
	if got := gocraft.StatePlay.String(); got != "play" {
		t.Errorf("StatePlay.String() = %q, want play", got)
	}
	if got := gocraft.Clientbound.String(); got != "clientbound" {
		t.Errorf("Clientbound.String() = %q, want clientbound", got)
	}
}
