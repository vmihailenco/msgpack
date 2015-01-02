package msgpack

import (
	"reflect"
	"strings"
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
	Index []int
	Name  string

	encoder encoderFunc
	decoder decoderFunc
}

func (f *field) EncodeValue(e *Encoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.Index)
	return f.encoder(e, fv)
}

func (f *field) DecodeValue(d *Decoder, v reflect.Value) error {
	fv := v.FieldByIndex(f.Index)
	return f.decoder(d, fv)
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
	m map[reflect.Type]map[string]*field
}

func newStructCache() *structCache {
	return &structCache{
		m: make(map[reflect.Type]map[string]*field),
	}
}

func (m *structCache) Fields(typ reflect.Type) map[string]*field {
	m.l.RLock()
	fs, ok := m.m[typ]
	m.l.RUnlock()
	if ok {
		return fs
	}

	m.l.Lock()
	fs, ok = m.m[typ]
	if !ok {
		fs = fields(typ)
		m.m[typ] = fs
	}
	m.l.Unlock()

	return fs
}

func fields(typ reflect.Type) map[string]*field {
	numField := typ.NumField()
	fs := make(map[string]*field, numField)
	for i := 0; i < numField; i++ {
		f := typ.Field(i)

		if f.Anonymous {
			typ := f.Type
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			for name, field := range fields(typ) {
				field.Index = append(f.Index, field.Index...)
				fs[name] = field
			}
			continue
		}

		if f.PkgPath != "" {
			continue
		}

		tokens := strings.Split(f.Tag.Get("msgpack"), ",")
		name := tokens[0]
		if name == "-" {
			continue
		}
		if name == "" {
			name = f.Name
		}

		field := &field{
			Index: f.Index,
			Name:  name,

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
