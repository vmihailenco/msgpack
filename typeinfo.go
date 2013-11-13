package msgpack

import (
	"reflect"
	"strings"
	"sync"
)

var structs = newStructCache()

//------------------------------------------------------------------------------

type field interface {
	Name() string

	EncodeValue(*Encoder, reflect.Value) error
	DecodeValue(*Decoder, reflect.Value) error
}

type baseField struct {
	idx  []int
	name string
}

func (f *baseField) Name() string {
	return f.name
}

func (f *baseField) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return e.EncodeValue(fv)
}

func (f *baseField) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return d.DecodeValue(fv)
}

//------------------------------------------------------------------------------

type boolField struct {
	*baseField
}

func (f *boolField) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return e.EncodeBool(fv.Bool())
}

func (f *boolField) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return d.boolValue(fv)
}

//------------------------------------------------------------------------------

type float32Field struct {
	*baseField
}

func (f *float32Field) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return e.EncodeFloat32(float32(fv.Float()))
}

func (f *float32Field) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return d.float32Value(fv)
}

//------------------------------------------------------------------------------

type float64Field struct {
	*baseField
}

func (f *float64Field) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return e.EncodeFloat64(fv.Float())
}

func (f *float64Field) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return d.float64Value(fv)
}

//------------------------------------------------------------------------------

type stringField struct {
	*baseField
}

func (f *stringField) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return e.EncodeString(fv.String())
}

func (f *stringField) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return d.stringValue(fv)
}

//------------------------------------------------------------------------------

type bytesField struct {
	*baseField
}

func (f *bytesField) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return e.EncodeBytes(fv.Bytes())
}

func (f *bytesField) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return d.bytesValue(fv)
}

//------------------------------------------------------------------------------

type intField struct {
	*baseField
}

func (f *intField) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return e.EncodeInt64(fv.Int())
}

func (f *intField) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return d.int64Value(fv)
}

//------------------------------------------------------------------------------

type uintField struct {
	*baseField
}

func (f *uintField) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return e.EncodeUint64(fv.Uint())
}

func (f *uintField) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return d.uint64Value(fv)
}

//------------------------------------------------------------------------------

type customField struct {
	*baseField
	encode encoderFunc
	decode decoderFunc
}

func (f *customField) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return f.encode(e, fv)
}

func (f *customField) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.idx)
	return f.decode(d, fv)
}

//------------------------------------------------------------------------------

type structCache struct {
	l sync.RWMutex
	m map[reflect.Type][]field
}

func newStructCache() *structCache {
	return &structCache{
		m: make(map[reflect.Type][]field),
	}
}

func (m *structCache) Fields(typ reflect.Type) []field {
	m.l.RLock()
	fields, ok := m.m[typ]
	m.l.RUnlock()
	if ok {
		return fields
	}

	numField := typ.NumField()
	fields = make([]field, 0, numField)
	for i := 0; i < numField; i++ {
		f := typ.Field(i)
		if f.PkgPath != "" {
			continue
		}
		finfo := m.newStructField(typ, &f)
		if finfo != nil {
			fields = append(fields, finfo)
		}
	}

	m.l.Lock()
	m.m[typ] = fields
	m.l.Unlock()

	return fields
}

func (m *structCache) newStructField(typ reflect.Type, f *reflect.StructField) field {
	tokens := strings.Split(f.Tag.Get("msgpack"), ",")
	name := tokens[0]
	if name == "-" {
		return nil
	} else if name == "" {
		name = f.Name
	}

	baseField := &baseField{
		idx:  f.Index,
		name: name,
	}

	ft := typ.FieldByIndex(f.Index).Type
	if encodeFunc, ok := typEncMap[ft]; ok {
		decodeFunc := typDecMap[ft]
		return &customField{
			encode:    encodeFunc,
			decode:    decodeFunc,
			baseField: baseField,
		}
	}

	switch ft.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &intField{
			baseField: baseField,
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &uintField{
			baseField: baseField,
		}
	case reflect.Bool:
		return &boolField{
			baseField: baseField,
		}
	case reflect.Float32:
		return &float32Field{
			baseField: baseField,
		}
	case reflect.Float64:
		return &float64Field{
			baseField: baseField,
		}
	case reflect.Array, reflect.Slice:
		if ft.Elem().Kind() == reflect.Uint8 {
			return &bytesField{
				baseField: baseField,
			}
		}
	case reflect.String:
		return &stringField{
			baseField: baseField,
		}
	}
	return baseField
}

func (m *structCache) Field(typ reflect.Type, name string) field {
	// TODO(vmihailenco): binary search?
	for _, f := range m.Fields(typ) {
		if f.Name() == name {
			return f
		}
	}
	return nil
}
