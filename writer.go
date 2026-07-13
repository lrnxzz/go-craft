package gocraft

import "context"

const outboundBuffer = 256

type writer struct {
	conn     *Conn
	outbound chan Packet
}

func newWriter(conn *Conn) *writer {
	return &writer{
		conn:     conn,
		outbound: make(chan Packet, outboundBuffer),
	}
}

func (w *writer) loop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case packet := <-w.outbound:
			if err := w.conn.WriteFrame(EncodeFrame(packet)); err != nil {
				return err
			}
		}
	}
}
