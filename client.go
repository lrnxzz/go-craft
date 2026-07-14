package gocraft

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
)

type Client struct {
	conn      *Conn
	protocol  *Protocol
	state     atomic.Uint32
	listeners listeners
	sender    *sender
	tick      ticker
	done      chan struct{}
	closeOnce sync.Once
}

func NewClient(conn *Conn, protocol *Protocol) *Client {
	client := &Client{
		conn:      conn,
		protocol:  protocol,
		listeners: make(listeners),
		sender:    newSender(conn),
		done:      make(chan struct{}),
	}

	return client
}

func (c *Client) State() State {
	return State(c.state.Load())
}

func (c *Client) SetState(state State) {
	slog.Debug("switched state", "to", state)
	c.state.Store(uint32(state))
}

func (c *Client) Send(packet Packet) error {
	select {
	case c.sender.outbound <- packet:
		return nil
	case <-c.done:
		return fmt.Errorf("gocraft: send on a closed client")
	}
}

func (c *Client) SetCompression(threshold int) {
	slog.Debug("enabled compression", "threshold", threshold)
	c.conn.SetThreshold(threshold)
}

func (c *Client) Close() error {
	c.closeOnce.Do(func() {
		slog.Debug("disconnecting")
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

	c.listeners.add(key, func(client *Client, packet Packet) error {
		return handler(client, packet.(P))
	})
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
