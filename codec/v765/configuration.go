package v765

import gocraft "github.com/lrnxzz/go-craft"

type ClientInformation struct {
	Locale              gocraft.String
	ViewDistance        gocraft.Byte
	ChatMode            gocraft.VarInt
	ChatColors          gocraft.Bool
	DisplayedSkinParts  gocraft.UByte
	MainHand            gocraft.VarInt
	EnableTextFiltering gocraft.Bool
	EnableServerListing gocraft.Bool
}

func (ClientInformation) ID() int32 {
	return 0x00
}

func (p ClientInformation) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Locale, p.ViewDistance, p.ChatMode, p.ChatColors,
		p.DisplayedSkinParts, p.MainHand, p.EnableTextFiltering, p.EnableServerListing)
}

func (p *ClientInformation) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Locale, &p.ViewDistance, &p.ChatMode, &p.ChatColors,
		&p.DisplayedSkinParts, &p.MainHand, &p.EnableTextFiltering, &p.EnableServerListing)
}

type ConfigDisconnect struct {
	Reason gocraft.NBT
}

func (ConfigDisconnect) ID() int32 {
	return 0x01
}

func (p ConfigDisconnect) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Reason)
}

func (p *ConfigDisconnect) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Reason)
}

type FinishConfiguration struct{}

func (FinishConfiguration) ID() int32 {
	return 0x02
}

func (FinishConfiguration) Append(dst []byte) []byte {
	return dst
}

func (*FinishConfiguration) Decode(*gocraft.Reader) error {
	return nil
}

type ConfigKeepAlive struct {
	KeepAliveID gocraft.Long
}

func (ConfigKeepAlive) ID() int32 {
	return 0x03
}

func (p ConfigKeepAlive) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.KeepAliveID)
}

func (p *ConfigKeepAlive) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.KeepAliveID)
}

type ConfigPing struct {
	PingID gocraft.Int
}

func (ConfigPing) ID() int32 {
	return 0x04
}

func (p ConfigPing) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.PingID)
}

func (p *ConfigPing) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.PingID)
}

type ConfigPong struct {
	PingID gocraft.Int
}

func (ConfigPong) ID() int32 {
	return 0x04
}

func (p ConfigPong) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.PingID)
}

func (p *ConfigPong) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.PingID)
}

type RegistryData struct {
	Codec gocraft.NBT
}

func (RegistryData) ID() int32 {
	return 0x05
}

func (p RegistryData) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Codec)
}

func (p *RegistryData) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Codec)
}

type FeatureFlags struct {
	Features gocraft.Slice[gocraft.Identifier]
}

func (FeatureFlags) ID() int32 {
	return 0x08
}

func (p FeatureFlags) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Features)
}

func (p *FeatureFlags) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Features)
}
