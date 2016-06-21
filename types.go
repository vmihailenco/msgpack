package msgpack

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"sync"
)

var marshalerType = reflect.TypeOf(new(Marshaler)).Elem()
var unmarshalerType = reflect.TypeOf(new(Unmarshaler)).Elem()

var encoderType = reflect.TypeOf(new(CustomEncoder)).Elem()
var decoderType = reflect.TypeOf(new(CustomDecoder)).Elem()

type encoderFunc func(*Encoder, reflect.Value) error
type decoderFunc func(*Decoder, reflect.Value) error

var typEncMap = make(map[reflect.Type]encoderFunc)
var typDecMap = make(map[reflect.Type]decoderFunc)

// Register registers encoder and decoder functions for a type.
// In most cases you should prefer implementing CustomEncoder and
// CustomDecoder interfaces.
func Register(typ reflect.Type, enc encoderFunc, dec decoderFunc) {
	typEncMap[typ] = enc
	typDecMap[typ] = dec
}

var structs = newStructCache()

//------------------------------------------------------------------------------

func encodeUnsupportedValue(e *Encoder, v reflect.Value) error {
	return fmt.Errorf("msgpack: Encode(unsupported %T)", v.Interface())
}

func decodeUnsupportedValue(d *Decoder, v reflect.Value) error {
	return fmt.Errorf("msgpack: Decode(unsupported %T)", v.Interface())
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

func ptrEncoderFunc(typ reflect.Type) encoderFunc {
	encoder := getEncoder(typ.Elem())
	return func(e *Encoder, v reflect.Value) error {
		if v.IsNil() {
			return e.EncodeNil()
		}
		return encoder(e, v.Elem())
	}
}

func ptrDecoderFunc(typ reflect.Type) decoderFunc {
	decoder := getDecoder(typ.Elem())
	return func(d *Decoder, v reflect.Value) error {
		if d.gotNilCode() {
			v.Set(reflect.Zero(v.Type()))
			return d.DecodeNil()
		}
		if v.IsNil() {
			if !v.CanSet() {
				return fmt.Errorf("msgpack: Decode(nonsettable %T)", v.Interface())
			}
			v.Set(reflect.New(v.Type().Elem()))
		}
		return decoder(d, v.Elem())
	}
}

//------------------------------------------------------------------------------

func encodeCustomValuePtr(e *Encoder, v reflect.Value) error {
	if !v.CanAddr() {
		return fmt.Errorf("msgpack: Encode(non-addressable %T)", v.Interface())
	}
	encoder := v.Addr().Interface().(CustomEncoder)
	return encoder.EncodeMsgpack(e)
}

func encodeCustomValue(e *Encoder, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if v.IsNil() {
			return e.EncodeNil()
		}
	}
	encoder := v.Interface().(CustomEncoder)
	return encoder.EncodeMsgpack(e)
}

func decodeCustomValuePtr(d *Decoder, v reflect.Value) error {
	if !v.CanAddr() {
		return fmt.Errorf("msgpack: Decode(nonsettable %T)", v.Interface())
	}
	if d.gotNilCode() {
		return d.DecodeNil()
	}
	decoder := v.Addr().Interface().(CustomDecoder)
	return decoder.DecodeMsgpack(d)
}

func decodeCustomValue(d *Decoder, v reflect.Value) error {
	if d.gotNilCode() {
		return d.DecodeNil()
	}
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	decoder := v.Interface().(CustomDecoder)
	return decoder.DecodeMsgpack(d)
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
	l sync.RWMutex
	m map[reflect.Type]*fields
}

func newStructCache() *structCache {
	return &structCache{
		m: make(map[reflect.Type]*fields),
	}
}

func (m *structCache) Fields(typ reflect.Type) *fields {
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

	return fs
}
