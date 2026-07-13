package v765

import gocraft "github.com/lrnxzz/go-craft"

type Handshake struct {
	ProtocolVersion gocraft.VarInt
	ServerAddress   gocraft.String
	ServerPort      gocraft.UShort
	NextState       gocraft.VarInt
}

func (*Handshake) ID() int32 {
	return 0x00
}

func (*Handshake) Name() string {
	return "Handshake"
}

func (p Handshake) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.ProtocolVersion, p.ServerAddress, p.ServerPort, p.NextState)
}

func (p *Handshake) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.ProtocolVersion, &p.ServerAddress, &p.ServerPort, &p.NextState)
}
