package gocraft

import (
	"context"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
)

type ticker struct {
	rate time.Duration
	fn   func()
}

type inbound struct {
	state  State
	packet Packet
}

func (c *Client) Tick(rate time.Duration, fn func()) {
	c.tick = ticker{rate: rate, fn: fn}
}

func (c *Client) Run(parent context.Context) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	stop := context.AfterFunc(ctx, func() {
		c.Close()
	})
	defer stop()

	packets := make(chan inbound, 256)

	var group errgroup.Group

	group.Go(func() error {
		defer cancel()

		return c.read(ctx, packets)
	})
	group.Go(func() error {
		defer cancel()

		return c.loop(ctx, packets)
	})
	group.Go(func() error {
		defer cancel()

		return c.sender.loop(ctx)
	})

	return group.Wait()
}

func (c *Client) read(ctx context.Context, packets chan<- inbound) error {
	for {
		frame, err := c.conn.ReadFrame()
		if err != nil {
			return c.exit(ctx, err)
		}

		state := c.State()
		packet, known, err := c.protocol.Decode(state, Clientbound, frame)
		if err != nil {
			return err
		}
		if !known {
			slog.Debug("skipped unknown packet", "id", frame.ID)
			continue
		}

		select {
		case packets <- inbound{state: state, packet: packet}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Client) loop(ctx context.Context, packets <-chan inbound) error {
	var pulse <-chan time.Time
	if c.tick.fn != nil {
		t := time.NewTicker(c.tick.rate)
		defer t.Stop()
		pulse = t.C
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case in := <-packets:
			slog.Debug("received", "packet", in.packet.Name())
			if err := c.listeners.dispatch(c, in.state, in.packet); err != nil {
				return err
			}
		case <-pulse:
			c.tick.fn()
		}
	}
}
