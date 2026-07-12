package gocraft

import (
	"fmt"
	"math"
	"unsafe"

	"golang.org/x/exp/constraints"
)

const (
	MaxStringLen   = 32767
	maxStringBytes = 3*MaxStringLen + 3
	maxPrealloc    = 4096
)

type Field interface {
	Append(dst []byte) []byte
}

type FieldPtr interface {
	Decode(r *Reader) error
}

func Marshal(fields ...Field) []byte {
	var payload []byte
	for _, field := range fields {
		payload = field.Append(payload)
	}
	return payload
}

func Unmarshal(payload []byte, fields ...FieldPtr) error {
	r := NewReader(payload)
	for _, field := range fields {
		if err := field.Decode(r); err != nil {
			return err
		}
	}
	return nil
}

func _appendBE[T constraints.Unsigned](dst []byte, v T) []byte {
	for shift := (int(unsafe.Sizeof(v)) - 1) * bitsPerByte; shift >= 0; shift -= bitsPerByte {
		dst = append(dst, byte(v>>shift))
	}
	return dst
}

func _readBE[T constraints.Unsigned](r *Reader) (T, error) {
	raw := r.take(int(unsafe.Sizeof(T(0))))
	if raw == nil {
		return 0, r.err
	}
	var v T
	for i, octet := range raw {
		v |= T(octet) << ((len(raw) - 1 - i) * bitsPerByte)
	}
	return v, nil
}

type (
	Bool    bool
	Byte    int8
	UByte   uint8
	Short   int16
	UShort  uint16
	Int     int32
	Long    int64
	Float   float32
	Double  float64
	VarInt  int32
	VarLong int64
	String  string
	UUID    [16]byte
)

func (v Bool) Append(dst []byte) []byte {
	var raw uint8
	if v {
		raw = 1
	}
	return _appendBE(dst, raw)
}

func (v *Bool) Decode(r *Reader) error {
	raw, err := _readBE[uint8](r)
	if err != nil {
		return err
	}
	*v = raw != 0
	return nil
}

func (v Byte) Append(dst []byte) []byte {
	return _appendBE(dst, uint8(v))
}

func (v *Byte) Decode(r *Reader) error {
	raw, err := _readBE[uint8](r)
	if err != nil {
		return err
	}
	*v = Byte(raw)
	return nil
}

func (v UByte) Append(dst []byte) []byte {
	return _appendBE(dst, uint8(v))
}

func (v *UByte) Decode(r *Reader) error {
	raw, err := _readBE[uint8](r)
	if err != nil {
		return err
	}
	*v = UByte(raw)
	return nil
}

func (v Short) Append(dst []byte) []byte {
	return _appendBE(dst, uint16(v))
}

func (v *Short) Decode(r *Reader) error {
	raw, err := _readBE[uint16](r)
	if err != nil {
		return err
	}
	*v = Short(raw)
	return nil
}

func (v UShort) Append(dst []byte) []byte {
	return _appendBE(dst, uint16(v))
}

func (v *UShort) Decode(r *Reader) error {
	raw, err := _readBE[uint16](r)
	if err != nil {
		return err
	}
	*v = UShort(raw)
	return nil
}

func (v Int) Append(dst []byte) []byte {
	return _appendBE(dst, uint32(v))
}

func (v *Int) Decode(r *Reader) error {
	raw, err := _readBE[uint32](r)
	if err != nil {
		return err
	}
	*v = Int(raw)
	return nil
}

func (v Long) Append(dst []byte) []byte {
	return _appendBE(dst, uint64(v))
}

func (v *Long) Decode(r *Reader) error {
	raw, err := _readBE[uint64](r)
	if err != nil {
		return err
	}
	*v = Long(raw)
	return nil
}

func (v Float) Append(dst []byte) []byte {
	return _appendBE(dst, math.Float32bits(float32(v)))
}

func (v *Float) Decode(r *Reader) error {
	raw, err := _readBE[uint32](r)
	if err != nil {
		return err
	}
	*v = Float(math.Float32frombits(raw))
	return nil
}

func (v Double) Append(dst []byte) []byte {
	return _appendBE(dst, math.Float64bits(float64(v)))
}

func (v *Double) Decode(r *Reader) error {
	raw, err := _readBE[uint64](r)
	if err != nil {
		return err
	}
	*v = Double(math.Float64frombits(raw))
	return nil
}

func (v VarInt) Append(dst []byte) []byte {
	return AppendVar(dst, v)
}

func (v *VarInt) Decode(r *Reader) error {
	val, err := ReadVar[VarInt](r)
	if err != nil {
		return r.fail(err)
	}
	*v = val
	return nil
}

func (v VarLong) Append(dst []byte) []byte {
	return AppendVar(dst, v)
}

func (v *VarLong) Decode(r *Reader) error {
	val, err := ReadVar[VarLong](r)
	if err != nil {
		return r.fail(err)
	}
	*v = val
	return nil
}

func (v String) Append(dst []byte) []byte {
	dst = AppendVar(dst, VarInt(len(v)))
	return append(dst, v...)
}

func (v *String) Decode(r *Reader) error {
	n, err := ReadVar[VarInt](r)
	if err != nil {
		return r.fail(err)
	}
	if n < 0 || n > maxStringBytes {
		return r.fail(fmt.Errorf("gocraft: string of %d bytes is out of range", n))
	}
	raw := r.take(int(n))
	if raw == nil {
		return r.err
	}
	*v = String(raw)
	return nil
}

func (v UUID) Append(dst []byte) []byte {
	return append(dst, v[:]...)
}

func (v *UUID) Decode(r *Reader) error {
	raw := r.take(len(v))
	if raw == nil {
		return r.err
	}
	copy(v[:], raw)
	return nil
}

type Slice[T Field] []T

func (s Slice[T]) Append(dst []byte) []byte {
	dst = AppendVar(dst, VarInt(len(s)))
	for _, element := range s {
		dst = element.Append(dst)
	}
	return dst
}

func (s *Slice[T]) Decode(r *Reader) error {
	var n VarInt
	if err := n.Decode(r); err != nil {
		return err
	}
	if n < 0 {
		return r.fail(fmt.Errorf("gocraft: slice of %d elements is out of range", n))
	}
	elements := make(Slice[T], 0, min(int(n), maxPrealloc))
	for range int(n) {
		var element T
		ptr, ok := any(&element).(FieldPtr)
		if !ok {
			return r.fail(fmt.Errorf("gocraft: %T does not implement FieldPtr", &element))
		}
		if err := ptr.Decode(r); err != nil {
			return err
		}
		elements = append(elements, element)
	}
	*s = elements
	return nil
}

type Option[T Field] struct {
	value   T
	present bool
}

func Some[T Field](v T) Option[T] {
	return Option[T]{value: v, present: true}
}

func None[T Field]() Option[T] {
	return Option[T]{}
}

func (o Option[T]) Get() (T, bool) {
	return o.value, o.present
}

func (o Option[T]) Append(dst []byte) []byte {
	dst = Bool(o.present).Append(dst)
	if o.present {
		dst = o.value.Append(dst)
	}
	return dst
}

func (o *Option[T]) Decode(r *Reader) error {
	var present Bool
	if err := present.Decode(r); err != nil {
		return err
	}
	o.present = bool(present)
	if !o.present {
		var zero T
		o.value = zero
		return nil
	}
	ptr, ok := any(&o.value).(FieldPtr)
	if !ok {
		return r.fail(fmt.Errorf("gocraft: %T does not implement FieldPtr", &o.value))
	}
	return ptr.Decode(r)
}

var (
	_ Field = Bool(false)
	_ Field = Byte(0)
	_ Field = UByte(0)
	_ Field = Short(0)
	_ Field = UShort(0)
	_ Field = Int(0)
	_ Field = Long(0)
	_ Field = Float(0)
	_ Field = Double(0)
	_ Field = VarInt(0)
	_ Field = VarLong(0)
	_ Field = String("")
	_ Field = UUID{}
	_ Field = Slice[VarInt](nil)
	_ Field = Option[VarInt]{}

	_ FieldPtr = (*Bool)(nil)
	_ FieldPtr = (*Byte)(nil)
	_ FieldPtr = (*UByte)(nil)
	_ FieldPtr = (*Short)(nil)
	_ FieldPtr = (*UShort)(nil)
	_ FieldPtr = (*Int)(nil)
	_ FieldPtr = (*Long)(nil)
	_ FieldPtr = (*Float)(nil)
	_ FieldPtr = (*Double)(nil)
	_ FieldPtr = (*VarInt)(nil)
	_ FieldPtr = (*VarLong)(nil)
	_ FieldPtr = (*String)(nil)
	_ FieldPtr = (*UUID)(nil)
	_ FieldPtr = (*Slice[VarInt])(nil)
	_ FieldPtr = (*Option[VarInt])(nil)
)
