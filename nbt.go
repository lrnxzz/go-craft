package gocraft

import "github.com/lrnxzz/go-craft/nbt"

type NBT nbt.Compound

func (n NBT) Append(dst []byte) []byte {
	return append(dst, nbt.Encode(nbt.Compound(n))...)
}

func (n *NBT) Decode(r *Reader) error {
	if r.err != nil {
		return r.err
	}

	root, consumed, err := nbt.DecodePrefix(r.buf[r.off:])
	if err != nil {
		return r.fail(err)
	}

	*n = NBT(root)
	r.off += consumed

	return nil
}
