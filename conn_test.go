package gocraft_test

import (
	"bytes"
	"math"
	"net"
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
)

func TestConnRoundTrip(t *testing.T) {
	tests := []struct {
		threshold int
		packet    gocraft.Packet
	}{
		{
			threshold: -1,
			packet: gocraft.Packet{
				ID:      0x00,
				Payload: gocraft.Marshal(gocraft.VarInt(765), gocraft.String("mc.local"), gocraft.UShort(25565), gocraft.VarInt(1)),
			},
		},
		{
			threshold: -1,
			packet: gocraft.Packet{
				ID: 0x01,
			},
		},
		{
			threshold: 64,
			packet: gocraft.Packet{
				ID:      0x02,
				Payload: gocraft.Marshal(gocraft.String("below threshold")),
			},
		},
		{
			threshold: 16,
			packet: gocraft.Packet{
				ID:      0x03,
				Payload: bytes.Repeat(gocraft.Marshal(gocraft.String("chunk data")), 256),
			},
		},
	}

	for _, tt := range tests {
		client, server := net.Pipe()

		in := gocraft.NewConn(client)
		out := gocraft.NewConn(server)
		in.SetThreshold(tt.threshold)
		out.SetThreshold(tt.threshold)

		errs := make(chan error, 1)
		go func() {
			errs <- in.WritePacket(tt.packet)
		}()

		got, err := out.ReadPacket()

		if err != nil {
			t.Errorf("ReadPacket (threshold %d): %v", tt.threshold, err)
		} else if got.ID != tt.packet.ID || !bytes.Equal(got.Payload, tt.packet.Payload) {
			t.Errorf("packet 0x%02x round trip yielded id 0x%02x with %d bytes, want %d bytes",
				tt.packet.ID, got.ID, len(got.Payload), len(tt.packet.Payload))
		}

		if err := <-errs; err != nil {
			t.Errorf("WritePacket (threshold %d): %v", tt.threshold, err)
		}

		client.Close()
		server.Close()
	}
}

func TestConnRejectsMalformedFrame(t *testing.T) {
	tests := []struct {
		frame []byte
	}{
		{
			frame: gocraft.AppendVar(nil, int32(0)),
		},
		{
			frame: gocraft.AppendVar(nil, int32(-1)),
		},
		{
			frame: gocraft.AppendVar(nil, int32(math.MaxInt32)),
		},
	}

	for _, tt := range tests {
		client, server := net.Pipe()

		out := gocraft.NewConn(server)

		go func() {
			client.Write(tt.frame)
		}()

		if _, err := out.ReadPacket(); err == nil {
			t.Errorf("ReadPacket(frame %x): expected an error, got nil", tt.frame)
		}

		client.Close()
		server.Close()
	}
}
