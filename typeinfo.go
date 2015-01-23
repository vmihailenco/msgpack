package msgpack

import (
	"io/ioutil"
	"reflect"
	"sync"
)

var (
	marshalerType   = reflect.TypeOf(new(Marshaler)).Elem()
	unmarshalerType = reflect.TypeOf(new(Unmarshaler)).Elem()
	stringsType     = reflect.TypeOf(([]string)(nil))
)

var structs = newStructCache()

var valueEncoders []encoderFunc
var valueDecoders []decoderFunc

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

func init() {
	valueEncoders = []encoderFunc{
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
		reflect.Array:         encodeArrayValue,
		reflect.Interface:     encodeInterfaceValue,
		reflect.Map:           encodeMapValue,
		reflect.Ptr:           encodePtrValue,
		reflect.Slice:         encodeSliceValue,
		reflect.String:        encodeStringValue,
		reflect.Struct:        encodeStructValue,
		reflect.UnsafePointer: nil,
	}
	valueDecoders = []decoderFunc{
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
		reflect.Array:         decodeArrayValue,
		reflect.Interface:     decodeInterfaceValue,
		reflect.Map:           decodeMapValue,
		reflect.Ptr:           decodePtrValue,
		reflect.Slice:         decodeSliceValue,
		reflect.String:        decodeStringValue,
		reflect.Struct:        decodeStructValue,
		reflect.UnsafePointer: nil,
	}
}

//------------------------------------------------------------------------------

type field struct {
	index     []int
	omitEmpty bool

	encoder encoderFunc
	decoder decoderFunc
}

func (f *field) value(strct reflect.Value) reflect.Value {
	return strct.FieldByIndex(f.index)
}

func (f *field) Omit(strct reflect.Value) bool {
	return f.omitEmpty && isEmptyValue(f.value(strct))
}

func (f *field) EncodeValue(e *Encoder, strct reflect.Value) error {
	return f.encoder(e, f.value(strct))
}

func (f *field) DecodeValue(d *Decoder, strct reflect.Value) error {
	return f.decoder(d, f.value(strct))
}

//------------------------------------------------------------------------------

type fields map[string]*field

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

func encodeSliceValue(e *Encoder, v reflect.Value) error {
	return e.encodeSlice(v)
}

func decodeSliceValue(d *Decoder, v reflect.Value) error {
	return d.sliceValue(v)
}

//------------------------------------------------------------------------------

func encodeArrayValue(e *Encoder, v reflect.Value) error {
	return e.encodeArray(v)
}

func decodeArrayValue(d *Decoder, v reflect.Value) error {
	return d.sliceValue(v)
}

//------------------------------------------------------------------------------

func encodeInterfaceValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}
	return e.EncodeValue(v.Elem())
}

func decodeInterfaceValue(d *Decoder, v reflect.Value) error {
	if v.IsNil() {
		return d.interfaceValue(v)
	}
	return d.DecodeValue(v.Elem())
}

//------------------------------------------------------------------------------

func encodeMapValue(e *Encoder, v reflect.Value) error {
	return e.encodeMap(v)
}

func decodeMapValue(d *Decoder, v reflect.Value) error {
	return d.mapValue(v)
}

//------------------------------------------------------------------------------

func encodePtrValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}
	return e.EncodeValue(v.Elem())
}

func decodePtrValue(d *Decoder, v reflect.Value) error {
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	return d.DecodeValue(v.Elem())
}

//------------------------------------------------------------------------------

func encodeStructValue(e *Encoder, v reflect.Value) error {
	return e.encodeStruct(v)
}

func decodeStructValue(d *Decoder, v reflect.Value) error {
	return d.structValue(v)
}

//------------------------------------------------------------------------------

func marshalValue(e *Encoder, v reflect.Value) error {
	marshaler := v.Interface().(Marshaler)
	b, err := marshaler.MarshalMsgpack()
	if err != nil {
		return err
	}
	_, err = e.w.Write(b)
	return err
}

func unmarshalValue(d *Decoder, v reflect.Value) error {
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}

	b, err := ioutil.ReadAll(d.r)
	if err != nil {
		return err
	}
	unmarshaler := v.Interface().(Unmarshaler)
	return unmarshaler.UnmarshalMsgpack(b)
}

//------------------------------------------------------------------------------

type structCache struct {
	l   sync.RWMutex
	m   map[reflect.Type]fields
	ext *Extensions
}

func newStructCache() *structCache {
	return &structCache{
		m: make(map[reflect.Type]fields),
	}
}

func (m *structCache) Fields(typ reflect.Type) fields {
	m.l.RLock()
	fs, ok := m.m[typ]
	m.l.RUnlock()
	if !ok {
		m.l.Lock()
		fs, ok = m.m[typ]
		if !ok {
			fs = m.getFields(typ)
			m.m[typ] = fs
		}
		m.l.Unlock()
	}

	return fs
}

func (m *structCache) getFields(typ reflect.Type) fields {
	numField := typ.NumField()
	fs := make(fields, numField)
	for i := 0; i < numField; i++ {
		f := typ.Field(i)

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

		fieldTyp := typ.FieldByIndex(f.Index).Type
		fs[name] = &field{
			index:     f.Index,
			omitEmpty: opts.Contains("omitempty"),

			encoder: m.getEncoder(fieldTyp),
			decoder: m.getDecoder(fieldTyp),
		}
	}
	return fs
}

func (m *structCache) getEncoder(typ reflect.Type) encoderFunc {
	if encoder, ok := typEncMap[typ]; ok {
		return encoder
	}
	if m.ext != nil {
		if encoder, ok := m.ext.encTypeMap[typ]; ok {
			return encoder
		}
		for iType, encoder := range m.ext.encIntMap {
			if typ.Implements(iType) {
				return encoder
			}
		}
	}

	if typ.Implements(marshalerType) {
		return marshalValue
	}

	kind := typ.Kind()
	switch kind {
	case reflect.Slice:
		elemKind := typ.Elem().Kind()
		if enc := sliceEncoders[elemKind]; enc != nil {
			return enc
		}
	}

	return valueEncoders[kind]
}

func (m *structCache) getDecoder(typ reflect.Type) decoderFunc {
	if decoder, ok := typDecMap[typ]; ok {
		return decoder
	}

	if typ.Implements(unmarshalerType) {
		return unmarshalValue
	}

	kind := typ.Kind()
	switch kind {
	case reflect.Slice:
		elemKind := typ.Elem().Kind()
		if dec := sliceDecoders[elemKind]; dec != nil {
			return dec
		}
	case reflect.Ptr:
		fallthrough
	case reflect.Struct:
		if m.ext != nil {
			if decoder, ok := m.ext.decTypeMap[typ]; ok {
				return decoder
			}
		}
	}

	return valueDecoders[kind]
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
