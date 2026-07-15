package agent

import (
	"errors"
	"fmt"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765/blocks"
	"github.com/lrnxzz/go-craft/lib"
)

var errDigAbandoned = errors.New("agent: digging abandoned")

type excavator interface {
	StartDigging(gocraft.RayHit) error
	CancelDigging(gocraft.RayHit) error
	FinishDigging(gocraft.RayHit) error
}

type excavation struct {
	hit      gocraft.RayHit
	reach    float64
	progress float64
	future   *lib.Future[gocraft.RayHit]
}

type miner struct {
	digger excavator
	dig    *excavation
}

func (m *miner) begin(hit gocraft.RayHit, reach float64, mode gocraft.GameMode, held gocraft.ItemID) *lib.Future[gocraft.RayHit] {
	if err := m.abandon(); err != nil {
		return lib.FailedFuture[gocraft.RayHit](err)
	}

	if mode == gocraft.Creative {
		if err := m.digger.StartDigging(hit); err != nil {
			return lib.FailedFuture[gocraft.RayHit](err)
		}

		future := lib.NewFuture[gocraft.RayHit]()
		future.Complete(hit, nil)

		return future
	}

	damage, breakable := blocks.DigDamage(hit.State, held)
	if !breakable {
		return lib.FailedFuture[gocraft.RayHit](fmt.Errorf("agent: block state %d cannot be broken", hit.State))
	}

	if err := m.digger.StartDigging(hit); err != nil {
		return lib.FailedFuture[gocraft.RayHit](err)
	}

	future := lib.NewFuture[gocraft.RayHit]()

	if damage >= 1 {
		future.Complete(hit, m.digger.FinishDigging(hit))

		return future
	}

	m.dig = &excavation{
		hit:      hit,
		reach:    reach,
		progress: damage,
		future:   future,
	}

	return future
}

func (m *miner) abandon() error {
	dig := m.dig
	if dig == nil {
		return nil
	}

	m.dig = nil
	dig.future.Complete(gocraft.RayHit{}, errDigAbandoned)

	return m.digger.CancelDigging(dig.hit)
}

func (m *miner) excavating() (float64, bool) {
	if m.dig == nil {
		return 0, false
	}

	return m.dig.reach, true
}

func (m *miner) tick(target gocraft.RayHit, sighted bool, held gocraft.ItemID) error {
	dig := m.dig
	if dig == nil {
		return nil
	}

	if !sighted || target.Block != dig.hit.Block {
		return m.abandon()
	}

	damage, breakable := blocks.DigDamage(target.State, held)
	if !breakable {
		return m.abandon()
	}

	dig.progress += damage
	if dig.progress < 1 {
		return nil
	}

	m.dig = nil
	err := m.digger.FinishDigging(dig.hit)
	dig.future.Complete(dig.hit, err)

	return err
}
