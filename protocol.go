package gocraft

import "fmt"

type State uint8

const (
	StateHandshaking   State = 0
	StateStatus        State = 1
	StateLogin         State = 2
	StateConfiguration State = 3
	StatePlay          State = 4
)

func (s State) String() string {
	switch s {
	case StateHandshaking:
		return "handshaking"
	case StateStatus:
		return "status"
	case StateLogin:
		return "login"
	case StateConfiguration:
		return "configuration"
	case StatePlay:
		return "play"
	}

	return fmt.Sprintf("state(%d)", uint8(s))
}

type Direction uint8

const (
	Serverbound Direction = iota
	Clientbound
)

func (d Direction) String() string {
	switch d {
	case Serverbound:
		return "serverbound"
	case Clientbound:
		return "clientbound"
	}

	return fmt.Sprintf("direction(%d)", uint8(d))
}

type Packet interface {
	ID() int32
	Append(dst []byte) []byte
	Decode(r *Reader) error
}

type packetKey struct {
	state State
	dir   Direction
	id    int32
}

type Protocol struct {
	factories map[packetKey]func() Packet
}

func NewProtocol() *Protocol {
	return &Protocol{factories: make(map[packetKey]func() Packet)}
}

func Bind[Value any, Ptr interface {
	*Value
	Packet
}](proto *Protocol, state State, dir Direction) {
	factory := func() Packet {
		return Ptr(new(Value))
	}

	key := packetKey{state: state, dir: dir, id: factory().ID()}
	proto.factories[key] = factory
}

func (proto *Protocol) New(state State, dir Direction, id int32) (Packet, bool) {
	key := packetKey{state: state, dir: dir, id: id}

	factory, ok := proto.factories[key]
	if !ok {
		return nil, false
	}

	return factory(), true
}

func (proto *Protocol) Decode(state State, dir Direction, frame Frame) (Packet, bool, error) {
	packet, ok := proto.New(state, dir, int32(frame.ID))
	if !ok {
		return nil, false, nil
	}

	if err := packet.Decode(NewReader(frame.Payload)); err != nil {
		return nil, true, err
	}

	return packet, true, nil
}

func EncodeFrame(packet Packet) Frame {
	return Frame{ID: VarInt(packet.ID()), Payload: packet.Append(nil)}
}
