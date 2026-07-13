package nbt

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"slices"
	"strings"
)

func Marshal(v any) ([]byte, error) {
	m := marshaler{buf: []byte{byte(TagCompound)}}
	m.compound(reflect.ValueOf(v))

	return m.buf, m.err
}

func MarshalNamed(name string, v any) ([]byte, error) {
	m := marshaler{buf: []byte{byte(TagCompound)}}
	m.writeString(name)
	m.compound(reflect.ValueOf(v))

	return m.buf, m.err
}

type marshaler struct {
	buf []byte
	err error
}

func (m *marshaler) fail(err error) {
	if m.err == nil {
		m.err = err
	}
}

func (m *marshaler) writeString(s string) {
	if n := mutf8Len(s); n > math.MaxUint16 {
		m.fail(fmt.Errorf("nbt: string of %d bytes exceeds the %d-byte limit", n, math.MaxUint16))
		return
	}

	m.buf = encodeString(m.buf, s)
}

func (m *marshaler) compound(v reflect.Value) {
	if m.err != nil {
		return
	}

	v = deref(v)

	switch v.Kind() {
	case reflect.Struct:
		for _, f := range fieldsOf(v.Type()) {
			target := fieldValue(v, f.index)
			if !target.IsValid() {
				continue
			}
			if f.omitempty && target.IsZero() {
				continue
			}

			m.namedTag(f.name, deref(target), f.asList)
		}
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			m.fail(fmt.Errorf("nbt: map key must be string, got %s", v.Type().Key()))
			return
		}

		keys := v.MapKeys()
		slices.SortFunc(keys, func(a, b reflect.Value) int {
			return strings.Compare(a.String(), b.String())
		})
		for _, key := range keys {
			m.namedTag(key.String(), deref(v.MapIndex(key)), false)
		}
	default:
		m.fail(fmt.Errorf("nbt: cannot marshal %s as a compound", v.Kind()))
		return
	}

	m.buf = append(m.buf, byte(TagEnd))
}

func (m *marshaler) namedTag(name string, v reflect.Value, asList bool) {
	if m.err != nil {
		return
	}

	tag := m.tagOf(v, asList)
	if m.err != nil {
		return
	}

	m.buf = append(m.buf, byte(tag))
	m.writeString(name)
	m.payload(v, asList)
}

func (m *marshaler) tagOf(v reflect.Value, asList bool) TagType {
	if tag, ok := asTag(v); ok {
		return tag.Type()
	}

	switch v.Kind() {
	case reflect.Bool, reflect.Int8, reflect.Uint8:
		return TagByte
	case reflect.Int16, reflect.Uint16:
		return TagShort
	case reflect.Int32, reflect.Uint32:
		return TagInt
	case reflect.Int64, reflect.Int, reflect.Uint64, reflect.Uint:
		return TagLong
	case reflect.Float32:
		return TagFloat
	case reflect.Float64:
		return TagDouble
	case reflect.String:
		return TagString
	case reflect.Slice, reflect.Array:
		return sequenceTag(v.Type().Elem(), asList)
	case reflect.Struct, reflect.Map:
		return TagCompound
	}

	m.fail(fmt.Errorf("nbt: unsupported type %s", v.Type()))

	return TagEnd
}

func (m *marshaler) payload(v reflect.Value, asList bool) {
	if m.err != nil {
		return
	}

	if tag, ok := asTag(v); ok {
		m.buf = encodePayload(m.buf, tag)
		return
	}

	switch v.Kind() {
	case reflect.Bool:
		flag := Byte(0)
		if v.Bool() {
			flag = 1
		}
		m.buf = encodePayload(m.buf, flag)
	case reflect.Int8:
		m.buf = encodePayload(m.buf, Byte(v.Int()))
	case reflect.Int16:
		m.buf = encodePayload(m.buf, Short(v.Int()))
	case reflect.Int32:
		m.buf = encodePayload(m.buf, Int(v.Int()))
	case reflect.Int64, reflect.Int:
		m.buf = encodePayload(m.buf, Long(v.Int()))
	case reflect.Uint8:
		m.buf = encodePayload(m.buf, Byte(v.Uint()))
	case reflect.Uint16:
		m.buf = encodePayload(m.buf, Short(v.Uint()))
	case reflect.Uint32:
		m.buf = encodePayload(m.buf, Int(v.Uint()))
	case reflect.Uint64, reflect.Uint:
		m.buf = encodePayload(m.buf, Long(v.Uint()))
	case reflect.Float32:
		m.buf = encodePayload(m.buf, Float(v.Float()))
	case reflect.Float64:
		m.buf = encodePayload(m.buf, Double(v.Float()))
	case reflect.String:
		m.writeString(v.String())
	case reflect.Slice, reflect.Array:
		m.sequence(v, asList)
	case reflect.Struct, reflect.Map:
		m.compound(v)
	default:
		m.fail(fmt.Errorf("nbt: unsupported type %s", v.Type()))
	}
}

func (m *marshaler) sequence(v reflect.Value, asList bool) {
	length := v.Len()

	switch sequenceTag(v.Type().Elem(), asList) {
	case TagByteArray:
		m.buf = binary.BigEndian.AppendUint32(m.buf, uint32(length))
		signed := v.Type().Elem().Kind() == reflect.Int8
		for i := range length {
			if signed {
				m.buf = append(m.buf, byte(v.Index(i).Int()))
			} else {
				m.buf = append(m.buf, byte(v.Index(i).Uint()))
			}
		}
	case TagIntArray:
		m.buf = binary.BigEndian.AppendUint32(m.buf, uint32(length))
		for i := range length {
			m.buf = binary.BigEndian.AppendUint32(m.buf, uint32(v.Index(i).Int()))
		}
	case TagLongArray:
		m.buf = binary.BigEndian.AppendUint32(m.buf, uint32(length))
		for i := range length {
			m.buf = binary.BigEndian.AppendUint64(m.buf, uint64(v.Index(i).Int()))
		}
	default:
		m.list(v)
	}
}

func (m *marshaler) list(v reflect.Value) {
	length := v.Len()

	elem := TagEnd
	if length > 0 {
		elem = m.tagOf(deref(v.Index(0)), false)
	}
	if m.err != nil {
		return
	}

	m.buf = append(m.buf, byte(elem))
	m.buf = binary.BigEndian.AppendUint32(m.buf, uint32(length))

	for i := range length {
		m.payload(deref(v.Index(i)), false)
	}
}

func sequenceTag(elem reflect.Type, asList bool) TagType {
	if !asList {
		switch elem.Kind() {
		case reflect.Uint8, reflect.Int8:
			return TagByteArray
		case reflect.Int32:
			return TagIntArray
		case reflect.Int64:
			return TagLongArray
		}
	}

	return TagList
}

func asTag(v reflect.Value) (Tag, bool) {
	if !v.IsValid() || !v.CanInterface() {
		return nil, false
	}

	tag, ok := v.Interface().(Tag)

	return tag, ok
}
