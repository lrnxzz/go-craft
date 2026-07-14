package v765

import (
	"context"
	"encoding/hex"
	"fmt"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/mojang"
)

const (
	overworldMinY   = -64
	overworldHeight = 384
)

type JoinHandler func(*gocraft.Client, *JoinGame) error

type Session struct {
	client  *gocraft.Client
	world   *gocraft.World
	player  *gocraft.Player
	ready   JoinHandler
	spawned bool
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
		client: client,
		world:  gocraft.NewWorld(),
		player: &gocraft.Player{},
		ready:  onReady,
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
	gocraft.On(s.client, gocraft.StateLogin, s.onCompression)
	gocraft.On(s.client, gocraft.StateLogin, s.onEncryption)
	gocraft.On(s.client, gocraft.StateLogin, s.onLoginSuccess)
	gocraft.On(s.client, gocraft.StateLogin, s.onLoginDisconnect)

	gocraft.On(s.client, gocraft.StateConfiguration, s.onConfigKeepAlive)
	gocraft.On(s.client, gocraft.StateConfiguration, s.onConfigPing)
	gocraft.On(s.client, gocraft.StateConfiguration, s.onFinishConfiguration)
	gocraft.On(s.client, gocraft.StateConfiguration, s.onConfigDisconnect)

	gocraft.On(s.client, gocraft.StatePlay, s.onJoinGame)
	gocraft.On(s.client, gocraft.StatePlay, s.onKeepAlive)
	gocraft.On(s.client, gocraft.StatePlay, s.onSyncPosition)
	gocraft.On(s.client, gocraft.StatePlay, s.onChunkData)
	gocraft.On(s.client, gocraft.StatePlay, s.onUnloadChunk)
	gocraft.On(s.client, gocraft.StatePlay, s.onBlockUpdate)
	gocraft.On(s.client, gocraft.StatePlay, s.onSectionBlocks)
	gocraft.On(s.client, gocraft.StatePlay, s.onHealth)
	gocraft.On(s.client, gocraft.StatePlay, s.onPlayDisconnect)
}

func (s *Session) onCompression(c *gocraft.Client, p *SetCompression) error {
	c.SetCompression(int(p.Threshold))

	return nil
}

func (s *Session) onEncryption(c *gocraft.Client, p *EncryptionBegin) error {
	return fmt.Errorf("v765: server requested encryption (online-mode); auth and encryption are not implemented")
}

func (s *Session) onLoginSuccess(c *gocraft.Client, p *LoginSuccess) error {
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
	if err := c.Send(&FinishConfiguration{}); err != nil {
		return err
	}

	c.SetState(gocraft.StatePlay)

	return nil
}

func (s *Session) onConfigDisconnect(c *gocraft.Client, p *ConfigDisconnect) error {
	return fmt.Errorf("v765: kicked during configuration")
}

func (s *Session) onJoinGame(c *gocraft.Client, p *JoinGame) error {
	s.player.GameMode = gocraft.GameMode(p.GameMode)

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
	column := gocraft.NewChunkColumn(int32(p.X), int32(p.Z), overworldMinY, overworldHeight)
	if err := column.Decode(gocraft.NewReader(p.Sections)); err != nil {
		return err
	}

	s.world.LoadColumn(column)

	return nil
}

func (s *Session) onUnloadChunk(c *gocraft.Client, p *UnloadChunk) error {
	s.world.UnloadColumn(int32(p.X), int32(p.Z))

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
	s.player.Health = float32(p.Health)
	s.player.Food = int32(p.Food)

	return nil
}

func (s *Session) onPlayDisconnect(c *gocraft.Client, p *PlayDisconnect) error {
	return fmt.Errorf("v765: kicked during play")
}
