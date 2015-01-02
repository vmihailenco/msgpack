package msgpack

import (
	"reflect"
	"sync"
)

var stringsType = reflect.TypeOf(([]string)(nil))

var structs = newStructCache()

var valueEncoders = [...]encoderFunc{
	reflect.Bool:          encodeBoolValue,
	reflect.Int:           encodeInt64Value,
	reflect.Int8:          encodeInt64Value,
	reflect.Int16:         encodeInt64Value,
	reflect.Int32:         encodeInt64Value,
	reflect.Int64:         encodeInt64Value,
	reflect.Uint:          encodeUint64Value,
	reflect.Uint8:         encodeUint64Value,
	reflect.Uint16:        encodeUint64Value,
	reflect.Uint32:        encodeUint64Value,
	reflect.Uint64:        encodeUint64Value,
	reflect.Float32:       encodeFloat64Value,
	reflect.Float64:       encodeFloat64Value,
	reflect.String:        encodeStringValue,
	reflect.UnsafePointer: nil,
}

var valueDecoders = [...]decoderFunc{
	reflect.Bool:          decodeBoolValue,
	reflect.Int:           decodeInt64Value,
	reflect.Int8:          decodeInt64Value,
	reflect.Int16:         decodeInt64Value,
	reflect.Int32:         decodeInt64Value,
	reflect.Int64:         decodeInt64Value,
	reflect.Uint:          decodeUint64Value,
	reflect.Uint8:         decodeUint64Value,
	reflect.Uint16:        decodeUint64Value,
	reflect.Uint32:        decodeUint64Value,
	reflect.Uint64:        decodeUint64Value,
	reflect.Float32:       decodeFloat64Value,
	reflect.Float64:       decodeFloat64Value,
	reflect.String:        decodeStringValue,
	reflect.UnsafePointer: nil,
}

var sliceEncoders = [...]encoderFunc{
	reflect.Uint8:         encodeBytesValue,
	reflect.String:        encodeStringsValue,
	reflect.UnsafePointer: nil,
}

var sliceDecoders = []decoderFunc{
	reflect.Uint8:         decodeBytesValue,
	reflect.String:        decodeStringsValue,
	reflect.UnsafePointer: nil,
}

//------------------------------------------------------------------------------

type field struct {
	index     []int
	omitEmpty bool

	v reflect.Value

	encoder encoderFunc
	decoder decoderFunc
}

func (f *field) setStruct(strct reflect.Value) {
	f.v = strct.FieldByIndex(f.index)
}

func (f *field) Omit() bool {
	return f.omitEmpty && isEmptyValue(f.v)
}

func (f *field) EncodeValue(e *Encoder) error {
	return f.encoder(e, f.v)
}

func (f *field) DecodeValue(d *Decoder) error {
	return f.decoder(d, f.v)
}

//------------------------------------------------------------------------------

type fields map[string]*field

func (fs fields) setStruct(strct reflect.Value) {
	for _, f := range fs {
		f.setStruct(strct)
	}
}

func (fs fields) Len() (length int) {
	for _, f := range fs {
		if !f.Omit() {
			length++
		}
	}
	return length
}

//------------------------------------------------------------------------------

func encodeValue(e *Encoder, v reflect.Value) error {
	return e.EncodeValue(v)
}

func decodeValue(d *Decoder, v reflect.Value) error {
	return d.DecodeValue(v)
}

//------------------------------------------------------------------------------

func encodeBoolValue(e *Encoder, v reflect.Value) error {
	return e.EncodeBool(v.Bool())
}

func decodeBoolValue(d *Decoder, v reflect.Value) error {
	return d.boolValue(v)
}

//------------------------------------------------------------------------------

func encodeFloat64Value(e *Encoder, v reflect.Value) error {
	return e.EncodeFloat64(v.Float())
}

func decodeFloat64Value(d *Decoder, v reflect.Value) error {
	return d.float64Value(v)
}

//------------------------------------------------------------------------------

func encodeStringValue(e *Encoder, v reflect.Value) error {
	return e.EncodeString(v.String())
}

func decodeStringValue(d *Decoder, v reflect.Value) error {
	return d.stringValue(v)
}

//------------------------------------------------------------------------------

func encodeBytesValue(e *Encoder, v reflect.Value) error {
	return e.EncodeBytes(v.Bytes())
}

func decodeBytesValue(d *Decoder, v reflect.Value) error {
	return d.bytesValue(v)
}

//------------------------------------------------------------------------------

func encodeStringsValue(e *Encoder, v reflect.Value) error {
	return e.encodeStringSlice(v.Convert(stringsType).Interface().([]string))
}

func decodeStringsValue(d *Decoder, v reflect.Value) error {
	return d.stringsValue(v)
}

//------------------------------------------------------------------------------

func encodeInt64Value(e *Encoder, v reflect.Value) error {
	return e.EncodeInt64(v.Int())
}

func decodeInt64Value(d *Decoder, v reflect.Value) error {
	return d.int64Value(v)
}

//------------------------------------------------------------------------------

func encodeUint64Value(e *Encoder, v reflect.Value) error {
	return e.EncodeUint64(v.Uint())
}

func decodeUint64Value(d *Decoder, v reflect.Value) error {
	return d.uint64Value(v)
}

//------------------------------------------------------------------------------

type structCache struct {
	l sync.RWMutex
	m map[reflect.Type]fields
}

func newStructCache() *structCache {
	return &structCache{
		m: make(map[reflect.Type]fields),
	}
}

func (m *structCache) Fields(strct reflect.Value) fields {
	typ := strct.Type()

	m.l.RLock()
	fs, ok := m.m[typ]
	m.l.RUnlock()
	if !ok {
		m.l.Lock()
		fs, ok = m.m[typ]
		if !ok {
			fs = getFields(typ)
			m.m[typ] = fs
		}
		m.l.Unlock()
	}

	fs.setStruct(strct)
	return fs
}

func getFields(typ reflect.Type) fields {
	numField := typ.NumField()
	fs := make(fields, numField)
	for i := 0; i < numField; i++ {
		f := typ.Field(i)

		if f.Anonymous {
			typ := f.Type
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			for name, field := range getFields(typ) {
				field.index = append(f.Index, field.index...)
				fs[name] = field
			}
			continue
		}

		if f.PkgPath != "" {
			continue
		}

		name, opts := parseTag(f.Tag.Get("msgpack"))
		if name == "-" {
			continue
		}
		if name == "" {
			name = f.Name
		}

		field := &field{
			index:     f.Index,
			omitEmpty: opts.Contains("omitempty"),

			encoder: encodeValue,
			decoder: decodeValue,
		}

		ft := typ.FieldByIndex(f.Index).Type
		if encoder, ok := typEncMap[ft]; ok {
			decoder := typDecMap[ft]
			field.encoder = encoder
			field.decoder = decoder
			fs[name] = field
			continue
		}

		kind := ft.Kind()
		if kind == reflect.Slice {
			kind = ft.Elem().Kind()
			if enc := sliceEncoders[kind]; enc != nil {
				field.encoder = enc
			}
			if dec := sliceDecoders[kind]; dec != nil {
				field.decoder = dec
			}
		} else {
			if enc := valueEncoders[kind]; enc != nil {
				field.encoder = enc
			}
			if dec := valueDecoders[kind]; dec != nil {
				field.decoder = dec
			}
		}

		fs[name] = field
	}
	return fs
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
