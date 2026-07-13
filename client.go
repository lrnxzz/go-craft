package gocraft

import (
	"context"
	"sync/atomic"
)

type handlerKey struct {
	state State
	id    int32
}

type Client struct {
	conn     *Conn
	protocol *Protocol
	state    atomic.Uint32
	stopped  atomic.Bool
	handlers map[handlerKey][]func(*Client, Packet) error
}

func NewClient(conn *Conn, protocol *Protocol) *Client {
	client := &Client{
		conn:     conn,
		protocol: protocol,
		handlers: make(map[handlerKey][]func(*Client, Packet) error),
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
	return c.conn.WriteFrame(EncodeFrame(packet))
}

func (c *Client) SetCompression(threshold int) {
	c.conn.SetThreshold(threshold)
}

func (c *Client) Close() error {
	c.stopped.Store(true)

	return c.conn.Close()
}

func On[P Packet](c *Client, state State, handler func(*Client, P) error) {
	var prototype P
	key := handlerKey{
		state: state,
		id:    prototype.ID(),
	}

	wrapped := func(client *Client, packet Packet) error {
		return handler(client, packet.(P))
	}

	c.handlers[key] = append(c.handlers[key], wrapped)
}

func (c *Client) Run(ctx context.Context) error {
	stop := context.AfterFunc(ctx, func() {
		c.conn.Close()
	})
	defer stop()

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

func (c *Client) receive(frame Frame) error {
	packet, known, err := c.protocol.Decode(c.State(), Clientbound, frame)
	if err != nil || !known {
		return err
	}

	return c.dispatch(packet)
}

func (c *Client) exit(ctx context.Context, readErr error) error {
	switch {
	case c.stopped.Load():
		return nil
	case ctx.Err() != nil:
		return ctx.Err()
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
