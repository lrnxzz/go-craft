package v765

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/mojang"
	"github.com/lrnxzz/go-craft/nbt"
)

type dimensionBounds struct {
	minY   int
	height int
}

var overworld = dimensionBounds{
	minY:   -64,
	height: 384,
}

type JoinHandler func(*gocraft.Client, *JoinGame) error

type Session struct {
	client     *gocraft.Client
	world      *gocraft.World
	player     *gocraft.Player
	ready      JoinHandler
	spawned    bool
	dimensions map[gocraft.Identifier]dimensionBounds
	bounds     dimensionBounds
}

func Join(client *gocraft.Client, host string, port uint16, username string, onReady JoinHandler) (*Session, error) {
	offline := mojang.Offline{Username: username}

	profile, err := offline.Authenticate(context.Background())
	if err != nil {
		return nil, err
	}

	var uuid gocraft.UUID
	if raw, err := hex.DecodeString(profile.Profile.ID); err == nil {
		copy(uuid[:], raw)
	}

	session := &Session{
		client:     client,
		world:      gocraft.NewWorld(),
		player:     &gocraft.Player{},
		ready:      onReady,
		dimensions: map[gocraft.Identifier]dimensionBounds{},
		bounds:     overworld,
	}
	session.listen()

	if err := client.Send(&Handshake{
		ProtocolVersion: ProtocolVersion,
		ServerAddress:   gocraft.String(host),
		ServerPort:      gocraft.UShort(port),
		NextState:       gocraft.VarInt(gocraft.StateLogin),
	}); err != nil {
		return nil, err
	}

	client.SetState(gocraft.StateLogin)

	return session, client.Send(&LoginStart{
		Username: gocraft.String(username),
		UUID:     uuid,
	})
}

func (s *Session) World() *gocraft.World {
	return s.world
}

func (s *Session) Player() *gocraft.Player {
	return s.player
}

func (s *Session) Spawned() bool {
	return s.spawned
}

func (s *Session) listen() {
	gocraft.On(s.client, s.onCompression)
	gocraft.On(s.client, s.onEncryption)
	gocraft.On(s.client, s.onLoginSuccess)
	gocraft.On(s.client, s.onLoginDisconnect)

	gocraft.On(s.client, s.onConfigKeepAlive)
	gocraft.On(s.client, s.onConfigPing)
	gocraft.On(s.client, s.onRegistryData)
	gocraft.On(s.client, s.onFinishConfiguration)
	gocraft.On(s.client, s.onConfigDisconnect)

	gocraft.On(s.client, s.onJoinGame)
	gocraft.On(s.client, s.onKeepAlive)
	gocraft.On(s.client, s.onSyncPosition)
	gocraft.On(s.client, s.onChunkData)
	gocraft.On(s.client, s.onUnloadChunk)
	gocraft.On(s.client, s.onBlockUpdate)
	gocraft.On(s.client, s.onSectionBlocks)
	gocraft.On(s.client, s.onHealth)
	gocraft.On(s.client, s.onAbilities)
	gocraft.On(s.client, s.onExperience)
	gocraft.On(s.client, s.onPlayDisconnect)
}

func (s *Session) onRegistryData(c *gocraft.Client, p *RegistryData) error {
	registry, ok := nbt.Get[nbt.Compound](nbt.Compound(p.Codec), "minecraft:dimension_type")
	if !ok {
		return nil
	}

	entries, ok := nbt.Get[nbt.List](registry, "value")
	if !ok {
		return nil
	}

	types, ok := nbt.Items[nbt.Compound](entries)
	if !ok {
		return nil
	}

	for _, entry := range types {
		name, ok := nbt.Get[nbt.String](entry, "name")
		if !ok {
			continue
		}

		element, ok := nbt.Get[nbt.Compound](entry, "element")
		if !ok {
			continue
		}

		minY, ok := nbt.Get[nbt.Int](element, "min_y")
		if !ok {
			continue
		}

		height, ok := nbt.Get[nbt.Int](element, "height")
		if !ok {
			continue
		}

		s.dimensions[gocraft.Identifier(name)] = dimensionBounds{
			minY:   int(minY),
			height: int(height),
		}
	}

	return nil
}

func (s *Session) onCompression(c *gocraft.Client, p *SetCompression) error {
	c.SetCompression(p.Threshold.Int())

	return nil
}

func (s *Session) onEncryption(c *gocraft.Client, p *EncryptionBegin) error {
	return errors.New("v765: server requested encryption (online-mode); auth and encryption are not implemented")
}

func (s *Session) onLoginSuccess(c *gocraft.Client, p *LoginSuccess) error {
	p.Apply(s.player)

	if err := c.Send(&LoginAcknowledged{}); err != nil {
		return err
	}

	c.SetState(gocraft.StateConfiguration)

	return c.Send(&ClientInformation{
		Locale:              "en_us",
		ViewDistance:        8,
		MainHand:            1,
		EnableServerListing: true,
	})
}

func (s *Session) onLoginDisconnect(c *gocraft.Client, p *LoginDisconnect) error {
	return fmt.Errorf("v765: kicked during login: %s", p.Reason)
}

func (s *Session) onConfigKeepAlive(c *gocraft.Client, p *ConfigKeepAlive) error {
	return c.Send(&ConfigKeepAliveResponse{KeepAliveID: p.KeepAliveID})
}

func (s *Session) onConfigPing(c *gocraft.Client, p *ConfigPing) error {
	return c.Send(&ConfigPong{PingID: p.PingID})
}

func (s *Session) onFinishConfiguration(c *gocraft.Client, p *FinishConfiguration) error {
	if err := c.Send(&AcknowledgeConfiguration{}); err != nil {
		return err
	}

	c.SetState(gocraft.StatePlay)

	return nil
}

func (s *Session) onConfigDisconnect(c *gocraft.Client, p *ConfigDisconnect) error {
	return errors.New("v765: kicked during configuration")
}

func (s *Session) onJoinGame(c *gocraft.Client, p *JoinGame) error {
	p.Apply(s.player)

	bounds, ok := s.dimensions[p.DimensionType]
	if !ok {
		bounds = overworld
	}
	s.bounds = bounds

	if s.ready != nil {
		return s.ready(c, p)
	}

	return nil
}

func (s *Session) onKeepAlive(c *gocraft.Client, p *PlayKeepAlive) error {
	return c.Send(&PlayKeepAliveResponse{KeepAliveID: p.KeepAliveID})
}

func (s *Session) onSyncPosition(c *gocraft.Client, p *SyncPlayerPosition) error {
	p.Apply(s.player)
	s.spawned = true

	if err := c.Send(&ConfirmTeleport{TeleportID: p.TeleportID}); err != nil {
		return err
	}

	return s.SendPosition()
}

func (s *Session) SendPosition() error {
	return s.client.Send(&SetPlayerPositionRotation{
		X:        gocraft.Double(s.player.Position.X),
		Y:        gocraft.Double(s.player.Position.Y),
		Z:        gocraft.Double(s.player.Position.Z),
		Yaw:      gocraft.Float(s.player.Yaw),
		Pitch:    gocraft.Float(s.player.Pitch),
		OnGround: gocraft.Bool(s.player.OnGround),
	})
}

func (s *Session) onChunkData(c *gocraft.Client, p *ChunkData) error {
	column, err := p.Column(s.bounds.minY, s.bounds.height)
	if err != nil {
		return err
	}

	s.world.LoadColumn(column)

	return nil
}

func (s *Session) onUnloadChunk(c *gocraft.Client, p *UnloadChunk) error {
	s.world.UnloadColumn(p.X.Int32(), p.Z.Int32())

	return nil
}

func (s *Session) onBlockUpdate(c *gocraft.Client, p *BlockUpdate) error {
	b := p.Change()
	s.world.SetBlock(b.X, b.Y, b.Z, b.State)

	return nil
}

func (s *Session) onSectionBlocks(c *gocraft.Client, p *SectionBlocksUpdate) error {
	for _, b := range p.Changes() {
		s.world.SetBlock(b.X, b.Y, b.Z, b.State)
	}

	return nil
}

func (s *Session) onHealth(c *gocraft.Client, p *SetHealth) error {
	p.Apply(s.player)

	return nil
}

func (s *Session) onAbilities(c *gocraft.Client, p *PlayerAbilities) error {
	p.Apply(s.player)

	return nil
}

func (s *Session) onExperience(c *gocraft.Client, p *SetExperience) error {
	p.Apply(s.player)

	return nil
}

func (s *Session) onPlayDisconnect(c *gocraft.Client, p *PlayDisconnect) error {
	return errors.New("v765: kicked during play")
}
