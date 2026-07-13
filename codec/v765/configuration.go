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

func (*ClientInformation) ID() int32 {
	return 0x00
}

func (*ClientInformation) Name() string {
	return "ClientInformation"
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

func (*ConfigDisconnect) ID() int32 {
	return 0x01
}

func (*ConfigDisconnect) Name() string {
	return "ConfigDisconnect"
}

func (p ConfigDisconnect) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Reason)
}

func (p *ConfigDisconnect) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Reason)
}

type FinishConfiguration struct{}

func (*FinishConfiguration) ID() int32 {
	return 0x02
}

func (*FinishConfiguration) Name() string {
	return "FinishConfiguration"
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

func (*ConfigKeepAlive) ID() int32 {
	return 0x03
}

func (*ConfigKeepAlive) Name() string {
	return "ConfigKeepAlive"
}

func (p ConfigKeepAlive) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.KeepAliveID)
}

func (p *ConfigKeepAlive) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.KeepAliveID)
}

type ConfigKeepAliveResponse struct {
	KeepAliveID gocraft.Long
}

func (*ConfigKeepAliveResponse) ID() int32 {
	return 0x03
}

func (*ConfigKeepAliveResponse) Name() string {
	return "ConfigKeepAliveResponse"
}

func (p ConfigKeepAliveResponse) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.KeepAliveID)
}

func (p *ConfigKeepAliveResponse) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.KeepAliveID)
}

type ConfigPing struct {
	PingID gocraft.Int
}

func (*ConfigPing) ID() int32 {
	return 0x04
}

func (*ConfigPing) Name() string {
	return "ConfigPing"
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

func (*ConfigPong) ID() int32 {
	return 0x04
}

func (*ConfigPong) Name() string {
	return "ConfigPong"
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

func (*RegistryData) ID() int32 {
	return 0x05
}

func (*RegistryData) Name() string {
	return "RegistryData"
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

func (*FeatureFlags) ID() int32 {
	return 0x07
}

func (*FeatureFlags) Name() string {
	return "FeatureFlags"
}

func (p FeatureFlags) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Features)
}

func (p *FeatureFlags) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Features)
}
