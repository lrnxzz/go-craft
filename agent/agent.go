package agent

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765"
	"github.com/lrnxzz/go-craft/codec/v765/blocks"
)

const tickRate = 50 * time.Millisecond

type Agent struct {
	client  *gocraft.Client
	session *v765.Session
	physics *gocraft.Physics
}

func Join(ctx context.Context, address, username string) (*Agent, error) {
	host, port, err := splitAddress(address)
	if err != nil {
		return nil, err
	}

	conn, err := gocraft.Dial(ctx, net.JoinHostPort(host, strconv.Itoa(int(port))))
	if err != nil {
		return nil, err
	}

	client := gocraft.NewClient(conn, v765.Protocol())
	a := &Agent{client: client, physics: gocraft.NewPhysics(blocks.Collision)}

	session, err := v765.Join(client, host, port, username, nil)
	if err != nil {
		client.Close()

		return nil, err
	}
	a.session = session

	client.Tick(tickRate, a.tick)

	return a, nil
}

func (a *Agent) Run(ctx context.Context) error {
	return a.client.Run(ctx)
}

func (a *Agent) World() *gocraft.World {
	return a.session.World()
}

func (a *Agent) Player() *gocraft.Player {
	return a.session.Player()
}

func (a *Agent) tick() {
	if !a.session.Spawned() {
		return
	}

	a.physics.Tick(a.session.World(), a.session.Player())
	_ = a.session.SendPosition()
}

func splitAddress(address string) (string, uint16, error) {
	host, raw, err := net.SplitHostPort(address)
	if err != nil {
		return address, 25565, nil
	}

	port, err := strconv.ParseUint(raw, 10, 16)
	if err != nil {
		return "", 0, fmt.Errorf("agent: invalid port in %q", address)
	}

	return host, uint16(port), nil
}
