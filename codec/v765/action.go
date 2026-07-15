package v765

import gocraft "github.com/lrnxzz/go-craft"

const (
	digStart  gocraft.VarInt = 0
	digCancel gocraft.VarInt = 1
	digFinish gocraft.VarInt = 2

	mainHand gocraft.VarInt = 0
)

type PlayerAction struct {
	Status   gocraft.VarInt
	Location gocraft.Position
	Face     gocraft.Byte
	Sequence gocraft.VarInt
}

func (*PlayerAction) ID() int32 {
	return 0x21
}

func (*PlayerAction) Name() string {
	return "PlayerAction"
}

func (*PlayerAction) State() gocraft.State {
	return gocraft.StatePlay
}

func (*PlayerAction) Direction() gocraft.Direction {
	return gocraft.Serverbound
}

func (p PlayerAction) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Status, p.Location, p.Face, p.Sequence)
}

func (p *PlayerAction) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Status, &p.Location, &p.Face, &p.Sequence)
}

type UseItemOn struct {
	Hand        gocraft.VarInt
	Location    gocraft.Position
	Face        gocraft.VarInt
	CursorX     gocraft.Float
	CursorY     gocraft.Float
	CursorZ     gocraft.Float
	InsideBlock gocraft.Bool
	Sequence    gocraft.VarInt
}

func (*UseItemOn) ID() int32 {
	return 0x35
}

func (*UseItemOn) Name() string {
	return "UseItemOn"
}

func (*UseItemOn) State() gocraft.State {
	return gocraft.StatePlay
}

func (*UseItemOn) Direction() gocraft.Direction {
	return gocraft.Serverbound
}

func (p UseItemOn) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Hand, p.Location, p.Face,
		p.CursorX, p.CursorY, p.CursorZ, p.InsideBlock, p.Sequence)
}

func (p *UseItemOn) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Hand, &p.Location, &p.Face,
		&p.CursorX, &p.CursorY, &p.CursorZ, &p.InsideBlock, &p.Sequence)
}

type AcknowledgeBlockChange struct {
	Sequence gocraft.VarInt
}

func (*AcknowledgeBlockChange) ID() int32 {
	return 0x05
}

func (*AcknowledgeBlockChange) Name() string {
	return "AcknowledgeBlockChange"
}

func (*AcknowledgeBlockChange) State() gocraft.State {
	return gocraft.StatePlay
}

func (*AcknowledgeBlockChange) Direction() gocraft.Direction {
	return gocraft.Clientbound
}

func (p AcknowledgeBlockChange) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Sequence)
}

func (p *AcknowledgeBlockChange) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Sequence)
}
