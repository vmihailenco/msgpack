package msgpack

import (
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"time"

	"github.com/vmihailenco/bufio"
)

type bufReader interface {
	Read([]byte) (int, error)
	ReadByte() (byte, error)
	UnreadByte() error
	Peek(int) ([]byte, error)
	ReadN(int) ([]byte, error)
}

func Unmarshal(data []byte, v ...interface{}) error {
	buf := bufio.NewBuffer(data)
	return NewDecoder(buf).Decode(v...)
}

type Decoder struct {
	R             bufReader
	DecodeMapFunc func(*Decoder) (interface{}, error)
}

func NewDecoder(rd io.Reader) *Decoder {
	brd, ok := rd.(bufReader)
	if !ok {
		brd = bufio.NewReader(rd)
	}
	return &Decoder{
		R:             brd,
		DecodeMapFunc: decodeMap,
	}
}

func (d *Decoder) Decode(v ...interface{}) error {
	for _, vv := range v {
		if err := d.decode(vv); err != nil {
			return err
		}
	}
	return nil
}

func (d *Decoder) decode(iv interface{}) error {
	var err error
	switch v := iv.(type) {
	case *string:
		if v != nil {
			*v, err = d.DecodeString()
			return err
		}
	case *[]byte:
		if v != nil {
			*v, err = d.DecodeBytes()
			return err
		}
	case *int:
		if v != nil {
			*v, err = d.DecodeInt()
			return err
		}
	case *int8:
		if v != nil {
			*v, err = d.DecodeInt8()
			return err
		}
	case *int16:
		if v != nil {
			*v, err = d.DecodeInt16()
			return err
		}
	case *int32:
		if v != nil {
			*v, err = d.DecodeInt32()
			return err
		}
	case *int64:
		if v != nil {
			*v, err = d.DecodeInt64()
			return err
		}
	case *uint:
		if v != nil {
			*v, err = d.DecodeUint()
			return err
		}
	case *uint8:
		if v != nil {
			*v, err = d.DecodeUint8()
			return err
		}
	case *uint16:
		if v != nil {
			*v, err = d.DecodeUint16()
			return err
		}
	case *uint32:
		if v != nil {
			*v, err = d.DecodeUint32()
			return err
		}
	case *uint64:
		if v != nil {
			*v, err = d.DecodeUint64()
			return err
		}
	case *bool:
		if v != nil {
			*v, err = d.DecodeBool()
			return err
		}
	case *float32:
		if v != nil {
			*v, err = d.DecodeFloat32()
			return err
		}
	case *float64:
		if v != nil {
			*v, err = d.DecodeFloat64()
			return err
		}
	case *[]string:
		return d.decodeIntoStrings(v)
	case *map[string]string:
		return d.decodeIntoMapStringString(v)
	case *time.Duration:
		if v != nil {
			vv, err := d.DecodeInt64()
			*v = time.Duration(vv)
			return err
		}
	case *time.Time:
		if v != nil {
			*v, err = d.DecodeTime()
			return err
		}
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
	b, err := d.R.ReadN(2)
	if err != nil {
		return 0, err
	}
	return (uint16(b[0]) << 8) | uint16(b[1]), nil
}

func (d *Decoder) uint32() (uint32, error) {
	b, err := d.R.ReadN(4)
	if err != nil {
		return 0, err
	}
	n := (uint32(b[0]) << 24) |
		(uint32(b[1]) << 16) |
		(uint32(b[2]) << 8) |
		uint32(b[3])
	return n, nil
}

func (d *Decoder) uint64() (uint64, error) {
	b, err := d.R.ReadN(8)
	if err != nil {
		return 0, err
	}
	n := (uint64(b[0]) << 56) |
		(uint64(b[1]) << 48) |
		(uint64(b[2]) << 40) |
		(uint64(b[3]) << 32) |
		(uint64(b[4]) << 24) |
		(uint64(b[5]) << 16) |
		(uint64(b[6]) << 8) |
		uint64(b[7])
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
		b, err := d.R.ReadN(2)
		if err != nil {
			return 0, err
		}
		return (uint64(b[0]) << 8) | uint64(b[1]), nil
	case uint32Code:
		b, err := d.R.ReadN(4)
		if err != nil {
			return 0, err
		}
		v := (uint64(b[0]) << 24) |
			(uint64(b[1]) << 16) |
			(uint64(b[2]) << 8) |
			uint64(b[3])
		return v, nil
	case uint64Code:
		b, err := d.R.ReadN(8)
		if err != nil {
			return 0, err
		}
		v := (uint64(b[0]) << 56) |
			(uint64(b[1]) << 48) |
			(uint64(b[2]) << 40) |
			(uint64(b[3]) << 32) |
			(uint64(b[4]) << 24) |
			(uint64(b[5]) << 16) |
			(uint64(b[6]) << 8) |
			uint64(b[7])
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
		b, err := d.R.ReadN(2)
		if err != nil {
			return 0, err
		}
		return int64((int16(b[0]) << 8) | int16(b[1])), nil
	case int32Code:
		b, err := d.R.ReadN(4)
		if err != nil {
			return 0, err
		}
		v := int64((int32(b[0]) << 24) |
			(int32(b[1]) << 16) |
			(int32(b[2]) << 8) |
			int32(b[3]))
		return v, nil
	case int64Code:
		b, err := d.R.ReadN(8)
		if err != nil {
			return 0, err
		}
		v := (int64(b[0]) << 56) |
			(int64(b[1]) << 48) |
			(int64(b[2]) << 40) |
			(int64(b[3]) << 32) |
			(int64(b[4]) << 24) |
			(int64(b[5]) << 16) |
			(int64(b[6]) << 8) |
			int64(b[7])
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

		f := structs.Field(typ, name)
		if f != nil {
			if err := f.DecodeValue(d, v); err != nil {
				return err
			}
		} else {
			_, err := d.DecodeInterface()
			if err != nil {
				return err
			}
		}
	}
	return nil
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
		b, err := d.R.ReadN(2)
		if err != nil {
			return 0, err
		}
		return (uint(b[0]) << 8) | uint(b[1]), nil
	case uint32Code:
		b, err := d.R.ReadN(4)
		if err != nil {
			return 0, err
		}
		v := (uint(b[0]) << 24) |
			(uint(b[1]) << 16) |
			(uint(b[2]) << 8) |
			uint(b[3])
		return v, nil
	case uint64Code:
		b, err := d.R.ReadN(8)
		if err != nil {
			return 0, err
		}
		v := (uint(b[0]) << 56) |
			(uint(b[1]) << 48) |
			(uint(b[2]) << 40) |
			(uint(b[3]) << 32) |
			(uint(b[4]) << 24) |
			(uint(b[5]) << 16) |
			(uint(b[6]) << 8) |
			uint(b[7])
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
		b, err := d.R.ReadN(2)
		if err != nil {
			return 0, err
		}
		return (uint16(b[0]) << 8) | uint16(b[1]), nil
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
		b, err := d.R.ReadN(2)
		if err != nil {
			return 0, err
		}
		return (uint32(b[0]) << 8) | uint32(b[1]), nil
	case uint32Code:
		b, err := d.R.ReadN(4)
		if err != nil {
			return 0, err
		}
		v := (uint32(b[0]) << 24) |
			(uint32(b[1]) << 16) |
			(uint32(b[2]) << 8) |
			uint32(b[3])
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
		b, err := d.R.ReadN(2)
		if err != nil {
			return 0, err
		}
		return int((int16(b[0]) << 8) | int16(b[1])), nil
	case int32Code:
		b, err := d.R.ReadN(4)
		if err != nil {
			return 0, err
		}
		v := int((int32(b[0]) << 24) |
			(int32(b[1]) << 16) |
			(int32(b[2]) << 8) |
			int32(b[3]))
		return v, nil
	case int64Code:
		b, err := d.R.ReadN(8)
		if err != nil {
			return 0, err
		}
		v := int((int64(b[0]) << 56) |
			(int64(b[1]) << 48) |
			(int64(b[2]) << 40) |
			(int64(b[3]) << 32) |
			(int64(b[4]) << 24) |
			(int64(b[5]) << 16) |
			(int64(b[6]) << 8) |
			int64(b[7]))
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
		b, err := d.R.ReadN(2)
		if err != nil {
			return 0, err
		}
		return (int16(b[0]) << 8) | int16(b[1]), nil
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
		b, err := d.R.ReadN(2)
		if err != nil {
			return 0, err
		}
		return int32((int16(b[0]) << 8) | int16(b[1])), nil
	case int32Code:
		b, err := d.R.ReadN(4)
		if err != nil {
			return 0, err
		}
		v := (int32(b[0]) << 24) |
			(int32(b[1]) << 16) |
			(int32(b[2]) << 8) |
			int32(b[3])
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
