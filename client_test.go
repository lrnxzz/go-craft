package gocraft_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
)

func TestClientDispatchesReceivedPacket(t *testing.T) {
	clientSide, serverSide := net.Pipe()

	proto := gocraft.NewProtocol()
	gocraft.Bind[keepAlivePacket](proto, gocraft.StatePlay, gocraft.Clientbound)

	client := gocraft.NewClient(gocraft.NewConn(clientSide), proto)
	client.SetState(gocraft.StatePlay)

	got := make(chan keepAlivePacket, 1)
	gocraft.On[*keepAlivePacket](client, gocraft.StatePlay, func(c *gocraft.Client, p *keepAlivePacket) error {
		got <- *p

		return c.Close()
	})

	go func() {
		server := gocraft.NewConn(serverSide)
		server.WriteFrame(gocraft.EncodeFrame(&keepAlivePacket{
			Nonce: 7,
			Label: "hi",
		}))
	}()

	if err := client.Run(context.Background()); err != nil {
		t.Fatal(err)
	}

	select {
	case p := <-got:
		if p.Nonce != 7 || p.Label != "hi" {
			t.Errorf("dispatched %+v, want {7 hi}", p)
		}
	default:
		t.Fatal("handler was not called")
	}
}

func TestClientSkipsUnhandledPackets(t *testing.T) {
	clientSide, serverSide := net.Pipe()

	proto := gocraft.NewProtocol()
	gocraft.Bind[keepAlivePacket](proto, gocraft.StatePlay, gocraft.Clientbound)

	client := gocraft.NewClient(gocraft.NewConn(clientSide), proto)
	client.SetState(gocraft.StatePlay)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go func() {
		server := gocraft.NewConn(serverSide)
		server.WriteFrame(gocraft.EncodeFrame(&keepAlivePacket{
			Nonce: 1,
		}))
	}()

	if err := client.Run(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Run with no handler = %v, want deadline exceeded (packet skipped, loop kept reading)", err)
	}
}

func TestClientStopsOnContextCancel(t *testing.T) {
	clientSide, _ := net.Pipe()

	client := gocraft.NewClient(gocraft.NewConn(clientSide), gocraft.NewProtocol())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := client.Run(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("Run = %v, want context.Canceled", err)
	}
}

func TestClientReaderNotBlockedByPendingWrites(t *testing.T) {
	clientSide, serverSide := net.Pipe()

	proto := gocraft.NewProtocol()
	gocraft.Bind[keepAlivePacket](proto, gocraft.StatePlay, gocraft.Clientbound)

	client := gocraft.NewClient(gocraft.NewConn(clientSide), proto)
	client.SetState(gocraft.StatePlay)

	const count = 2
	seen := make(chan gocraft.Long, count)
	gocraft.On[*keepAlivePacket](client, gocraft.StatePlay, func(c *gocraft.Client, p *keepAlivePacket) error {
		seen <- p.Nonce

		return c.Send(&keepAlivePacket{Nonce: p.Nonce})
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go client.Run(ctx)

	server := gocraft.NewConn(serverSide)
	for i := range count {
		if err := server.WriteFrame(gocraft.EncodeFrame(&keepAlivePacket{Nonce: gocraft.Long(i + 1)})); err != nil {
			t.Fatal(err)
		}
	}

	for range count {
		select {
		case <-seen:
		case <-time.After(time.Second):
			t.Fatal("reader stalled — a handler's Send blocked the read loop")
		}
	}
}

func TestClientSendEncodesFrame(t *testing.T) {
	clientSide, serverSide := net.Pipe()

	client := gocraft.NewClient(gocraft.NewConn(clientSide), gocraft.NewProtocol())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go client.Run(ctx)

	if err := client.Send(&keepAlivePacket{Nonce: 9, Label: "x"}); err != nil {
		t.Fatal(err)
	}

	server := gocraft.NewConn(serverSide)
	frame, err := server.ReadFrame()
	if err != nil {
		t.Fatal(err)
	}

	if frame.ID != 0x2A {
		t.Errorf("frame id = 0x%02x, want 0x2A", frame.ID)
	}
}
