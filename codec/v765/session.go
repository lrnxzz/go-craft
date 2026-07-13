package v765

import (
	"context"
	"encoding/hex"
	"fmt"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/mojang"
)

type JoinHandler func(*gocraft.Client, *JoinGame) error

func Join(client *gocraft.Client, host string, port uint16, username string, onReady JoinHandler) error {
	offline := mojang.Offline{
		Username: username,
	}

	session, err := offline.Authenticate(context.Background())
	if err != nil {
		return err
	}

	var uuid gocraft.UUID
	if raw, err := hex.DecodeString(session.Profile.ID); err == nil {
		copy(uuid[:], raw)
	}

	installLogin(client)
	installConfiguration(client)
	installPlay(client, onReady)

	handshake := &Handshake{
		ProtocolVersion: ProtocolVersion,
		ServerAddress:   gocraft.String(host),
		ServerPort:      gocraft.UShort(port),
		NextState:       gocraft.VarInt(gocraft.StateLogin),
	}
	if err := client.Send(handshake); err != nil {
		return err
	}

	client.SetState(gocraft.StateLogin)

	start := &LoginStart{
		Username: gocraft.String(username),
		UUID:     uuid,
	}

	return client.Send(start)
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

func installPlay(client *gocraft.Client, onReady JoinHandler) {
	gocraft.On[*PlayKeepAlive](client, gocraft.StatePlay, onPlayKeepAlive)
	gocraft.On[*SyncPlayerPosition](client, gocraft.StatePlay, onSyncPlayerPosition)
	gocraft.On[*JoinGame](client, gocraft.StatePlay, onReady)
	gocraft.On[*PlayDisconnect](client, gocraft.StatePlay, onPlayDisconnect)
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
	reply := &ConfigKeepAlive{
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

func onSyncPlayerPosition(c *gocraft.Client, p *SyncPlayerPosition) error {
	confirm := &ConfirmTeleport{
		TeleportID: p.TeleportID,
	}

	return c.Send(confirm)
}

func onPlayDisconnect(c *gocraft.Client, p *PlayDisconnect) error {
	return fmt.Errorf("v765: kicked during play")
}
