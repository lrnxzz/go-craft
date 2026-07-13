package gocraft

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

const outboundBuffer = 256

type handlerKey struct {
	state State
	id    int32
}

type packetHandler func(*Client, Packet) error

type Client struct {
	conn      *Conn
	protocol  *Protocol
	state     atomic.Uint32
	handlers  map[handlerKey][]packetHandler
	outbound  chan Packet
	done      chan struct{}
	closeOnce sync.Once
}

func NewClient(conn *Conn, protocol *Protocol) *Client {
	client := &Client{
		conn:     conn,
		protocol: protocol,
		handlers: make(map[handlerKey][]packetHandler),
		outbound: make(chan Packet, outboundBuffer),
		done:     make(chan struct{}),
	}

	return client
}

func (c *Client) State() State {
	return State(c.state.Load())
}

func (c *Client) SetState(state State) {
	c.state.Store(uint32(state))
}

func (c *Client) Send(packet Packet) error {
	select {
	case c.outbound <- packet:
		return nil
	case <-c.done:
		return fmt.Errorf("gocraft: send on a closed client")
	}
}

func (c *Client) SetCompression(threshold int) {
	c.conn.SetThreshold(threshold)
}

func (c *Client) Close() error {
	c.closeOnce.Do(func() {
		close(c.done)
	})

	return c.conn.Close()
}

func (c *Client) closed() bool {
	select {
	case <-c.done:
		return true
	default:
		return false
	}
}

func On[P Packet](c *Client, state State, handler func(*Client, P) error) {
	var prototype P
	key := handlerKey{
		state: state,
		id:    prototype.ID(),
	}

	c.handlers[key] = append(c.handlers[key], func(client *Client, packet Packet) error {
		return handler(client, packet.(P))
	})
}

func (c *Client) Run(parent context.Context) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	stop := context.AfterFunc(ctx, func() {
		c.Close()
	})
	defer stop()

	var group errgroup.Group

	group.Go(func() error {
		defer cancel()

		return c.readLoop(ctx)
	})
	group.Go(func() error {
		defer cancel()

		return c.writeLoop(ctx)
	})

	return group.Wait()
}

func (c *Client) readLoop(ctx context.Context) error {
	for {
		frame, err := c.conn.ReadFrame()
		if err != nil {
			return c.exit(ctx, err)
		}

		if err := c.receive(frame); err != nil {
			return err
		}
	}
}

func (c *Client) writeLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case packet := <-c.outbound:
			if err := c.conn.WriteFrame(EncodeFrame(packet)); err != nil {
				return err
			}
		}
	}
}

func (c *Client) receive(frame Frame) error {
	packet, known, err := c.protocol.Decode(c.State(), Clientbound, frame)
	if err != nil || !known {
		return err
	}

	return c.dispatch(packet)
}

func (c *Client) exit(ctx context.Context, readErr error) error {
	switch {
	case ctx.Err() != nil:
		return ctx.Err()
	case c.closed():
		return nil
	default:
		return readErr
	}
}

func (c *Client) dispatch(packet Packet) error {
	key := handlerKey{
		state: c.State(),
		id:    packet.ID(),
	}

	for _, handler := range c.handlers[key] {
		if err := handler(c, packet); err != nil {
			return err
		}
	}

	return nil
}
