package gocraft

import (
	"errors"
	"fmt"
	"io"
	"math/bits"
	"unsafe"
)

const (
	segmentBits byte = 0x7F
	continueBit byte = 0x80
	segmentLen  int  = 7
	bitsPerByte int  = 8
)

type varint interface {
	~int32 | ~int64
}

func _widthOf[T varint]() int {
	return int(unsafe.Sizeof(T(0))) * bitsPerByte
}

func _maxLenOf[T varint]() int {
	return (_widthOf[T]() + segmentLen - 1) / segmentLen
}

func _toUnsigned[T varint](v T) uint64 {
	if _widthOf[T]() == 32 {
		return uint64(uint32(v))
	}
	return uint64(v)
}

func _fromUnsigned[T varint](u uint64) T {
	if _widthOf[T]() == 32 {
		return T(int32(uint32(u)))
	}
	return T(int64(u))
}

func AppendVar[T varint](dst []byte, v T) []byte {
	u := _toUnsigned(v)
	for u > uint64(segmentBits) {
		dst = append(dst, byte(u)&segmentBits|continueBit)
		u >>= uint(segmentLen)
	}
	return append(dst, byte(u))
}

func ReadVar[T varint](r io.ByteReader) (T, error) {
	var u uint64
	for i := range _maxLenOf[T]() {
		b, err := r.ReadByte()
		if err != nil {
			if i > 0 && errors.Is(err, io.EOF) {
				err = io.ErrUnexpectedEOF
			}
			return 0, err
		}
		u |= uint64(b&segmentBits) << (i * segmentLen)
		if b&continueBit == 0 {
			return _fromUnsigned[T](u), nil
		}
	}
	return 0, fmt.Errorf("gocraft: variable-length integer exceeds %d bytes", _maxLenOf[T]())
}

func VarLen[T varint](v T) int {
	u := _toUnsigned(v)
	return (bits.Len64(u|1) + segmentLen - 1) / segmentLen
}
