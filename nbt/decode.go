package nbt

import (
	"encoding/binary"
	"fmt"
	"math"
)

const (
	maxPrealloc = 1024
	maxDepth    = 512
)

type decoder struct {
	buf   []byte
	off   int
	depth int
	err   error
}

func Decode(data []byte) (Compound, error) {
	compound, _, err := DecodePrefix(data)

	return compound, err
}

func DecodePrefix(data []byte) (Compound, int, error) {
	dec := &decoder{buf: data}

	root := dec.u8()
	if dec.err != nil {
		return nil, 0, dec.err
	}
	if TagType(root) != TagCompound {
		return nil, 0, fmt.Errorf("nbt: root tag is %d, want compound", root)
	}

	compound := dec.compound()
	if dec.err != nil {
		return nil, 0, dec.err
	}

	return compound, dec.off, nil
}

func DecodeNamed(data []byte) (string, Compound, error) {
	dec := &decoder{buf: data}

	root := dec.u8()
	if dec.err != nil {
		return "", nil, dec.err
	}
	if TagType(root) != TagCompound {
		return "", nil, fmt.Errorf("nbt: root tag is %d, want compound", root)
	}

	name := dec.str()
	compound := dec.compound()
	if dec.err != nil {
		return "", nil, dec.err
	}

	return name, compound, nil
}

func (d *decoder) enter() bool {
	d.depth++
	if d.depth > maxDepth {
		d.fail(fmt.Errorf("nbt: nesting exceeds %d", maxDepth))
		return false
	}

	return true
}

func (d *decoder) leave() {
	d.depth--
}

func (d *decoder) fail(err error) {
	if d.err == nil {
		d.err = err
	}
}

func (d *decoder) take(n int) []byte {
	if d.err != nil {
		return nil
	}
	if n < 0 || n > len(d.buf)-d.off {
		d.fail(fmt.Errorf("nbt: payload needs %d bytes, has %d", n, len(d.buf)-d.off))
		return nil
	}

	view := d.buf[d.off : d.off+n]
	d.off += n

	return view
}

func (d *decoder) u8() byte {
	raw := d.take(1)
	if raw == nil {
		return 0
	}

	return raw[0]
}

func (d *decoder) u16() uint16 {
	raw := d.take(2)
	if raw == nil {
		return 0
	}

	return binary.BigEndian.Uint16(raw)
}

func (d *decoder) u32() uint32 {
	raw := d.take(4)
	if raw == nil {
		return 0
	}

	return binary.BigEndian.Uint32(raw)
}

func (d *decoder) u64() uint64 {
	raw := d.take(8)
	if raw == nil {
		return 0
	}

	return binary.BigEndian.Uint64(raw)
}

func (d *decoder) length() int {
	n := int32(d.u32())
	if n < 0 {
		d.fail(fmt.Errorf("nbt: negative length %d", n))
		return 0
	}

	return int(n)
}

func (d *decoder) str() string {
	raw := d.take(int(d.u16()))
	if raw == nil {
		return ""
	}

	decoded, err := decodeMUTF8(raw)
	if err != nil {
		d.fail(err)
		return ""
	}

	return decoded
}

func (d *decoder) compound() Compound {
	if !d.enter() {
		return nil
	}
	defer d.leave()

	compound := Compound{}

	for {
		tag := TagType(d.u8())
		if d.err != nil || tag == TagEnd {
			return compound
		}

		name := d.str()
		compound[name] = d.payload(tag)
	}
}

func (d *decoder) list() List {
	if !d.enter() {
		return List{}
	}
	defer d.leave()

	elem := TagType(d.u8())
	n := d.length()

	items := make([]Tag, 0, min(n, maxPrealloc))
	for range n {
		items = append(items, d.payload(elem))
		if d.err != nil {
			break
		}
	}

	return List{Elem: elem, Items: items}
}

func (d *decoder) payload(tag TagType) Tag {
	switch tag {
	case TagByte:
		return Byte(d.u8())
	case TagShort:
		return Short(d.u16())
	case TagInt:
		return Int(d.u32())
	case TagLong:
		return Long(d.u64())
	case TagFloat:
		return Float(math.Float32frombits(d.u32()))
	case TagDouble:
		return Double(math.Float64frombits(d.u64()))
	case TagByteArray:
		return ByteArray(append([]byte(nil), d.take(d.length())...))
	case TagString:
		return String(d.str())
	case TagList:
		return d.list()
	case TagCompound:
		return d.compound()
	case TagIntArray:
		return d.intArray()
	case TagLongArray:
		return d.longArray()
	}

	d.fail(fmt.Errorf("nbt: unknown tag %d", tag))

	return nil
}

func (d *decoder) intArray() IntArray {
	n := d.length()

	array := make(IntArray, 0, min(n, maxPrealloc))
	for range n {
		array = append(array, int32(d.u32()))
		if d.err != nil {
			break
		}
	}

	return array
}

func (d *decoder) longArray() LongArray {
	n := d.length()

	array := make(LongArray, 0, min(n, maxPrealloc))
	for range n {
		array = append(array, int64(d.u64()))
		if d.err != nil {
			break
		}
	}

	return array
}
