package msgpack

import (
	"reflect"
	"sync"
)

type fieldInfo struct {
	Ind  int
	Name string
}

type fieldsInfo struct {
	lock   sync.RWMutex
	fields map[reflect.Type][]*fieldInfo
}

var reflectCache = &fieldsInfo{
	fields: make(map[reflect.Type][]*fieldInfo),
}

func (i *fieldsInfo) Fields(typ reflect.Type) []*fieldInfo {
	i.lock.RLock()
	if fields, ok := i.fields[typ]; ok {
		i.lock.RUnlock()
		return fields
	}
	i.lock.RUnlock()

	num := typ.NumField()
	fields := make([]*fieldInfo, 0, num)
	for i := 0; i < num; i++ {
		fieldType := typ.Field(i)
		if fieldType.PkgPath != "" {
			continue
		}
		fields = append(fields, &fieldInfo{Ind: i, Name: fieldType.Name})
	}

	i.lock.Lock()
	i.fields[typ] = fields
	i.lock.Unlock()

	return fields
}

func (i *fieldsInfo) Field(typ reflect.Type, name string) *fieldInfo {
	for _, f := range i.Fields(typ) {
		if f.Name == name {
			return f
		}
	}
	return nil
}
