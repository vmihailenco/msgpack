package msgpack

import (
	"reflect"
	"strings"
	"sync"
)

var tinfoMap = newTypeInfoMap()

//------------------------------------------------------------------------------

type fieldInfo struct {
	idx  []int
	name string
}

type typeInfo struct {
	fields []*fieldInfo
}

type typeInfoMap struct {
	l sync.RWMutex
	m map[reflect.Type]*typeInfo
}

func newTypeInfoMap() *typeInfoMap {
	return &typeInfoMap{
		m: make(map[reflect.Type]*typeInfo),
	}
}

func (m *typeInfoMap) TypeInfo(typ reflect.Type) *typeInfo {
	m.l.RLock()
	tinfo, ok := m.m[typ]
	m.l.RUnlock()
	if ok {
		return tinfo
	}

	numField := typ.NumField()
	fields := make([]*fieldInfo, 0, numField)
	for i := 0; i < numField; i++ {
		f := typ.Field(i)
		if f.PkgPath != "" {
			continue
		}
		finfo := m.structFieldInfo(typ, &f)
		if finfo != nil {
			fields = append(fields, finfo)
		}
	}
	tinfo = &typeInfo{fields: fields}

	m.l.Lock()
	m.m[typ] = tinfo
	m.l.Unlock()

	return tinfo
}

func (m *typeInfoMap) structFieldInfo(typ reflect.Type, f *reflect.StructField) *fieldInfo {
	tokens := strings.Split(f.Tag.Get("msgpack"), ",")
	name := tokens[0]
	if name == "-" {
		return nil
	} else if name == "" {
		name = f.Name
	}

	return &fieldInfo{
		idx:  f.Index,
		name: name,
	}
}

func (m *typeInfoMap) Fields(typ reflect.Type) []*fieldInfo {
	return m.TypeInfo(typ).fields
}

func (m *typeInfoMap) Field(typ reflect.Type, name string) *fieldInfo {
	for _, f := range m.Fields(typ) {
		if f.name == name {
			return f
		}
	}
	return nil
}
