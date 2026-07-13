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
	session, err := mojang.Offline{Username: username}.Authenticate(context.Background())
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

	return client.Send(&LoginStart{Username: gocraft.String(username), UUID: uuid})
}

func installLogin(client *gocraft.Client) {
	gocraft.On[SetCompression](client, gocraft.StateLogin, func(c *gocraft.Client, p *SetCompression) error {
		c.SetCompression(int(p.Threshold))

		return nil
	})

	gocraft.On[EncryptionBegin](client, gocraft.StateLogin, func(c *gocraft.Client, p *EncryptionBegin) error {
		return fmt.Errorf("v765: server requested encryption (online-mode) — auth and encryption are not implemented yet")
	})

	gocraft.On[LoginSuccess](client, gocraft.StateLogin, func(c *gocraft.Client, p *LoginSuccess) error {
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
	})

	gocraft.On[LoginDisconnect](client, gocraft.StateLogin, func(c *gocraft.Client, p *LoginDisconnect) error {
		return fmt.Errorf("v765: kicked during login: %s", p.Reason)
	})
}

func installConfiguration(client *gocraft.Client) {
	gocraft.On[ConfigKeepAlive](client, gocraft.StateConfiguration, func(c *gocraft.Client, p *ConfigKeepAlive) error {
		return c.Send(&ConfigKeepAlive{KeepAliveID: p.KeepAliveID})
	})

	gocraft.On[FinishConfiguration](client, gocraft.StateConfiguration, func(c *gocraft.Client, p *FinishConfiguration) error {
		if err := c.Send(&FinishConfiguration{}); err != nil {
			return err
		}

		c.SetState(gocraft.StatePlay)

		return nil
	})

	gocraft.On[ConfigDisconnect](client, gocraft.StateConfiguration, func(c *gocraft.Client, p *ConfigDisconnect) error {
		return fmt.Errorf("v765: kicked during configuration")
	})
}

func installPlay(client *gocraft.Client, onReady JoinHandler) {
	gocraft.On[PlayKeepAlive](client, gocraft.StatePlay, func(c *gocraft.Client, p *PlayKeepAlive) error {
		return c.Send(&PlayKeepAliveResponse{KeepAliveID: p.KeepAliveID})
	})

	gocraft.On[SyncPlayerPosition](client, gocraft.StatePlay, func(c *gocraft.Client, p *SyncPlayerPosition) error {
		return c.Send(&ConfirmTeleport{TeleportID: p.TeleportID})
	})

	gocraft.On[JoinGame](client, gocraft.StatePlay, onReady)

	gocraft.On[PlayDisconnect](client, gocraft.StatePlay, func(c *gocraft.Client, p *PlayDisconnect) error {
		return fmt.Errorf("v765: kicked during play")
	})
}
