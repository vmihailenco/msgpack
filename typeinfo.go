package msgpack

import (
	"reflect"
	"strings"
	"sync"
)

var tinfoMap = newTypeInfoMap()

//------------------------------------------------------------------------------

type fieldInfo interface {
	Idx() []int
	Name() string
	EncodeValue(*Encoder, reflect.Value) error
	DecodeValue(*Decoder, reflect.Value) error
}

type defaultFieldInfo struct {
	idx  []int
	name string
}

func (f *defaultFieldInfo) Idx() []int {
	return f.idx
}

func (f *defaultFieldInfo) Name() string {
	return f.name
}

func (f *defaultFieldInfo) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return e.EncodeValue(fv)
}

func (f *defaultFieldInfo) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return d.DecodeValue(fv)
}

//------------------------------------------------------------------------------

type boolFieldInfo struct {
	*defaultFieldInfo
}

func (f *boolFieldInfo) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return e.EncodeBool(fv.Bool())
}

func (f *boolFieldInfo) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return d.boolValue(fv)
}

//------------------------------------------------------------------------------

type float32FieldInfo struct {
	*defaultFieldInfo
}

func (f *float32FieldInfo) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return e.EncodeFloat32(float32(fv.Float()))
}

func (f *float32FieldInfo) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return d.float32Value(fv)
}

//------------------------------------------------------------------------------

type float64FieldInfo struct {
	*defaultFieldInfo
}

func (f *float64FieldInfo) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return e.EncodeFloat64(fv.Float())
}

func (f *float64FieldInfo) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return d.float64Value(fv)
}

//------------------------------------------------------------------------------

type stringFieldInfo struct {
	*defaultFieldInfo
}

func (f *stringFieldInfo) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return e.EncodeString(fv.String())
}

func (f *stringFieldInfo) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return d.stringValue(fv)
}

//------------------------------------------------------------------------------

type bytesFieldInfo struct {
	*defaultFieldInfo
}

func (f *bytesFieldInfo) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return e.EncodeBytes(fv.Bytes())
}

func (f *bytesFieldInfo) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return d.bytesValue(fv)
}

//------------------------------------------------------------------------------

type intFieldInfo struct {
	*defaultFieldInfo
}

func (f *intFieldInfo) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return e.EncodeInt64(fv.Int())
}

func (f *intFieldInfo) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return d.int64Value(fv)
}

//------------------------------------------------------------------------------

type uintFieldInfo struct {
	*defaultFieldInfo
}

func (f *uintFieldInfo) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return e.EncodeUint64(fv.Uint())
}

func (f *uintFieldInfo) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return d.uint64Value(fv)
}

//------------------------------------------------------------------------------

type customFieldInfo struct {
	*defaultFieldInfo
	encode encoderFunc
	decode decoderFunc
}

func (f *customFieldInfo) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return f.encode(e, fv)
}

func (f *customFieldInfo) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return f.decode(d, fv)
}

//------------------------------------------------------------------------------

type typeInfo struct {
	fields []fieldInfo
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
	fields := make([]fieldInfo, 0, numField)
	for i := 0; i < numField; i++ {
		f := typ.Field(i)
		if f.PkgPath != "" {
			continue
		}
		finfo := m.newStructFieldInfo(typ, &f)
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

func (m *typeInfoMap) newStructFieldInfo(typ reflect.Type, f *reflect.StructField) fieldInfo {
	tokens := strings.Split(f.Tag.Get("msgpack"), ",")
	name := tokens[0]
	if name == "-" {
		return nil
	} else if name == "" {
		name = f.Name
	}

	ft := typ.FieldByIndex(f.Index).Type
	if encodeFunc, ok := typEncMap[ft]; ok {
		decodeFunc := typDecMap[ft]
		return &customFieldInfo{
			encode: encodeFunc,
			decode: decodeFunc,
			defaultFieldInfo: &defaultFieldInfo{
				idx:  f.Index,
				name: name,
			},
		}
	}

	switch ft.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &intFieldInfo{
			defaultFieldInfo: &defaultFieldInfo{
				idx:  f.Index,
				name: name,
			},
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &uintFieldInfo{
			defaultFieldInfo: &defaultFieldInfo{
				idx:  f.Index,
				name: name,
			},
		}
	case reflect.Bool:
		return &boolFieldInfo{
			defaultFieldInfo: &defaultFieldInfo{
				idx:  f.Index,
				name: name,
			},
		}
	case reflect.Float32:
		return &float32FieldInfo{
			defaultFieldInfo: &defaultFieldInfo{
				idx:  f.Index,
				name: name,
			},
		}
	case reflect.Float64:
		return &float64FieldInfo{
			defaultFieldInfo: &defaultFieldInfo{
				idx:  f.Index,
				name: name,
			},
		}
	case reflect.Array, reflect.Slice:
		if ft.Elem().Kind() == reflect.Uint8 {
			return &bytesFieldInfo{
				defaultFieldInfo: &defaultFieldInfo{
					idx:  f.Index,
					name: name,
				},
			}
		}
	case reflect.String:
		return &stringFieldInfo{
			defaultFieldInfo: &defaultFieldInfo{
				idx:  f.Index,
				name: name,
			},
		}
	}
	return &defaultFieldInfo{
		idx:  f.Index,
		name: name,
	}
}

func (m *typeInfoMap) Fields(typ reflect.Type) []fieldInfo {
	return m.TypeInfo(typ).fields
}

func (m *typeInfoMap) Field(typ reflect.Type, name string) fieldInfo {
	// TODO(vmihailenco): binary search?
	for _, f := range m.Fields(typ) {
		if f.Name() == name {
			return f
		}
	}
	return nil
}
