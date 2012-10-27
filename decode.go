package msgpack

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"reflect"
)

func Unmarshal(data []byte, v interface{}) error {
	buf := bytes.NewBuffer(data)
	return NewDecoder(buf).Decode(v)
}

// An InvalidUnmarshalError describes an invalid argument passed to Unmarshal.
// (The argument to Unmarshal must be a non-nil pointer.)
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "msgpack: Decode(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "msgpack: Decode(non-pointer " + e.Type.String() + ")"
	}
	return "msgpack: Decode(nil " + e.Type.String() + ")"
}

type InvalidCodeError struct {
	Value reflect.Value
	Code  byte
}

func (e *InvalidCodeError) Error() string {
	return fmt.Sprintf(
		"msgpack: invalid code %x decoding %v",
		e.Code,
		e.Value.Kind().String(),
	)
}

type Decoder struct {
	R                  io.Reader
	b1, b2, b3, b4, b8 []byte
}

func NewDecoder(reader io.Reader) *Decoder {
	b := make([]byte, 8)
	return &Decoder{
		R:  reader,
		b1: b[:1],
		b2: b[:2],
		b4: b[:4],
		b8: b[:8],
	}
}

func (d *Decoder) Decode(v interface{}) error {
	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	value = value.Elem()
	return d.DecodeValue(value)
}

func (d *Decoder) DecodeValue(v reflect.Value) error {
	c, err := d.readByte()
	if err != nil {
		return err
	}
	return d.DecodeValueByte(v, c)
}

func (d *Decoder) DecodeValueByte(v reflect.Value, c byte) error {
	// TODO(vmihailenco): is this correct?
	if c == nilCode {
		return nil
	}

	switch v.Kind() {
	case reflect.Bool:
		return d.DecodeBoolValue(v, c)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return d.DecodeUint64Value(v, c)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return d.DecodeInt64Value(v, c)
	case reflect.Float32:
		return d.DecodeFloat32Value(v, c)
	case reflect.Float64:
		return d.DecodeFloat64Value(v, c)
	case reflect.String:
		return d.DecodeStringValue(v, c)
	case reflect.Array, reflect.Slice:
		return d.DecodeArrayValue(v, c)
	case reflect.Map:
		return d.DecodeMapValue(v, c)
	case reflect.Struct:
		if dec, ok := typDecMap[v.Type()]; ok {
			return dec(d, v, c)
		}
		return d.DecodeStructValue(v, c)
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		if dec, ok := typDecMap[v.Type()]; ok {
			return dec(d, v, c)
		}
		return d.DecodeValueByte(v.Elem(), c)
	case reflect.Interface:
		if !v.IsNil() {
			return d.DecodeValueByte(v.Elem(), c)
		}
	default:
		return fmt.Errorf("msgpack: unsupported type %v", v.Type().String())
	}
	panic("not reached")
}

func (d *Decoder) DecodeBoolValue(v reflect.Value, c byte) error {
	switch c {
	case falseCode:
		v.SetBool(false)
	case trueCode:
		v.SetBool(true)
	default:
		return &InvalidCodeError{v, c}
	}
	return nil
}

func (d *Decoder) DecodeUint64Byte(c byte) (uint64, error) {
	if c <= posFixNumHighCode {
		return uint64(c), nil
	}
	switch c {
	case uint8Code:
		c, err := d.readByte()
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
		n := (uint64(d.b4[0]) << 24) |
			(uint64(d.b4[1]) << 16) |
			(uint64(d.b4[2]) << 8) |
			uint64(d.b4[3])
		return n, nil
	case uint64Code:
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
	default:
		return 0, fmt.Errorf("msgpack: invalid code %x decoding uint", c)
	}
	panic("not reached")
}

func (d *Decoder) DecodeUint16() (uint64, error) {
	return d.DecodeUint64Byte(uint16Code)
}

func (d *Decoder) DecodeUint32() (uint64, error) {
	return d.DecodeUint64Byte(uint32Code)
}

func (d *Decoder) DecodeUint64() (uint64, error) {
	return d.DecodeUint64Byte(uint64Code)
}

func (d *Decoder) DecodeUint64Value(v reflect.Value, c byte) error {
	n, err := d.DecodeUint64Byte(c)
	if err != nil {
		return err
	}
	v.SetUint(n)
	return nil
}

func (d *Decoder) DecodeInt64(c byte) (int64, error) {
	if c <= posFixNumHighCode || c >= negFixNumLowCode {
		return int64(int8(c)), nil
	}

	switch c {
	case int8Code:
		c, err := d.readByte()
		if err != nil {
			return 0, err
		}
		return int64(int8(c)), nil
	case int16Code:
		if err := d.read(d.b2); err != nil {
			return 0, err
		}
		n := (int16(d.b2[0]) << 8) | int16(d.b2[1])
		return int64(n), nil
	case int32Code:
		if err := d.read(d.b4); err != nil {
			return 0, err
		}
		n := (int32(d.b4[0]) << 24) |
			(int32(d.b4[1]) << 16) |
			(int32(d.b4[2]) << 8) |
			int32(d.b4[3])
		return int64(n), nil
	case int64Code:
		if err := d.read(d.b8); err != nil {
			return 0, err
		}
		n := (int64(d.b8[0]) << 56) |
			(int64(d.b8[1]) << 48) |
			(int64(d.b8[2]) << 40) |
			(int64(d.b8[3]) << 32) |
			(int64(d.b8[4]) << 24) |
			(int64(d.b8[5]) << 16) |
			(int64(d.b8[6]) << 8) |
			int64(d.b8[7])
		return n, nil
	default:
		return 0, fmt.Errorf("msgpack: invalid code %x decoding int", c)
	}
	panic("not reached")
}

func (d *Decoder) DecodeInt64Value(v reflect.Value, c byte) error {
	n, err := d.DecodeInt64(c)
	if err != nil {
		return err
	}
	v.SetInt(n)
	return nil
}

func (d *Decoder) DecodeFloat32Value(v reflect.Value, c byte) error {
	if c != floatCode {
		return fmt.Errorf("msgpack: invalid code %x decoding float32", c)
	}
	b, err := d.DecodeUint32()
	if err != nil {
		return err
	}
	v.SetFloat(float64(math.Float32frombits(uint32(b))))
	return nil
}

func (d *Decoder) DecodeFloat64Value(v reflect.Value, c byte) error {
	if c != doubleCode {
		return fmt.Errorf("msgpack: invalid code %x decoding float64", c)
	}
	b, err := d.DecodeUint64()
	if err != nil {
		return err
	}
	v.SetFloat(math.Float64frombits(b))
	return nil
}

func (d *Decoder) DecodeBytesLen(c byte) (int, error) {
	if c >= fixRawLowCode && c <= fixRawHighCode {
		return int(c & fixRawMask), nil
	}
	switch c {
	case raw16Code:
		n, err := d.DecodeUint16()
		return int(n), err
	case raw32Code:
		n, err := d.DecodeUint32()
		return int(n), err
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding bytes length", c)
}

func (d *Decoder) DecodeBytes(c byte) ([]byte, error) {
	n, err := d.DecodeBytesLen(c)
	if err != nil {
		return nil, err
	}
	data := make([]byte, n)
	if err := d.read(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (d *Decoder) DecodeBytesValue(v reflect.Value, c byte) error {
	data, err := d.DecodeBytes(c)
	if err != nil {
		return err
	}
	v.SetBytes(data)
	return nil
}

func (d *Decoder) DecodeStringValue(v reflect.Value, c byte) error {
	data, err := d.DecodeBytes(c)
	if err != nil {
		return err
	}
	v.SetString(string(data))
	return nil
}

func (d *Decoder) decodeArrayLen(c byte) (int, error) {
	if c >= fixArrayLowCode && c <= fixArrayHighCode {
		return int(c & fixArrayMask), nil
	}
	switch c {
	case array16Code:
		n, err := d.DecodeUint16()
		return int(n), err
	case array32Code:
		n, err := d.DecodeUint32()
		return int(n), err
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding array length", c)
}

// TODO(vmihailenco): consider using reflect.Append.
// TODO(vmihailenco): specialized path for int*/uint* arrays.
func (d *Decoder) DecodeArrayValue(v reflect.Value, c byte) error {
	if v.Type().Elem().Kind() == reflect.Uint8 {
		return d.DecodeBytesValue(v, c)
	}

	n, err := d.decodeArrayLen(c)
	if err != nil {
		return err
	}

	if v.IsNil() || v.Len() < n {
		v.Set(reflect.MakeSlice(v.Type(), n, n))
	}

	elemType := v.Type().Elem()
	for i := 0; i < n; i++ {
		elem := reflect.New(elemType).Elem()
		if err := d.DecodeValue(elem); err != nil {
			return err
		}
		v.Index(i).Set(elem)
	}

	return nil
}

func (d *Decoder) decodeMapLen(c byte) (int, error) {
	if c >= fixMapLowCode && c <= fixMapHighCode {
		return int(c & fixMapMask), nil
	}
	switch c {
	case map16Code:
		n, err := d.DecodeUint16()
		return int(n), err
	case map32Code:
		n, err := d.DecodeUint32()
		return int(n), err
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding map length", c)
}

func (d *Decoder) DecodeMapValue(v reflect.Value, c byte) error {
	n, err := d.decodeMapLen(c)
	if err != nil {
		return err
	}

	typ := v.Type()
	if v.IsNil() {
		v.Set(reflect.MakeMap(typ))
	}
	keyType := typ.Key()
	valueType := typ.Elem()

	for i := 0; i < n; i++ {
		key := reflect.New(keyType).Elem()
		d.DecodeValue(key)

		value := reflect.New(valueType).Elem()
		d.DecodeValue(value)

		v.SetMapIndex(key, value)
	}

	return nil
}

func (d *Decoder) decodeStructLen(c byte) (int, error) {
	if c >= fixMapLowCode && c <= fixMapHighCode {
		return int(c & fixMapMask), nil
	}
	switch c {
	case map16Code:
		n, err := d.DecodeUint16()
		return int(n), err
	case map32Code:
		n, err := d.DecodeUint32()
		return int(n), err
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding struct length", c)
}

func (d *Decoder) DecodeStructValue(v reflect.Value, c byte) error {
	n, err := d.decodeStructLen(c)
	if err != nil {
		return err
	}

	typ := v.Type()
	for i := 0; i < n; i++ {
		c, err := d.readByte()
		if err != nil {
			return err
		}

		data, err := d.DecodeBytes(c)
		if err != nil {
			return err
		}
		name := string(data)

		f := reflectCache.Field(typ, name)
		if f == nil {
			continue
		}

		if err := d.DecodeValue(v.Field(f.Ind)); err != nil {
			return fmt.Errorf("msgpack: can't decode value for field %q: %v", name, err)
		}
	}
	return nil
}

func (d *Decoder) read(b []byte) error {
	n, err := d.R.Read(b)
	if n == len(b) {
		return nil
	}
	return err
}

func (d *Decoder) readByte() (byte, error) {
	err := d.read(d.b1)
	return d.b1[0], err
}
