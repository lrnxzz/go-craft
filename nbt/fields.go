package nbt

import (
	"reflect"
	"slices"
	"strings"
	"sync"
)

type field struct {
	name      string
	index     []int
	omitempty bool
	asList    bool
}

var (
	_fieldCache    sync.Map
	_fieldMapCache sync.Map
)

func _fieldsOf(t reflect.Type) []field {
	if cached, ok := _fieldCache.Load(t); ok {
		return cached.([]field)
	}

	fields := _collectFields(t, nil)
	_fieldCache.Store(t, fields)

	return fields
}

func _fieldMapOf(t reflect.Type) map[string]field {
	if cached, ok := _fieldMapCache.Load(t); ok {
		return cached.(map[string]field)
	}

	byName := make(map[string]field, t.NumField())
	for _, f := range _fieldsOf(t) {
		byName[f.name] = f
	}
	_fieldMapCache.Store(t, byName)

	return byName
}

func _collectFields(t reflect.Type, prefix []int) []field {
	var fields []field

	for i := range t.NumField() {
		structField := t.Field(i)

		tag, tagged := structField.Tag.Lookup("nbt")
		if tag == "-" {
			continue
		}

		promoted := structField.Anonymous && !tagged && _structType(structField.Type).Kind() == reflect.Struct
		if !structField.IsExported() && !promoted {
			continue
		}

		index := append(slices.Clone(prefix), i)

		if promoted {
			fields = append(fields, _collectFields(_structType(structField.Type), index)...)
			continue
		}

		name, opts, _ := strings.Cut(tag, ",")
		if name == "" {
			name = structField.Name
		}

		options := strings.Split(opts, ",")
		fields = append(fields, field{
			name:      name,
			index:     index,
			omitempty: slices.Contains(options, "omitempty"),
			asList:    slices.Contains(options, "list"),
		})
	}

	return _resolveShadows(fields)
}

func _resolveShadows(fields []field) []field {
	depth := make(map[string]int, len(fields))
	for _, f := range fields {
		if current, ok := depth[f.name]; !ok || len(f.index) < current {
			depth[f.name] = len(f.index)
		}
	}

	resolved := fields[:0]
	for _, f := range fields {
		if len(f.index) == depth[f.name] {
			resolved = append(resolved, f)
		}
	}

	return resolved
}

func _structType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Pointer {
		return t.Elem()
	}

	return t
}

func _fieldValue(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return reflect.Value{}
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}

	return v
}

func _fieldTarget(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}

	return v
}

func _deref(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return v
		}
		v = v.Elem()
	}

	return v
}
