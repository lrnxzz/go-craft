package v765

import gocraft "github.com/lrnxzz/go-craft"

const ProtocolVersion = 765

func Protocol() *gocraft.Protocol {
	proto := gocraft.NewProtocol()

	gocraft.Bind[Handshake](proto, gocraft.StateHandshaking, gocraft.Serverbound)

	gocraft.Bind[LoginStart](proto, gocraft.StateLogin, gocraft.Serverbound)
	gocraft.Bind[LoginAcknowledged](proto, gocraft.StateLogin, gocraft.Serverbound)

	gocraft.Bind[LoginDisconnect](proto, gocraft.StateLogin, gocraft.Clientbound)
	gocraft.Bind[LoginSuccess](proto, gocraft.StateLogin, gocraft.Clientbound)
	gocraft.Bind[SetCompression](proto, gocraft.StateLogin, gocraft.Clientbound)

	gocraft.Bind[ClientInformation](proto, gocraft.StateConfiguration, gocraft.Serverbound)
	gocraft.Bind[FinishConfiguration](proto, gocraft.StateConfiguration, gocraft.Serverbound)
	gocraft.Bind[ConfigKeepAlive](proto, gocraft.StateConfiguration, gocraft.Serverbound)
	gocraft.Bind[ConfigPong](proto, gocraft.StateConfiguration, gocraft.Serverbound)

	gocraft.Bind[ConfigDisconnect](proto, gocraft.StateConfiguration, gocraft.Clientbound)
	gocraft.Bind[FinishConfiguration](proto, gocraft.StateConfiguration, gocraft.Clientbound)
	gocraft.Bind[ConfigKeepAlive](proto, gocraft.StateConfiguration, gocraft.Clientbound)
	gocraft.Bind[ConfigPing](proto, gocraft.StateConfiguration, gocraft.Clientbound)
	gocraft.Bind[RegistryData](proto, gocraft.StateConfiguration, gocraft.Clientbound)
	gocraft.Bind[FeatureFlags](proto, gocraft.StateConfiguration, gocraft.Clientbound)

	gocraft.Bind[ConfirmTeleport](proto, gocraft.StatePlay, gocraft.Serverbound)
	gocraft.Bind[PlayKeepAliveResponse](proto, gocraft.StatePlay, gocraft.Serverbound)

	gocraft.Bind[JoinGame](proto, gocraft.StatePlay, gocraft.Clientbound)
	gocraft.Bind[PlayKeepAlive](proto, gocraft.StatePlay, gocraft.Clientbound)
	gocraft.Bind[SyncPlayerPosition](proto, gocraft.StatePlay, gocraft.Clientbound)
	gocraft.Bind[PlayDisconnect](proto, gocraft.StatePlay, gocraft.Clientbound)

	return proto
}
