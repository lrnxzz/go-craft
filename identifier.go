package gocraft

import (
	"fmt"
	"strings"
)

const DefaultNamespace = "minecraft"

type Identifier string

func NewIdentifier(namespace, path string) Identifier {
	return Identifier(namespace + ":" + path)
}

func (i Identifier) Append(dst []byte) []byte {
	return String(i).Append(dst)
}

func (i *Identifier) Decode(r *Reader) error {
	var raw String
	if err := raw.Decode(r); err != nil {
		return err
	}

	*i = Identifier(raw)

	return nil
}

func (i Identifier) Namespace() string {
	if namespace, _, ok := strings.Cut(string(i), ":"); ok {
		return namespace
	}

	return DefaultNamespace
}

func (i Identifier) Path() string {
	if _, path, ok := strings.Cut(string(i), ":"); ok {
		return path
	}

	return string(i)
}

func (i Identifier) Valid() bool {
	namespace, path := i.Namespace(), i.Path()
	if path == "" {
		return false
	}

	return _validSegment(namespace, _validNamespaceByte) && _validSegment(path, _validPathByte)
}

func (i Identifier) String() string {
	return i.Namespace() + ":" + i.Path()
}

func _validSegment(segment string, allow func(byte) bool) bool {
	if segment == "" {
		return false
	}

	for i := range len(segment) {
		if !allow(segment[i]) {
			return false
		}
	}

	return true
}

func _validNamespaceByte(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= '0' && b <= '9' || b == '_' || b == '-' || b == '.'
}

func _validPathByte(b byte) bool {
	return _validNamespaceByte(b) || b == '/'
}

var (
	_ Field        = Identifier("")
	_ FieldPtr     = (*Identifier)(nil)
	_ fmt.Stringer = Identifier("")
)
