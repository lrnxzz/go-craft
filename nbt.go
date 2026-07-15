package gocraft

import "github.com/lrnxzz/go-craft/nbt"

type NBT nbt.Compound

func (n NBT) Append(dst []byte) []byte {
	if n == nil {
		return append(dst, byte(nbt.TagEnd))
	}

	return append(dst, nbt.Encode(nbt.Compound(n))...)
}

func (n *NBT) Decode(r *Reader) error {
	if r.err != nil {
		return r.err
	}
	if r.Remaining() > 0 && r.buf[r.off] == byte(nbt.TagEnd) {
		r.off++
		*n = nil

		return nil
	}

	root, consumed, err := nbt.DecodePrefix(r.buf[r.off:])
	if err != nil {
		return r.fail(err)
	}

	*n = NBT(root)
	r.off += consumed

	return nil
}
