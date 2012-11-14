package msgpack

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"reflect"
)

func Marshal(v interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := NewEncoder(buf).Encode(v)
	return buf.Bytes(), err
}

type Encoder struct {
	W io.Writer
}

func NewEncoder(writer io.Writer) *Encoder {
	return &Encoder{
		W: writer,
	}
}

func (e *Encoder) Encode(v interface{}) error {
	if v == nil {
		return e.EncodeNil()
	}

	switch value := v.(type) {
	case string:
		return e.EncodeBytes([]byte(value))
	case []byte:
		return e.EncodeBytes(value)
	case int:
		return e.EncodeInt64(int64(value))
	case int8:
		return e.EncodeInt64(int64(value))
	case int16:
		return e.EncodeInt64(int64(value))
	case int32:
		return e.EncodeInt64(int64(value))
	case int64:
		return e.EncodeInt64(value)
	case uint:
		return e.EncodeUint64(uint64(value))
	case uint8:
		return e.EncodeUint64(uint64(value))
	case uint16:
		return e.EncodeUint64(uint64(value))
	case uint32:
		return e.EncodeUint64(uint64(value))
	case uint64:
		return e.EncodeUint64(value)
	case bool:
		return e.EncodeBool(value)
	case float32:
		return e.EncodeFloat32(value)
	case float64:
		return e.EncodeFloat64(value)
	}
	return e.EncodeValue(reflect.ValueOf(v))
}

func (e *Encoder) EncodeValue(v reflect.Value) error {
	switch v.Kind() {
	case reflect.String:
		return e.EncodeBytes([]byte(v.String()))
	case reflect.Bool:
		return e.EncodeBool(v.Bool())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return e.EncodeUint64(v.Uint())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return e.EncodeInt64(v.Int())
	case reflect.Float32:
		return e.EncodeFloat32(float32(v.Float()))
	case reflect.Float64:
		return e.EncodeFloat64(v.Float())
	case reflect.Array, reflect.Slice:
		return e.encodeArray(v)
	case reflect.Map:
		return e.encodeMap(v)
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return e.EncodeNil()
		}
		if enc, ok := typEncMap[v.Type()]; ok {
			return enc(e, v)
		}
		return e.EncodeValue(v.Elem())
	case reflect.Struct:
		if enc, ok := typEncMap[v.Type()]; ok {
			return enc(e, v)
		}
		return e.encodeStruct(v)
	default:
		return fmt.Errorf("msgpack: unsupported type %v", v.Type().String())
	}
	panic("not reached")
}

func (e *Encoder) write(data []byte) error {
	n, err := e.W.Write(data)
	if err != nil {
		return err
	}
	if n < len(data) {
		return io.ErrShortWrite
	}
	return nil
}

func (e *Encoder) EncodeNil() error {
	return e.write([]byte{nilCode})
}

func (e *Encoder) EncodeUint(v uint) error {
	return e.EncodeUint64(uint64(v))
}

func (e *Encoder) EncodeUint8(v uint8) error {
	return e.EncodeUint64(uint64(v))
}

func (e *Encoder) EncodeUint16(v uint16) error {
	return e.EncodeUint64(uint64(v))
}

func (e *Encoder) EncodeUint32(v uint32) error {
	return e.EncodeUint64(uint64(v))
}

func (e *Encoder) EncodeUint64(v uint64) error {
	switch {
	case v < 128:
		return e.write([]byte{byte(v)})
	case v < 256:
		return e.write([]byte{uint8Code, byte(v)})
	case v < 65536:
		return e.write([]byte{uint16Code, byte(v >> 8), byte(v)})
	case v < 4294967296:
		return e.write([]byte{
			uint32Code,
			byte(v >> 24),
			byte(v >> 16),
			byte(v >> 8),
			byte(v),
		})
	default:
		return e.write([]byte{
			uint64Code,
			byte(v >> 56),
			byte(v >> 48),
			byte(v >> 40),
			byte(v >> 32),
			byte(v >> 24),
			byte(v >> 16),
			byte(v >> 8),
			byte(v),
		})
	}
	panic("not reached")
}

func (e *Encoder) EncodeInt(v int) error {
	return e.EncodeInt64(int64(v))
}

func (e *Encoder) EncodeInt8(v int8) error {
	return e.EncodeInt64(int64(v))
}

func (e *Encoder) EncodeInt16(v int16) error {
	return e.EncodeInt64(int64(v))
}

func (e *Encoder) EncodeInt32(v int32) error {
	return e.EncodeInt64(int64(v))
}

func (e *Encoder) EncodeInt64(v int64) error {
	switch {
	case v < -2147483648 || v >= 2147483648:
		return e.write([]byte{
			int64Code,
			byte(v >> 56),
			byte(v >> 48),
			byte(v >> 40),
			byte(v >> 32),
			byte(v >> 24),
			byte(v >> 16),
			byte(v >> 8),
			byte(v),
		})
	case v < -32768 || v >= 32768:
		return e.write([]byte{
			int32Code,
			byte(v >> 24),
			byte(v >> 16),
			byte(v >> 8),
			byte(v),
		})
	case v < -128 || v >= 128:
		return e.write([]byte{int16Code, byte(v >> 8), byte(v)})
	case v < -32:
		return e.write([]byte{int8Code, byte(v)})
	default:
		return e.write([]byte{byte(v)})
	}
	panic("not reached")
}

func (e *Encoder) EncodeBool(value bool) error {
	if value {
		return e.write([]byte{trueCode})
	}
	return e.write([]byte{falseCode})
}

func (e *Encoder) EncodeFloat32(value float32) error {
	v := math.Float32bits(value)
	return e.write([]byte{
		floatCode,
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v),
	})
}

func (e *Encoder) EncodeFloat64(value float64) error {
	v := math.Float64bits(value)
	return e.write([]byte{
		doubleCode,
		byte(v >> 56),
		byte(v >> 48),
		byte(v >> 40),
		byte(v >> 32),
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v),
	})
}

func (e *Encoder) EncodeString(v string) error {
	return e.EncodeBytes([]byte(v))
}

func (e *Encoder) EncodeBytes(v []byte) error {
	switch l := len(v); {
	case l < 32:
		if err := e.write([]byte{fixRawLowCode | uint8(l)}); err != nil {
			return err
		}
	case l < 65536:
		if err := e.write([]byte{
			raw16Code,
			byte(l >> 8),
			byte(l),
		}); err != nil {
			return err
		}
	default:
		if err := e.write([]byte{
			raw32Code,
			byte(l >> 24),
			byte(l >> 16),
			byte(l >> 8),
			byte(l),
		}); err != nil {
			return err
		}
	}
	return e.write(v)
}

func (e *Encoder) encodeArray(value reflect.Value) error {
	elemType := value.Type().Elem()
	if elemType.Kind() == reflect.Uint8 {
		return e.EncodeBytes(value.Bytes())
	}

	l := value.Len()
	switch {
	case l < 16:
		if err := e.write([]byte{fixArrayLowCode | byte(l)}); err != nil {
			return err
		}
	case l < 65536:
		if err := e.write([]byte{array16Code, byte(l >> 8), byte(l)}); err != nil {
			return err
		}
	default:
		if err := e.write([]byte{
			array32Code,
			byte(l >> 24),
			byte(l >> 16),
			byte(l >> 8),
			byte(l),
		}); err != nil {
			return err
		}
	}
	for i := 0; i < l; i++ {
		if err := e.EncodeValue(value.Index(i)); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) encodeMap(value reflect.Value) error {
	switch l := value.Len(); {
	case l < 16:
		if err := e.write([]byte{fixMapLowCode | byte(l)}); err != nil {
			return err
		}
	case l < 65536:
		if err := e.write([]byte{
			map16Code,
			byte(l >> 8),
			byte(l),
		}); err != nil {
			return err
		}
	default:
		if err := e.write([]byte{
			map32Code,
			byte(l >> 24),
			byte(l >> 16),
			byte(l >> 8),
			byte(l),
		}); err != nil {
			return err
		}
	}
	keys := value.MapKeys()
	for _, k := range keys {
		if err := e.EncodeValue(k); err != nil {
			return err
		}
		if err := e.EncodeValue(value.MapIndex(k)); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) encodeStruct(value reflect.Value) error {
	fields := tinfoMap.Fields(value.Type())
	switch l := len(fields); {
	case l < 16:
		if err := e.write([]byte{fixMapLowCode | byte(l)}); err != nil {
			return err
		}
	case l < 65536:
		if err := e.write([]byte{
			map16Code,
			byte(l >> 8),
			byte(l),
		}); err != nil {
			return err
		}
	default:
		if err := e.write([]byte{
			map32Code,
			byte(l >> 24),
			byte(l >> 16),
			byte(l >> 8),
			byte(l),
		}); err != nil {
			return err
		}
	}
	for _, field := range fields {
		if err := e.EncodeBytes([]byte(field.name)); err != nil {
			return err
		}
		if err := e.EncodeValue(value.FieldByIndex(field.idx)); err != nil {
			return err
		}
	}
	return nil
}
