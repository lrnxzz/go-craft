package v765

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/mojang"
)

const (
	overworldMinY   = -64
	overworldHeight = 384
)

type JoinHandler func(*gocraft.Client, *JoinGame) error

type Session struct {
	client *gocraft.Client
	world  *gocraft.World
	player *gocraft.Player
	ready  JoinHandler
}

func (s *Session) World() *gocraft.World {
	return s.world
}

func (s *Session) Player() *gocraft.Player {
	return s.player
}

func Join(client *gocraft.Client, host string, port uint16, username string, onReady JoinHandler) (*Session, error) {
	offline := mojang.Offline{
		Username: username,
	}

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
	session.install()

	handshake := &Handshake{
		ProtocolVersion: ProtocolVersion,
		ServerAddress:   gocraft.String(host),
		ServerPort:      gocraft.UShort(port),
		NextState:       gocraft.VarInt(gocraft.StateLogin),
	}
	if err := client.Send(handshake); err != nil {
		return nil, err
	}

	client.SetState(gocraft.StateLogin)

	start := &LoginStart{
		Username: gocraft.String(username),
		UUID:     uuid,
	}
	if err := client.Send(start); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Session) install() {
	installLogin(s.client)
	installConfiguration(s.client)

	gocraft.On[*JoinGame](s.client, gocraft.StatePlay, s.onJoinGame)
	gocraft.On[*PlayKeepAlive](s.client, gocraft.StatePlay, onPlayKeepAlive)
	gocraft.On[*SyncPlayerPosition](s.client, gocraft.StatePlay, s.onSyncPlayerPosition)
	gocraft.On[*ChunkData](s.client, gocraft.StatePlay, s.onChunkData)
	gocraft.On[*UnloadChunk](s.client, gocraft.StatePlay, s.onUnloadChunk)
	gocraft.On[*BlockUpdate](s.client, gocraft.StatePlay, s.onBlockUpdate)
	gocraft.On[*SectionBlocksUpdate](s.client, gocraft.StatePlay, s.onSectionBlocksUpdate)
	gocraft.On[*SetHealth](s.client, gocraft.StatePlay, s.onSetHealth)
	gocraft.On[*PlayDisconnect](s.client, gocraft.StatePlay, onPlayDisconnect)
}

func installLogin(client *gocraft.Client) {
	gocraft.On[*SetCompression](client, gocraft.StateLogin, onSetCompression)
	gocraft.On[*EncryptionBegin](client, gocraft.StateLogin, onEncryptionBegin)
	gocraft.On[*LoginSuccess](client, gocraft.StateLogin, onLoginSuccess)
	gocraft.On[*LoginDisconnect](client, gocraft.StateLogin, onLoginDisconnect)
}

func installConfiguration(client *gocraft.Client) {
	gocraft.On[*ConfigKeepAlive](client, gocraft.StateConfiguration, onConfigKeepAlive)
	gocraft.On[*ConfigPing](client, gocraft.StateConfiguration, onConfigPing)
	gocraft.On[*FinishConfiguration](client, gocraft.StateConfiguration, onFinishConfiguration)
	gocraft.On[*ConfigDisconnect](client, gocraft.StateConfiguration, onConfigDisconnect)
}

func onSetCompression(c *gocraft.Client, p *SetCompression) error {
	c.SetCompression(int(p.Threshold))

	return nil
}

func onEncryptionBegin(c *gocraft.Client, p *EncryptionBegin) error {
	return fmt.Errorf("v765: server requested encryption (online-mode) — auth and encryption are not implemented yet")
}

func onLoginSuccess(c *gocraft.Client, p *LoginSuccess) error {
	ack := &LoginAcknowledged{}
	if err := c.Send(ack); err != nil {
		return err
	}

	c.SetState(gocraft.StateConfiguration)

	info := &ClientInformation{
		Locale:              "en_us",
		ViewDistance:        8,
		MainHand:            1,
		EnableServerListing: true,
	}

	return c.Send(info)
}

func onLoginDisconnect(c *gocraft.Client, p *LoginDisconnect) error {
	return fmt.Errorf("v765: kicked during login: %s", p.Reason)
}

func onConfigKeepAlive(c *gocraft.Client, p *ConfigKeepAlive) error {
	reply := &ConfigKeepAliveResponse{
		KeepAliveID: p.KeepAliveID,
	}

	return c.Send(reply)
}

func onConfigPing(c *gocraft.Client, p *ConfigPing) error {
	pong := &ConfigPong{
		PingID: p.PingID,
	}

	return c.Send(pong)
}

func onFinishConfiguration(c *gocraft.Client, p *FinishConfiguration) error {
	done := &FinishConfiguration{}
	if err := c.Send(done); err != nil {
		return err
	}

	c.SetState(gocraft.StatePlay)

	return nil
}

func onConfigDisconnect(c *gocraft.Client, p *ConfigDisconnect) error {
	return fmt.Errorf("v765: kicked during configuration")
}

func onPlayKeepAlive(c *gocraft.Client, p *PlayKeepAlive) error {
	reply := &PlayKeepAliveResponse{
		KeepAliveID: p.KeepAliveID,
	}

	return c.Send(reply)
}

func onPlayDisconnect(c *gocraft.Client, p *PlayDisconnect) error {
	return fmt.Errorf("v765: kicked during play")
}

func (s *Session) onJoinGame(c *gocraft.Client, p *JoinGame) error {
	s.player.GameMode = gocraft.GameMode(p.GameMode)

	if s.ready != nil {
		return s.ready(c, p)
	}

	return nil
}

func (s *Session) onSyncPlayerPosition(c *gocraft.Client, p *SyncPlayerPosition) error {
	flags := byte(p.Flags)
	s.player.Position.X = resolve(s.player.Position.X, float64(p.X), flags&0x01)
	s.player.Position.Y = resolve(s.player.Position.Y, float64(p.Y), flags&0x02)
	s.player.Position.Z = resolve(s.player.Position.Z, float64(p.Z), flags&0x04)
	s.player.Yaw = resolve(s.player.Yaw, float32(p.Yaw), flags&0x08)
	s.player.Pitch = resolve(s.player.Pitch, float32(p.Pitch), flags&0x10)

	slog.Debug("teleported", "position", s.player.Position)

	confirm := &ConfirmTeleport{
		TeleportID: p.TeleportID,
	}
	if err := c.Send(confirm); err != nil {
		return err
	}

	reply := &SetPlayerPositionRotation{
		X:        gocraft.Double(s.player.Position.X),
		Y:        gocraft.Double(s.player.Position.Y),
		Z:        gocraft.Double(s.player.Position.Z),
		Yaw:      gocraft.Float(s.player.Yaw),
		Pitch:    gocraft.Float(s.player.Pitch),
		OnGround: gocraft.Bool(s.player.OnGround),
	}

	return c.Send(reply)
}

func resolve[T ~float32 | ~float64](current, value T, relative byte) T {
	if relative != 0 {
		return current + value
	}

	return value
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
	s.world.SetBlock(int(p.Location.X), int(p.Location.Y), int(p.Location.Z), gocraft.BlockState(p.State))

	return nil
}

func (s *Session) onSectionBlocksUpdate(c *gocraft.Client, p *SectionBlocksUpdate) error {
	baseX := int(p.Section>>42) * 16
	baseZ := int(p.Section<<22>>42) * 16
	baseY := int(p.Section<<44>>44) * 16

	for _, record := range p.Records {
		value := int64(record)
		state := gocraft.BlockState(value >> 12)
		s.world.SetBlock(baseX+int(value>>8&0xf), baseY+int(value&0xf), baseZ+int(value>>4&0xf), state)
	}

	return nil
}

func (s *Session) onSetHealth(c *gocraft.Client, p *SetHealth) error {
	s.player.Health = float32(p.Health)
	s.player.Food = int32(p.Food)

	return nil
}
