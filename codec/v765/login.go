package v765

import gocraft "github.com/lrnxzz/go-craft"

type LoginStart struct {
	Username gocraft.String
	UUID     gocraft.UUID
}

func (*LoginStart) ID() int32 {
	return 0x00
}

func (*LoginStart) Name() string {
	return "LoginStart"
}

func (p LoginStart) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Username, p.UUID)
}

func (p *LoginStart) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Username, &p.UUID)
}

type EncryptionBegin struct {
	ServerID gocraft.String
}

func (*EncryptionBegin) ID() int32 {
	return 0x01
}

func (*EncryptionBegin) Name() string {
	return "EncryptionBegin"
}

func (p EncryptionBegin) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.ServerID)
}

func (p *EncryptionBegin) Decode(r *gocraft.Reader) error {
	if err := p.ServerID.Decode(r); err != nil {
		return err
	}

	r.Rest()

	return nil
}

type LoginAcknowledged struct{}

func (*LoginAcknowledged) ID() int32 {
	return 0x03
}

func (*LoginAcknowledged) Name() string {
	return "LoginAcknowledged"
}

func (LoginAcknowledged) Append(dst []byte) []byte {
	return dst
}

func (*LoginAcknowledged) Decode(*gocraft.Reader) error {
	return nil
}

type Property struct {
	Name      gocraft.String
	Value     gocraft.String
	Signature gocraft.Option[gocraft.String]
}

func (p Property) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Name, p.Value, p.Signature)
}

func (p *Property) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Name, &p.Value, &p.Signature)
}

type LoginSuccess struct {
	UUID       gocraft.UUID
	Username   gocraft.String
	Properties gocraft.Slice[Property]
}

func (*LoginSuccess) ID() int32 {
	return 0x02
}

func (*LoginSuccess) Name() string {
	return "LoginSuccess"
}

func (p LoginSuccess) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.UUID, p.Username, p.Properties)
}

func (p *LoginSuccess) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.UUID, &p.Username, &p.Properties)
}

type SetCompression struct {
	Threshold gocraft.VarInt
}

func (*SetCompression) ID() int32 {
	return 0x03
}

func (*SetCompression) Name() string {
	return "SetCompression"
}

func (p SetCompression) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Threshold)
}

func (p *SetCompression) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Threshold)
}

type LoginDisconnect struct {
	Reason gocraft.String
}

func (*LoginDisconnect) ID() int32 {
	return 0x00
}

func (*LoginDisconnect) Name() string {
	return "LoginDisconnect"
}

func (p LoginDisconnect) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, p.Reason)
}

func (p *LoginDisconnect) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &p.Reason)
}
