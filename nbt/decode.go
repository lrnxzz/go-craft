package nbt

import (
	"encoding/binary"
	"fmt"
	"math"
)

const maxPrealloc = 1024

type decoder struct {
	buf []byte
	off int
	err error
}

func Decode(data []byte) (Compound, error) {
	dec := &decoder{buf: data}

	root := dec._byte()
	if dec.err != nil {
		return nil, dec.err
	}
	if TagType(root) != TagCompound {
		return nil, fmt.Errorf("nbt: root tag is %d, want compound", root)
	}

	compound := dec._compound()
	if dec.err != nil {
		return nil, dec.err
	}

	return compound, nil
}

func (d *decoder) _fail(err error) {
	if d.err == nil {
		d.err = err
	}
}

func (d *decoder) _take(n int) []byte {
	if d.err != nil {
		return nil
	}
	if n < 0 || n > len(d.buf)-d.off {
		d._fail(fmt.Errorf("nbt: payload needs %d bytes, has %d", n, len(d.buf)-d.off))
		return nil
	}

	view := d.buf[d.off : d.off+n]
	d.off += n

	return view
}

func (d *decoder) _byte() byte {
	raw := d._take(1)
	if raw == nil {
		return 0
	}

	return raw[0]
}

func (d *decoder) _u16() uint16 {
	raw := d._take(2)
	if raw == nil {
		return 0
	}

	return binary.BigEndian.Uint16(raw)
}

func (d *decoder) _u32() uint32 {
	raw := d._take(4)
	if raw == nil {
		return 0
	}

	return binary.BigEndian.Uint32(raw)
}

func (d *decoder) _u64() uint64 {
	raw := d._take(8)
	if raw == nil {
		return 0
	}

	return binary.BigEndian.Uint64(raw)
}

func (d *decoder) _length() int {
	n := int32(d._u32())
	if n < 0 {
		d._fail(fmt.Errorf("nbt: negative length %d", n))
		return 0
	}

	return int(n)
}

func (d *decoder) _string() string {
	raw := d._take(int(d._u16()))
	if raw == nil {
		return ""
	}

	return string(raw)
}

func (d *decoder) _compound() Compound {
	compound := Compound{}

	for {
		tag := TagType(d._byte())
		if d.err != nil || tag == TagEnd {
			return compound
		}

		name := d._string()
		compound[name] = d._payload(tag)
	}
}

func (d *decoder) _list() List {
	elem := TagType(d._byte())
	n := d._length()

	items := make([]Tag, 0, min(n, maxPrealloc))
	for range n {
		items = append(items, d._payload(elem))
		if d.err != nil {
			break
		}
	}

	return List{Elem: elem, Items: items}
}

func (d *decoder) _payload(tag TagType) Tag {
	switch tag {
	case TagByte:
		return Byte(d._byte())
	case TagShort:
		return Short(d._u16())
	case TagInt:
		return Int(d._u32())
	case TagLong:
		return Long(d._u64())
	case TagFloat:
		return Float(math.Float32frombits(d._u32()))
	case TagDouble:
		return Double(math.Float64frombits(d._u64()))
	case TagByteArray:
		return ByteArray(append([]byte(nil), d._take(d._length())...))
	case TagString:
		return String(d._string())
	case TagList:
		return d._list()
	case TagCompound:
		return d._compound()
	case TagIntArray:
		return d._intArray()
	case TagLongArray:
		return d._longArray()
	}

	d._fail(fmt.Errorf("nbt: unknown tag %d", tag))

	return nil
}

func (d *decoder) _intArray() IntArray {
	n := d._length()

	array := make(IntArray, 0, min(n, maxPrealloc))
	for range n {
		array = append(array, int32(d._u32()))
		if d.err != nil {
			break
		}
	}

	return array
}

func (d *decoder) _longArray() LongArray {
	n := d._length()

	array := make(LongArray, 0, min(n, maxPrealloc))
	for range n {
		array = append(array, int64(d._u64()))
		if d.err != nil {
			break
		}
	}

	return array
}
