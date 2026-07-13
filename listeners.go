package gocraft

type handlerKey struct {
	state State
	id    int32
}

type packetHandler func(*Client, Packet) error

type listeners map[handlerKey][]packetHandler

func (l listeners) add(key handlerKey, handler packetHandler) {
	l[key] = append(l[key], handler)
}

func (l listeners) dispatch(c *Client, state State, packet Packet) error {
	key := handlerKey{
		state: state,
		id:    packet.ID(),
	}

	for _, handler := range l[key] {
		if err := handler(c, packet); err != nil {
			return err
		}
	}

	return nil
}
