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
	closed   atomic.Bool
	handlers map[handlerKey][]func(*Client, Packet) error
}

func NewClient(conn *Conn, protocol *Protocol) *Client {
	return &Client{
		conn:     conn,
		protocol: protocol,
		handlers: make(map[handlerKey][]func(*Client, Packet) error),
	}
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
	c.closed.Store(true)

	return c.conn.Close()
}

func On[T any, PT interface {
	*T
	Packet
}](c *Client, state State, handler func(*Client, PT) error) {
	var zero T
	key := handlerKey{state: state, id: PT(&zero).ID()}

	c.handlers[key] = append(c.handlers[key], func(client *Client, packet Packet) error {
		return handler(client, packet.(PT))
	})
}

func (c *Client) Run(ctx context.Context) error {
	defer context.AfterFunc(ctx, func() { c.conn.Close() })()

	for {
		frame, err := c.conn.ReadFrame()
		if err != nil {
			switch {
			case c.closed.Load():
				return nil
			case ctx.Err() != nil:
				return ctx.Err()
			default:
				return err
			}
		}

		packet, ok, err := c.protocol.Decode(c.State(), Clientbound, frame)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		if err := c.dispatch(packet); err != nil {
			return err
		}
	}
}

func (c *Client) dispatch(packet Packet) error {
	key := handlerKey{state: c.State(), id: packet.ID()}

	for _, handler := range c.handlers[key] {
		if err := handler(c, packet); err != nil {
			return err
		}
	}

	return nil
}
