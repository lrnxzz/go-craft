package gocraft

import (
	"context"
	"log/slog"
)

const outboundBuffer = 256

type sender struct {
	conn     *Conn
	outbound chan Packet
}

func newSender(conn *Conn) *sender {
	return &sender{
		conn:     conn,
		outbound: make(chan Packet, outboundBuffer),
	}
}

func (s *sender) loop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case packet := <-s.outbound:
			if err := s.conn.WriteFrame(EncodeFrame(packet)); err != nil {
				return err
			}

			slog.Debug("sent", "packet", packet.Name())
		}
	}
}
