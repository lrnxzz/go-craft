package nbt

import (
	"fmt"
	"math"
	"reflect"
)

var _tagInterface = reflect.TypeFor[Tag]()

func Unmarshal(data []byte, v any) error {
	_, err := _unmarshalRoot(data, false, v)

	return err
}

func UnmarshalNamed(data []byte, v any) (string, error) {
	return _unmarshalRoot(data, true, v)
}

func _unmarshalRoot(data []byte, named bool, v any) (string, error) {
	target := reflect.ValueOf(v)
	if target.Kind() != reflect.Pointer || target.IsNil() {
		return "", fmt.Errorf("nbt: unmarshal target must be a non-nil pointer, got %T", v)
	}

	dec := &decoder{buf: data}
	if TagType(dec._byte()) != TagCompound {
		if dec.err != nil {
			return "", dec.err
		}
		return "", fmt.Errorf("nbt: root tag is not a compound")
	}

	name := ""
	if named {
		name = dec._string()
	}

	u := unmarshaler{dec: dec}
	u._compound(target.Elem())

	return name, dec.err
}

type unmarshaler struct {
	dec *decoder
}

func (u *unmarshaler) _compound(target reflect.Value) {
	if !u.dec._enter() {
		return
	}
	defer u.dec._leave()

	target = _settable(target)

	switch target.Kind() {
	case reflect.Struct:
		u._structEntries(target)
	case reflect.Map:
		u._mapEntries(target)
	case reflect.Interface:
		if tree := u.dec._compound(); u.dec.err == nil {
			target.Set(reflect.ValueOf(tree))
		}
	default:
		u.dec._fail(fmt.Errorf("nbt: cannot unmarshal compound into %s", target.Kind()))
	}
}

func (u *unmarshaler) _structEntries(target reflect.Value) {
	fields := _fieldMapOf(target.Type())

	for {
		tag := TagType(u.dec._byte())
		if u.dec.err != nil || tag == TagEnd {
			return
		}

		name := u.dec._string()
		if f, ok := fields[name]; ok {
			u._value(tag, _fieldTarget(target, f.index))
		} else {
			u.dec._payload(tag)
		}
	}
}

func (u *unmarshaler) _mapEntries(target reflect.Value) {
	if target.Type().Key().Kind() != reflect.String {
		u.dec._fail(fmt.Errorf("nbt: map key must be string, got %s", target.Type().Key()))
		return
	}
	if target.IsNil() {
		target.Set(reflect.MakeMap(target.Type()))
	}

	elemType := target.Type().Elem()

	for {
		tag := TagType(u.dec._byte())
		if u.dec.err != nil || tag == TagEnd {
			return
		}

		name := u.dec._string()
		elem := reflect.New(elemType).Elem()
		u._value(tag, elem)
		if u.dec.err != nil {
			return
		}

		target.SetMapIndex(reflect.ValueOf(name), elem)
	}
}

func (u *unmarshaler) _value(tag TagType, target reflect.Value) {
	if u.dec.err != nil {
		return
	}

	if target.Kind() == reflect.Pointer {
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}
		u._value(tag, target.Elem())
		return
	}

	if target.Kind() == reflect.Interface || target.Type().Implements(_tagInterface) {
		u._dynamic(tag, target)
		return
	}

	switch tag {
	case TagByte:
		u._setInt(target, int64(int8(u.dec._byte())))
	case TagShort:
		u._setInt(target, int64(int16(u.dec._u16())))
	case TagInt:
		u._setInt(target, int64(int32(u.dec._u32())))
	case TagLong:
		u._setInt(target, int64(u.dec._u64()))
	case TagFloat:
		u._setFloat(target, float64(math.Float32frombits(u.dec._u32())))
	case TagDouble:
		u._setFloat(target, math.Float64frombits(u.dec._u64()))
	case TagString:
		u._setString(target, u.dec._string())
	case TagByteArray, TagIntArray, TagLongArray:
		u._array(tag, target)
	case TagList:
		u._list(target)
	case TagCompound:
		u._compound(target)
	default:
		u.dec._fail(fmt.Errorf("nbt: unknown tag %d", tag))
	}
}

func (u *unmarshaler) _dynamic(tag TagType, target reflect.Value) {
	tree := u.dec._payload(tag)
	if u.dec.err != nil {
		return
	}

	value := reflect.ValueOf(tree)
	switch {
	case value.Type().AssignableTo(target.Type()):
		target.Set(value)
	case value.Type().ConvertibleTo(target.Type()):
		target.Set(value.Convert(target.Type()))
	default:
		u.dec._fail(fmt.Errorf("nbt: cannot assign %s to %s", value.Type(), target.Type()))
	}
}

func (u *unmarshaler) _list(target reflect.Value) {
	elem := TagType(u.dec._byte())
	n := u.dec._length()
	if u.dec.err != nil {
		return
	}

	if target.Kind() != reflect.Slice {
		for range n {
			u.dec._payload(elem)
		}
		u.dec._fail(fmt.Errorf("nbt: cannot unmarshal list into %s", target.Type()))
		return
	}

	slice := reflect.MakeSlice(target.Type(), 0, min(n, maxPrealloc))
	for range n {
		item := reflect.New(target.Type().Elem()).Elem()
		u._value(elem, item)
		if u.dec.err != nil {
			return
		}
		slice = reflect.Append(slice, item)
	}

	target.Set(slice)
}

func (u *unmarshaler) _array(tag TagType, target reflect.Value) {
	n := u.dec._length()
	if u.dec.err != nil {
		return
	}

	if target.Kind() != reflect.Slice {
		u._skipArray(tag, n)
		u.dec._fail(fmt.Errorf("nbt: cannot unmarshal array into %s", target.Type()))
		return
	}

	slice := reflect.MakeSlice(target.Type(), n, n)
	for i := range n {
		switch tag {
		case TagByteArray:
			slice.Index(i).SetInt(int64(int8(u.dec._byte())))
		case TagIntArray:
			slice.Index(i).SetInt(int64(int32(u.dec._u32())))
		case TagLongArray:
			slice.Index(i).SetInt(int64(u.dec._u64()))
		}
		if u.dec.err != nil {
			return
		}
	}

	target.Set(slice)
}

func (u *unmarshaler) _skipArray(tag TagType, n int) {
	width := map[TagType]int{TagByteArray: 1, TagIntArray: 4, TagLongArray: 8}[tag]
	u.dec._take(n * width)
}

func (u *unmarshaler) _setInt(target reflect.Value, value int64) {
	switch target.Kind() {
	case reflect.Bool:
		target.SetBool(value != 0)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		target.SetInt(value)
	default:
		u.dec._fail(fmt.Errorf("nbt: cannot assign integer to %s", target.Type()))
	}
}

func (u *unmarshaler) _setFloat(target reflect.Value, value float64) {
	if target.Kind() == reflect.Float32 || target.Kind() == reflect.Float64 {
		target.SetFloat(value)
		return
	}

	u.dec._fail(fmt.Errorf("nbt: cannot assign float to %s", target.Type()))
}

func (u *unmarshaler) _setString(target reflect.Value, value string) {
	if target.Kind() == reflect.String {
		target.SetString(value)
		return
	}

	u.dec._fail(fmt.Errorf("nbt: cannot assign string to %s", target.Type()))
}

func _settable(target reflect.Value) reflect.Value {
	for target.Kind() == reflect.Pointer {
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}
		target = target.Elem()
	}

	return target
}
