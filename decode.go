package msgpack

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"time"
)

func Unmarshal(data []byte, v interface{}) error {
	buf := bytes.NewBuffer(data)
	return NewDecoder(buf).Decode(v)
}

type BufferedReader interface {
	Read([]byte) (int, error)
	Peek(int) ([]byte, error)
	ReadByte() (byte, error)
	UnreadByte() error
}

type Decoder struct {
	R                  BufferedReader
	DecodeMapFunc      func(*Decoder) (interface{}, error)
	b1, b2, b3, b4, b8 []byte
}

func NewDecoder(reader io.Reader) *Decoder {
	b := make([]byte, 8)
	r, ok := reader.(BufferedReader)
	if !ok {
		r = bufio.NewReader(reader)
	}
	return &Decoder{
		R:             r,
		DecodeMapFunc: DecodeMap,
		b1:            b[:1],
		b2:            b[:2],
		b4:            b[:4],
		b8:            b[:8],
	}
}

func (d *Decoder) Decode(iv interface{}) error {
	var err error
	switch v := iv.(type) {
	case *string:
		*v, err = d.DecodeString()
		return err
	case *[]byte:
		*v, err = d.DecodeBytes()
		return err
	case *int:
		*v, err = d.DecodeInt()
		return err
	case *int8:
		*v, err = d.DecodeInt8()
		return err
	case *int16:
		*v, err = d.DecodeInt16()
		return err
	case *int32:
		*v, err = d.DecodeInt32()
		return err
	case *int64:
		*v, err = d.DecodeInt64()
		return err
	case *uint:
		*v, err = d.DecodeUint()
		return err
	case *uint8:
		*v, err = d.DecodeUint8()
		return err
	case *uint16:
		*v, err = d.DecodeUint16()
		return err
	case *uint32:
		*v, err = d.DecodeUint32()
		return err
	case *uint64:
		*v, err = d.DecodeUint64()
		return err
	case *bool:
		*v, err = d.DecodeBool()
		return err
	case *float32:
		*v, err = d.DecodeFloat32()
		return err
	case *float64:
		*v, err = d.DecodeFloat64()
		return err
	case *[]string:
		return d.decodeIntoStrings(v)
	case *map[string]string:
		return d.decodeIntoMapStringString(v)
	case *time.Duration:
		vv, err := d.DecodeInt64()
		*v = time.Duration(vv)
		return err
	case *time.Time:
		*v, err = d.DecodeTime()
		return err
	}

	v := reflect.ValueOf(iv)
	if !v.IsValid() {
		return errors.New("msgpack: Decode(" + v.String() + ")")
	}
	if v.Kind() != reflect.Ptr {
		return errors.New("msgpack: pointer expected")
	}
	return d.DecodeValue(v)
}

func (d *Decoder) DecodeMulti(values ...interface{}) error {
	for _, v := range values {
		if err := d.Decode(v); err != nil {
			return err
		}
	}
	return nil
}

func (d *Decoder) DecodeValue(v reflect.Value) error {
	c, err := d.R.ReadByte()
	if err != nil {
		return err
	}
	if c == nilCode {
		return nil
	}
	if err := d.R.UnreadByte(); err != nil {
		return err
	}

	switch v.Kind() {
	case reflect.Bool:
		return d.boolValue(v)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return d.uint64Value(v)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return d.int64Value(v)
	case reflect.Float32:
		return d.float32Value(v)
	case reflect.Float64:
		return d.float64Value(v)
	case reflect.String:
		return d.stringValue(v)
	case reflect.Array, reflect.Slice:
		return d.sliceValue(v)
	case reflect.Map:
		return d.mapValue(v)
	case reflect.Struct:
		typ := v.Type()
		if dec, ok := typDecMap[typ]; ok {
			return dec(d, v)
		}
		if dec, ok := v.Interface().(decoder); ok {
			return dec.DecodeMsgpack(d.R)
		}
		return d.structValue(v)
	case reflect.Ptr:
		typ := v.Type()
		if v.IsNil() {
			v.Set(reflect.New(typ.Elem()))
		}
		if dec, ok := typDecMap[typ]; ok {
			return dec(d, v)
		}
		if dec, ok := v.Interface().(decoder); ok {
			return dec.DecodeMsgpack(d.R)
		}
		return d.DecodeValue(v.Elem())
	case reflect.Interface:
		if v.IsNil() {
			return d.interfaceValue(v)
		} else {
			return d.DecodeValue(v.Elem())
		}
	}
	return fmt.Errorf("msgpack: unsupported type %v", v.Type().String())
}

func (d *Decoder) DecodeBool() (bool, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return false, err
	}
	switch c {
	case falseCode:
		return false, nil
	case trueCode:
		return true, nil
	}
	return false, fmt.Errorf("msgpack: invalid code %x decoding bool", c)
}

func (d *Decoder) boolValue(value reflect.Value) error {
	v, err := d.DecodeBool()
	if err != nil {
		return err
	}
	value.SetBool(v)
	return nil
}

func (d *Decoder) uint16() (uint16, error) {
	if err := d.read(d.b2); err != nil {
		return 0, err
	}
	return (uint16(d.b2[0]) << 8) | uint16(d.b2[1]), nil
}

func (d *Decoder) uint32() (uint32, error) {
	if err := d.read(d.b4); err != nil {
		return 0, err
	}
	n := (uint32(d.b4[0]) << 24) |
		(uint32(d.b4[1]) << 16) |
		(uint32(d.b4[2]) << 8) |
		uint32(d.b4[3])
	return n, nil
}

func (d *Decoder) uint64() (uint64, error) {
	if err := d.read(d.b8); err != nil {
		return 0, err
	}
	n := (uint64(d.b8[0]) << 56) |
		(uint64(d.b8[1]) << 48) |
		(uint64(d.b8[2]) << 40) |
		(uint64(d.b8[3]) << 32) |
		(uint64(d.b8[4]) << 24) |
		(uint64(d.b8[5]) << 16) |
		(uint64(d.b8[6]) << 8) |
		uint64(d.b8[7])
	return n, nil
}

func (d *Decoder) DecodeUint64() (uint64, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode {
		return uint64(c), nil
	}
	switch c {
	case uint8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return uint64(c), nil
	case uint16Code:
		if err := d.read(d.b2); err != nil {
			return 0, err
		}
		return (uint64(d.b2[0]) << 8) | uint64(d.b2[1]), nil
	case uint32Code:
		if err := d.read(d.b4); err != nil {
			return 0, err
		}
		v := (uint64(d.b4[0]) << 24) |
			(uint64(d.b4[1]) << 16) |
			(uint64(d.b4[2]) << 8) |
			uint64(d.b4[3])
		return v, nil
	case uint64Code:
		if err := d.read(d.b8); err != nil {
			return 0, err
		}
		v := (uint64(d.b8[0]) << 56) |
			(uint64(d.b8[1]) << 48) |
			(uint64(d.b8[2]) << 40) |
			(uint64(d.b8[3]) << 32) |
			(uint64(d.b8[4]) << 24) |
			(uint64(d.b8[5]) << 16) |
			(uint64(d.b8[6]) << 8) |
			uint64(d.b8[7])
		return v, nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding uint64", c)
}

func (d *Decoder) uint64Value(value reflect.Value) error {
	v, err := d.DecodeUint64()
	if err != nil {
		return err
	}
	value.SetUint(v)
	return nil
}

func (d *Decoder) DecodeInt64() (int64, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode || c >= negFixNumLowCode {
		return int64(int8(c)), nil
	}
	switch c {
	case int8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return int64(int8(c)), nil
	case int16Code:
		if err := d.read(d.b2); err != nil {
			return 0, err
		}
		return int64((int16(d.b2[0]) << 8) | int16(d.b2[1])), nil
	case int32Code:
		if err := d.read(d.b4); err != nil {
			return 0, err
		}
		v := int64((int32(d.b4[0]) << 24) |
			(int32(d.b4[1]) << 16) |
			(int32(d.b4[2]) << 8) |
			int32(d.b4[3]))
		return v, nil
	case int64Code:
		if err := d.read(d.b8); err != nil {
			return 0, err
		}
		v := (int64(d.b8[0]) << 56) |
			(int64(d.b8[1]) << 48) |
			(int64(d.b8[2]) << 40) |
			(int64(d.b8[3]) << 32) |
			(int64(d.b8[4]) << 24) |
			(int64(d.b8[5]) << 16) |
			(int64(d.b8[6]) << 8) |
			int64(d.b8[7])
		return v, nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding int64", c)
}

func (d *Decoder) int64Value(value reflect.Value) error {
	v, err := d.DecodeInt64()
	if err != nil {
		return err
	}
	value.SetInt(v)
	return nil
}

func (d *Decoder) DecodeFloat32() (float32, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c != floatCode {
		return 0, fmt.Errorf("msgpack: invalid code %x decoding float32", c)
	}
	b, err := d.uint32()
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(b), nil
}

func (d *Decoder) float32Value(value reflect.Value) error {
	v, err := d.DecodeFloat32()
	if err != nil {
		return err
	}
	value.SetFloat(float64(v))
	return nil
}

func (d *Decoder) DecodeFloat64() (float64, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c != doubleCode {
		return 0, fmt.Errorf("msgpack: invalid code %x decoding float64", c)
	}
	b, err := d.uint64()
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(b), nil
}

func (d *Decoder) float64Value(value reflect.Value) error {
	v, err := d.DecodeFloat64()
	if err != nil {
		return err
	}
	value.SetFloat(v)
	return nil
}

func (d *Decoder) structLen() (int, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c >= fixMapLowCode && c <= fixMapHighCode {
		return int(c & fixMapMask), nil
	}
	switch c {
	case map16Code:
		n, err := d.uint16()
		return int(n), err
	case map32Code:
		n, err := d.uint32()
		return int(n), err
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding struct length", c)
}

func (d *Decoder) structValue(v reflect.Value) error {
	n, err := d.structLen()
	if err != nil {
		return err
	}

	typ := v.Type()
	for i := 0; i < n; i++ {
		name, err := d.DecodeString()
		if err != nil {
			return err
		}

		f := tinfoMap.Field(typ, name)
		if f == nil {
			continue
		}

		if err := d.DecodeValue(v.FieldByIndex(f.idx)); err != nil {
			return fmt.Errorf("msgpack: can't decode value for field %q: %v", name, err)
		}
	}
	return nil
}

func (d *Decoder) read(b []byte) error {
	_, err := io.ReadFull(d.R, b)
	return err
}

//------------------------------------------------------------------------------

func (d *Decoder) DecodeUint() (uint, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode {
		return uint(c), nil
	}
	switch c {
	case uint8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return uint(c), nil
	case uint16Code:
		if err := d.read(d.b2); err != nil {
			return 0, err
		}
		return (uint(d.b2[0]) << 8) | uint(d.b2[1]), nil
	case uint32Code:
		if err := d.read(d.b4); err != nil {
			return 0, err
		}
		v := (uint(d.b4[0]) << 24) |
			(uint(d.b4[1]) << 16) |
			(uint(d.b4[2]) << 8) |
			uint(d.b4[3])
		return v, nil
	case uint64Code:
		if err := d.read(d.b8); err != nil {
			return 0, err
		}
		v := (uint(d.b8[0]) << 56) |
			(uint(d.b8[1]) << 48) |
			(uint(d.b8[2]) << 40) |
			(uint(d.b8[3]) << 32) |
			(uint(d.b8[4]) << 24) |
			(uint(d.b8[5]) << 16) |
			(uint(d.b8[6]) << 8) |
			uint(d.b8[7])
		return v, nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding uint", c)
}

func (d *Decoder) DecodeUint8() (uint8, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode {
		return uint8(c), nil
	}
	switch c {
	case uint8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return uint8(c), nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding uint8", c)
}

func (d *Decoder) DecodeUint16() (uint16, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode {
		return uint16(c), nil
	}
	switch c {
	case uint8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return uint16(c), nil
	case uint16Code:
		if err := d.read(d.b2); err != nil {
			return 0, err
		}
		return (uint16(d.b2[0]) << 8) | uint16(d.b2[1]), nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding uint16", c)
}

func (d *Decoder) DecodeUint32() (uint32, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode {
		return uint32(c), nil
	}
	switch c {
	case uint8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return uint32(c), nil
	case uint16Code:
		if err := d.read(d.b2); err != nil {
			return 0, err
		}
		return (uint32(d.b2[0]) << 8) | uint32(d.b2[1]), nil
	case uint32Code:
		if err := d.read(d.b4); err != nil {
			return 0, err
		}
		v := (uint32(d.b4[0]) << 24) |
			(uint32(d.b4[1]) << 16) |
			(uint32(d.b4[2]) << 8) |
			uint32(d.b4[3])
		return v, nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding uint32", c)
}

//------------------------------------------------------------------------------

func (d *Decoder) DecodeInt() (int, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode || c >= negFixNumLowCode {
		return int(int8(c)), nil
	}
	switch c {
	case int8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return int(int8(c)), nil
	case int16Code:
		if err := d.read(d.b2); err != nil {
			return 0, err
		}
		return int((int16(d.b2[0]) << 8) | int16(d.b2[1])), nil
	case int32Code:
		if err := d.read(d.b4); err != nil {
			return 0, err
		}
		v := int((int32(d.b4[0]) << 24) |
			(int32(d.b4[1]) << 16) |
			(int32(d.b4[2]) << 8) |
			int32(d.b4[3]))
		return v, nil
	case int64Code:
		if err := d.read(d.b8); err != nil {
			return 0, err
		}
		v := int((int64(d.b8[0]) << 56) |
			(int64(d.b8[1]) << 48) |
			(int64(d.b8[2]) << 40) |
			(int64(d.b8[3]) << 32) |
			(int64(d.b8[4]) << 24) |
			(int64(d.b8[5]) << 16) |
			(int64(d.b8[6]) << 8) |
			int64(d.b8[7]))
		return v, nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding int64", c)
}

func (d *Decoder) DecodeInt8() (int8, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode || c >= negFixNumLowCode {
		return int8(c), nil
	}
	switch c {
	case int8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return int8(c), nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding int8", c)
}

func (d *Decoder) DecodeInt16() (int16, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode || c >= negFixNumLowCode {
		return int16(int8(c)), nil
	}
	switch c {
	case int8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return int16(int8(c)), nil
	case int16Code:
		if err := d.read(d.b2); err != nil {
			return 0, err
		}
		return (int16(d.b2[0]) << 8) | int16(d.b2[1]), nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding int16", c)
}

func (d *Decoder) DecodeInt32() (int32, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode || c >= negFixNumLowCode {
		return int32(int8(c)), nil
	}
	switch c {
	case int8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return int32(int8(c)), nil
	case int16Code:
		if err := d.read(d.b2); err != nil {
			return 0, err
		}
		return int32((int16(d.b2[0]) << 8) | int16(d.b2[1])), nil
	case int32Code:
		if err := d.read(d.b4); err != nil {
			return 0, err
		}
		v := (int32(d.b4[0]) << 24) |
			(int32(d.b4[1]) << 16) |
			(int32(d.b4[2]) << 8) |
			int32(d.b4[3])
		return v, nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding int32", c)
}

//------------------------------------------------------------------------------

func (d *Decoder) interfaceValue(v reflect.Value) error {
	iface, err := d.DecodeInterface()
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(iface))
	return nil
}

// Decodes value into interface. Possible value types are:
//   - nil,
//   - int64,
//   - uint64,
//   - bool,
//   - float32 and float64,
//   - string,
//   - slices of any of the above,
//   - maps of any of the above.
func (d *Decoder) DecodeInterface() (interface{}, error) {
	b, err := d.R.Peek(1)
	if err != nil {
		return nil, err
	}
	c := b[0]

	if c <= posFixNumHighCode || c >= negFixNumLowCode {
		return d.DecodeInt64()
	} else if c >= fixMapLowCode && c <= fixMapHighCode {
		return d.DecodeMap()
	} else if c >= fixArrayLowCode && c <= fixArrayHighCode {
		return d.DecodeSlice()
	} else if c >= fixRawLowCode && c <= fixRawHighCode {
		return d.DecodeString()
	}

	switch c {
	case nilCode:
		_, err := d.R.ReadByte()
		return nil, err
	case falseCode, trueCode:
		return d.DecodeBool()
	case floatCode:
		return d.DecodeFloat32()
	case doubleCode:
		return d.DecodeFloat64()
	case uint8Code, uint16Code, uint32Code, uint64Code:
		return d.DecodeUint64()
	case int8Code, int16Code, int32Code, int64Code:
		return d.DecodeInt64()
	case raw16Code, raw32Code:
		return d.DecodeString()
	case array16Code, array32Code:
		return d.DecodeSlice()
	case map16Code, map32Code:
		return d.DecodeMap()
	}

	return 0, fmt.Errorf("msgpack: invalid code %x decoding interface{}", c)
}
