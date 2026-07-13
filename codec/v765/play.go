package v765

import gocraft "github.com/lrnxzz/go-craft"

type DeathLocation struct {
	DimensionName gocraft.Identifier
	Location      gocraft.Position
}

func (p DeathLocation) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.DimensionName, p.Location)
}

func (p *DeathLocation) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.DimensionName, &p.Location)
}

type JoinGame struct {
	EntityID            gocraft.Int
	Hardcore            gocraft.Bool
	Worlds              gocraft.Slice[gocraft.Identifier]
	MaxPlayers          gocraft.VarInt
	ViewDistance        gocraft.VarInt
	SimulationDistance  gocraft.VarInt
	ReducedDebugInfo    gocraft.Bool
	EnableRespawnScreen gocraft.Bool
	LimitedCrafting     gocraft.Bool
	DimensionType       gocraft.Identifier
	DimensionName       gocraft.Identifier
	HashedSeed          gocraft.Long
	GameMode            gocraft.UByte
	PreviousGameMode    gocraft.Byte
	Debug               gocraft.Bool
	Flat                gocraft.Bool
	Death               gocraft.Option[DeathLocation]
	PortalCooldown      gocraft.VarInt
}

func (*JoinGame) ID() int32 {
	return 0x29
}

func (p JoinGame) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.EntityID, p.Hardcore, p.Worlds, p.MaxPlayers, p.ViewDistance,
		p.SimulationDistance, p.ReducedDebugInfo, p.EnableRespawnScreen, p.LimitedCrafting,
		p.DimensionType, p.DimensionName, p.HashedSeed, p.GameMode, p.PreviousGameMode,
		p.Debug, p.Flat, p.Death, p.PortalCooldown)
}

func (p *JoinGame) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.EntityID, &p.Hardcore, &p.Worlds, &p.MaxPlayers, &p.ViewDistance,
		&p.SimulationDistance, &p.ReducedDebugInfo, &p.EnableRespawnScreen, &p.LimitedCrafting,
		&p.DimensionType, &p.DimensionName, &p.HashedSeed, &p.GameMode, &p.PreviousGameMode,
		&p.Debug, &p.Flat, &p.Death, &p.PortalCooldown)
}

type PlayKeepAlive struct {
	KeepAliveID gocraft.Long
}

func (*PlayKeepAlive) ID() int32 {
	return 0x24
}

func (p PlayKeepAlive) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.KeepAliveID)
}

func (p *PlayKeepAlive) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.KeepAliveID)
}

type PlayKeepAliveResponse struct {
	KeepAliveID gocraft.Long
}

func (*PlayKeepAliveResponse) ID() int32 {
	return 0x15
}

func (p PlayKeepAliveResponse) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.KeepAliveID)
}

func (p *PlayKeepAliveResponse) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.KeepAliveID)
}

type SyncPlayerPosition struct {
	X          gocraft.Double
	Y          gocraft.Double
	Z          gocraft.Double
	Yaw        gocraft.Float
	Pitch      gocraft.Float
	Flags      gocraft.Byte
	TeleportID gocraft.VarInt
}

func (*SyncPlayerPosition) ID() int32 {
	return 0x3E
}

func (p SyncPlayerPosition) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.X, p.Y, p.Z, p.Yaw, p.Pitch, p.Flags, p.TeleportID)
}

func (p *SyncPlayerPosition) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.X, &p.Y, &p.Z, &p.Yaw, &p.Pitch, &p.Flags, &p.TeleportID)
}

type ConfirmTeleport struct {
	TeleportID gocraft.VarInt
}

func (*ConfirmTeleport) ID() int32 {
	return 0x00
}

func (p ConfirmTeleport) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.TeleportID)
}

func (p *ConfirmTeleport) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.TeleportID)
}

type PlayDisconnect struct {
	Reason gocraft.NBT
}

func (*PlayDisconnect) ID() int32 {
	return 0x1B
}

func (p PlayDisconnect) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Reason)
}

func (p *PlayDisconnect) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Reason)
}
