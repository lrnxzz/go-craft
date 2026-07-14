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

func (*JoinGame) Name() string {
	return "JoinGame"
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

func (*PlayKeepAlive) Name() string {
	return "PlayKeepAlive"
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

func (*PlayKeepAliveResponse) Name() string {
	return "PlayKeepAliveResponse"
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

func (*SyncPlayerPosition) Name() string {
	return "SyncPlayerPosition"
}

func (p SyncPlayerPosition) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.X, p.Y, p.Z, p.Yaw, p.Pitch, p.Flags, p.TeleportID)
}

func (p *SyncPlayerPosition) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.X, &p.Y, &p.Z, &p.Yaw, &p.Pitch, &p.Flags, &p.TeleportID)
}

const (
	relativeX byte = 1 << iota
	relativeY
	relativeZ
	relativeYaw
	relativePitch
)

func (p *SyncPlayerPosition) Apply(player *gocraft.Player) {
	flags := byte(p.Flags)

	target := gocraft.Vec3d{X: float64(p.X), Y: float64(p.Y), Z: float64(p.Z)}
	if flags&relativeX != 0 {
		target.X += player.Position.X
	}
	if flags&relativeY != 0 {
		target.Y += player.Position.Y
	}
	if flags&relativeZ != 0 {
		target.Z += player.Position.Z
	}

	yaw, pitch := float32(p.Yaw), float32(p.Pitch)
	if flags&relativeYaw != 0 {
		yaw += player.Yaw
	}
	if flags&relativePitch != 0 {
		pitch += player.Pitch
	}

	player.Position = target
	player.Yaw = yaw
	player.Pitch = pitch
}

type ConfirmTeleport struct {
	TeleportID gocraft.VarInt
}

func (*ConfirmTeleport) ID() int32 {
	return 0x00
}

func (*ConfirmTeleport) Name() string {
	return "ConfirmTeleport"
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

func (*PlayDisconnect) Name() string {
	return "PlayDisconnect"
}

func (p PlayDisconnect) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Reason)
}

func (p *PlayDisconnect) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Reason)
}

type ChunkData struct {
	X          gocraft.Int
	Z          gocraft.Int
	Heightmaps gocraft.NBT
	Sections   gocraft.Bytes
}

func (*ChunkData) ID() int32 {
	return 0x25
}

func (*ChunkData) Name() string {
	return "ChunkData"
}

func (p ChunkData) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.X, p.Z, p.Heightmaps, p.Sections)
}

func (p *ChunkData) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.X, &p.Z, &p.Heightmaps, &p.Sections)
}

type UnloadChunk struct {
	Z gocraft.Int
	X gocraft.Int
}

func (*UnloadChunk) ID() int32 {
	return 0x1F
}

func (*UnloadChunk) Name() string {
	return "UnloadChunk"
}

func (p UnloadChunk) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Z, p.X)
}

func (p *UnloadChunk) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Z, &p.X)
}

type BlockUpdate struct {
	Location gocraft.Position
	State    gocraft.VarInt
}

func (*BlockUpdate) ID() int32 {
	return 0x09
}

func (*BlockUpdate) Name() string {
	return "BlockUpdate"
}

func (p BlockUpdate) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Location, p.State)
}

func (p *BlockUpdate) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Location, &p.State)
}

type BlockChange struct {
	X     int
	Y     int
	Z     int
	State gocraft.BlockState
}

func (p *BlockUpdate) Change() BlockChange {
	return BlockChange{
		X:     p.Location.X,
		Y:     p.Location.Y,
		Z:     p.Location.Z,
		State: gocraft.BlockState(p.State),
	}
}

type SectionBlocksUpdate struct {
	Section gocraft.Long
	Packed  gocraft.Slice[gocraft.VarLong]
}

func (*SectionBlocksUpdate) ID() int32 {
	return 0x47
}

func (*SectionBlocksUpdate) Name() string {
	return "SectionBlocksUpdate"
}

func (p SectionBlocksUpdate) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Section, p.Packed)
}

func (p *SectionBlocksUpdate) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Section, &p.Packed)
}

func (p *SectionBlocksUpdate) Changes() []BlockChange {
	baseX := int(p.Section.Signed(42, 22)) * 16
	baseZ := int(p.Section.Signed(20, 22)) * 16
	baseY := int(p.Section.Signed(0, 20)) * 16

	changes := make([]BlockChange, len(p.Packed))
	for i, packed := range p.Packed {
		block := gocraft.Long(packed)
		changes[i] = BlockChange{
			X:     baseX + int(block.Unsigned(8, 4)),
			Y:     baseY + int(block.Unsigned(0, 4)),
			Z:     baseZ + int(block.Unsigned(4, 4)),
			State: gocraft.BlockState(block.Unsigned(12, 52)),
		}
	}

	return changes
}

type SetHealth struct {
	Health     gocraft.Float
	Food       gocraft.VarInt
	Saturation gocraft.Float
}

func (*SetHealth) ID() int32 {
	return 0x5B
}

func (*SetHealth) Name() string {
	return "SetHealth"
}

func (p SetHealth) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Health, p.Food, p.Saturation)
}

func (p *SetHealth) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Health, &p.Food, &p.Saturation)
}

type SetPlayerPosition struct {
	X        gocraft.Double
	Y        gocraft.Double
	Z        gocraft.Double
	OnGround gocraft.Bool
}

func (*SetPlayerPosition) ID() int32 {
	return 0x17
}

func (*SetPlayerPosition) Name() string {
	return "SetPlayerPosition"
}

func (p SetPlayerPosition) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.X, p.Y, p.Z, p.OnGround)
}

func (p *SetPlayerPosition) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.X, &p.Y, &p.Z, &p.OnGround)
}

type SetPlayerPositionRotation struct {
	X        gocraft.Double
	Y        gocraft.Double
	Z        gocraft.Double
	Yaw      gocraft.Float
	Pitch    gocraft.Float
	OnGround gocraft.Bool
}

func (*SetPlayerPositionRotation) ID() int32 {
	return 0x18
}

func (*SetPlayerPositionRotation) Name() string {
	return "SetPlayerPositionRotation"
}

func (p SetPlayerPositionRotation) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.X, p.Y, p.Z, p.Yaw, p.Pitch, p.OnGround)
}

func (p *SetPlayerPositionRotation) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.X, &p.Y, &p.Z, &p.Yaw, &p.Pitch, &p.OnGround)
}
