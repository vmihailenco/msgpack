package msgpack

import (
	"bufio"
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

type BufferedReader interface {
	Read([]byte) (int, error)
	ReadByte() (byte, error)
	Peek(int) ([]byte, error)
}

type Decoder struct {
	R                  BufferedReader
	b1, b2, b3, b4, b8 []byte
}

func NewDecoder(reader io.Reader) *Decoder {
	b := make([]byte, 8)
	r, ok := reader.(BufferedReader)
	if !ok {
		r = bufio.NewReader(reader)
	}
	return &Decoder{
		R:  r,
		b1: b[:1],
		b2: b[:2],
		b4: b[:4],
		b8: b[:8],
	}
}

func (d *Decoder) Decode(iv interface{}) error {
	switch v := iv.(type) {
	case *string:
		return d.DecodeString(v)
	case *[]byte:
		return d.DecodeBytes(v)
	case *int:
		return d.DecodeInt(v)
	case *int8:
		return d.DecodeInt8(v)
	case *int16:
		return d.DecodeInt16(v)
	case *int32:
		return d.DecodeInt32(v)
	case *int64:
		return d.DecodeInt64(v)
	case *uint:
		return d.DecodeUint(v)
	case *uint8:
		return d.DecodeUint8(v)
	case *uint16:
		return d.DecodeUint16(v)
	case *uint32:
		return d.DecodeUint32(v)
	case *uint64:
		return d.DecodeUint64(v)
	case *bool:
		return d.DecodeBool(v)
	case *float32:
		return d.DecodeFloat32(v)
	case *float64:
		return d.DecodeFloat64(v)
	}

	v := reflect.ValueOf(iv)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	return d.DecodeValue(v.Elem())
}

func (d *Decoder) DecodeValue(v reflect.Value) error {
	b, err := d.R.Peek(1)
	if err != nil {
		return err
	}
	if b[0] == nilCode {
		return nil
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
		return d.arrayValue(v)
	case reflect.Map:
		return d.mapValue(v)
	case reflect.Struct:
		if dec, ok := typDecMap[v.Type()]; ok {
			return dec(d, v)
		}
		return d.structValue(v)
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		if dec, ok := typDecMap[v.Type()]; ok {
			return dec(d, v)
		}
		return d.DecodeValue(v.Elem())
	case reflect.Interface:
		if !v.IsNil() {
			return d.DecodeValue(v.Elem())
		}
	}
	return fmt.Errorf("msgpack: unsupported type %v", v.Type().String())
}

func (d *Decoder) DecodeBool(v *bool) error {
	c, err := d.R.ReadByte()
	if err != nil {
		return err
	}
	switch c {
	case falseCode:
		*v = false
		return nil
	case trueCode:
		*v = true
		return nil
	}
	return fmt.Errorf("msgpack: invalid code %x decoding bool", c)
}

func (d *Decoder) boolValue(value reflect.Value) error {
	var v bool
	if err := d.DecodeBool(&v); err != nil {
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

func (d *Decoder) DecodeUint64(v *uint64) error {
	c, err := d.R.ReadByte()
	if err != nil {
		return err
	}
	if c <= posFixNumHighCode {
		*v = uint64(c)
		return nil
	}
	switch c {
	case uint8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return err
		}
		*v = uint64(c)
		return nil
	case uint16Code:
		if err := d.read(d.b2); err != nil {
			return err
		}
		*v = (uint64(d.b2[0]) << 8) | uint64(d.b2[1])
		return nil
	case uint32Code:
		if err := d.read(d.b4); err != nil {
			return err
		}
		*v = (uint64(d.b4[0]) << 24) |
			(uint64(d.b4[1]) << 16) |
			(uint64(d.b4[2]) << 8) |
			uint64(d.b4[3])
		return nil
	case uint64Code:
		if err := d.read(d.b8); err != nil {
			return err
		}
		*v = (uint64(d.b8[0]) << 56) |
			(uint64(d.b8[1]) << 48) |
			(uint64(d.b8[2]) << 40) |
			(uint64(d.b8[3]) << 32) |
			(uint64(d.b8[4]) << 24) |
			(uint64(d.b8[5]) << 16) |
			(uint64(d.b8[6]) << 8) |
			uint64(d.b8[7])
		return nil
	}
	return fmt.Errorf("msgpack: invalid code %x decoding uint64", c)
}

func (d *Decoder) uint64Value(value reflect.Value) error {
	var v uint64
	if err := d.DecodeUint64(&v); err != nil {
		return err
	}
	value.SetUint(v)
	return nil
}

func (d *Decoder) DecodeInt64(v *int64) error {
	c, err := d.R.ReadByte()
	if err != nil {
		return err
	}
	if c <= posFixNumHighCode || c >= negFixNumLowCode {
		*v = int64(int8(c))
		return nil
	}
	switch c {
	case int8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return err
		}
		*v = int64(int8(c))
		return nil
	case int16Code:
		if err := d.read(d.b2); err != nil {
			return err
		}
		*v = int64((int16(d.b2[0]) << 8) | int16(d.b2[1]))
		return nil
	case int32Code:
		if err := d.read(d.b4); err != nil {
			return err
		}
		*v = int64((int32(d.b4[0]) << 24) |
			(int32(d.b4[1]) << 16) |
			(int32(d.b4[2]) << 8) |
			int32(d.b4[3]))
		return nil
	case int64Code:
		if err := d.read(d.b8); err != nil {
			return err
		}
		*v = (int64(d.b8[0]) << 56) |
			(int64(d.b8[1]) << 48) |
			(int64(d.b8[2]) << 40) |
			(int64(d.b8[3]) << 32) |
			(int64(d.b8[4]) << 24) |
			(int64(d.b8[5]) << 16) |
			(int64(d.b8[6]) << 8) |
			int64(d.b8[7])
		return nil
	}
	return fmt.Errorf("msgpack: invalid code %x decoding int64", c)
}

func (d *Decoder) int64Value(value reflect.Value) error {
	var v int64
	if err := d.DecodeInt64(&v); err != nil {
		return err
	}
	value.SetInt(v)
	return nil
}

func (d *Decoder) DecodeFloat32(v *float32) error {
	c, err := d.R.ReadByte()
	if err != nil {
		return err
	}
	if c != floatCode {
		return fmt.Errorf("msgpack: invalid code %x decoding float32", c)
	}
	b, err := d.uint32()
	if err != nil {
		return err
	}
	*v = math.Float32frombits(b)
	return nil
}

func (d *Decoder) float32Value(value reflect.Value) error {
	var v float32
	if err := d.DecodeFloat32(&v); err != nil {
		return err
	}
	value.SetFloat(float64(v))
	return nil
}

func (d *Decoder) DecodeFloat64(v *float64) error {
	c, err := d.R.ReadByte()
	if err != nil {
		return err
	}
	if c != doubleCode {
		return fmt.Errorf("msgpack: invalid code %x decoding float64", c)
	}
	b, err := d.uint64()
	if err != nil {
		return err
	}
	*v = math.Float64frombits(b)
	return nil
}

func (d *Decoder) float64Value(value reflect.Value) error {
	var v float64
	if err := d.DecodeFloat64(&v); err != nil {
		return err
	}
	value.SetFloat(v)
	return nil
}

func (d *Decoder) DecodeBytes(v *[]byte) error {
	n, err := d.bytesLen()
	if err != nil {
		return err
	}
	if n == -1 {
		return nil
	}
	*v = make([]byte, n)
	if err := d.read(*v); err != nil {
		return err
	}
	return nil
}

func (d *Decoder) bytesLen() (int, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c == nilCode {
		return -1, nil
	} else if c >= fixRawLowCode && c <= fixRawHighCode {
		return int(c & fixRawMask), nil
	}
	switch c {
	case raw16Code:
		n, err := d.uint16()
		return int(n), err
	case raw32Code:
		n, err := d.uint32()
		return int(n), err
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding bytes length", c)
}

func (d *Decoder) bytesValue(value reflect.Value) error {
	var v []byte
	if err := d.DecodeBytes(&v); err != nil {
		return err
	}
	value.SetBytes(v)
	return nil
}

func (d *Decoder) DecodeString(v *string) error {
	var b []byte
	if err := d.DecodeBytes(&b); err != nil {
		return err
	}
	*v = string(b)
	return nil
}

func (d *Decoder) stringValue(value reflect.Value) error {
	var v string
	if err := d.DecodeString(&v); err != nil {
		return err
	}
	value.SetString(v)
	return nil
}

func (d *Decoder) arrayLen() (int, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c >= fixArrayLowCode && c <= fixArrayHighCode {
		return int(c & fixArrayMask), nil
	}
	switch c {
	case array16Code:
		n, err := d.uint16()
		return int(n), err
	case array32Code:
		n, err := d.uint32()
		return int(n), err
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding array length", c)
}

// TODO(vmihailenco): consider using reflect.Append.
// TODO(vmihailenco): specialized path for int*/uint* arrays.
func (d *Decoder) arrayValue(v reflect.Value) error {
	if v.Type().Elem().Kind() == reflect.Uint8 {
		return d.bytesValue(v)
	}

	n, err := d.arrayLen()
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

func (d *Decoder) mapLen() (int, error) {
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
	return 0, fmt.Errorf("msgpack: invalid code %x decoding map length", c)
}

func (d *Decoder) mapValue(v reflect.Value) error {
	n, err := d.mapLen()
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
		var name string
		if err := d.DecodeString(&name); err != nil {
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

func (d *Decoder) DecodeUint(v *uint) error {
	c, err := d.R.ReadByte()
	if err != nil {
		return err
	}
	if c <= posFixNumHighCode {
		*v = uint(c)
		return nil
	}
	switch c {
	case uint8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return err
		}
		*v = uint(c)
		return nil
	case uint16Code:
		if err := d.read(d.b2); err != nil {
			return err
		}
		*v = (uint(d.b2[0]) << 8) | uint(d.b2[1])
		return nil
	case uint32Code:
		if err := d.read(d.b4); err != nil {
			return err
		}
		*v = (uint(d.b4[0]) << 24) |
			(uint(d.b4[1]) << 16) |
			(uint(d.b4[2]) << 8) |
			uint(d.b4[3])
		return nil
	case uint64Code:
		if err := d.read(d.b8); err != nil {
			return err
		}
		*v = (uint(d.b8[0]) << 56) |
			(uint(d.b8[1]) << 48) |
			(uint(d.b8[2]) << 40) |
			(uint(d.b8[3]) << 32) |
			(uint(d.b8[4]) << 24) |
			(uint(d.b8[5]) << 16) |
			(uint(d.b8[6]) << 8) |
			uint(d.b8[7])
		return nil
	}
	return fmt.Errorf("msgpack: invalid code %x decoding uint", c)
}

func (d *Decoder) DecodeUint8(v *uint8) error {
	c, err := d.R.ReadByte()
	if err != nil {
		return err
	}
	if c <= posFixNumHighCode {
		*v = uint8(c)
		return nil
	}
	switch c {
	case uint8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return err
		}
		*v = uint8(c)
		return nil
	}
	return fmt.Errorf("msgpack: invalid code %x decoding uint8", c)
}

func (d *Decoder) DecodeUint16(v *uint16) error {
	c, err := d.R.ReadByte()
	if err != nil {
		return err
	}
	if c <= posFixNumHighCode {
		*v = uint16(c)
		return nil
	}
	switch c {
	case uint8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return err
		}
		*v = uint16(c)
		return nil
	case uint16Code:
		if err := d.read(d.b2); err != nil {
			return err
		}
		*v = (uint16(d.b2[0]) << 8) | uint16(d.b2[1])
		return nil
	}
	return fmt.Errorf("msgpack: invalid code %x decoding uint16", c)
}

func (d *Decoder) DecodeUint32(v *uint32) error {
	c, err := d.R.ReadByte()
	if err != nil {
		return err
	}
	if c <= posFixNumHighCode {
		*v = uint32(c)
		return nil
	}
	switch c {
	case uint8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return err
		}
		*v = uint32(c)
		return nil
	case uint16Code:
		if err := d.read(d.b2); err != nil {
			return err
		}
		*v = (uint32(d.b2[0]) << 8) | uint32(d.b2[1])
		return nil
	case uint32Code:
		if err := d.read(d.b4); err != nil {
			return err
		}
		*v = (uint32(d.b4[0]) << 24) |
			(uint32(d.b4[1]) << 16) |
			(uint32(d.b4[2]) << 8) |
			uint32(d.b4[3])
		return nil
	}
	return fmt.Errorf("msgpack: invalid code %x decoding uint32", c)
}

//------------------------------------------------------------------------------

func (d *Decoder) DecodeInt(v *int) error {
	c, err := d.R.ReadByte()
	if err != nil {
		return err
	}
	if c <= posFixNumHighCode || c >= negFixNumLowCode {
		*v = int(int8(c))
		return nil
	}
	switch c {
	case int8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return err
		}
		*v = int(int8(c))
		return nil
	case int16Code:
		if err := d.read(d.b2); err != nil {
			return err
		}
		*v = int((int16(d.b2[0]) << 8) | int16(d.b2[1]))
		return nil
	case int32Code:
		if err := d.read(d.b4); err != nil {
			return err
		}
		*v = int((int32(d.b4[0]) << 24) |
			(int32(d.b4[1]) << 16) |
			(int32(d.b4[2]) << 8) |
			int32(d.b4[3]))
		return nil
	case int64Code:
		if err := d.read(d.b8); err != nil {
			return err
		}
		*v = int((int64(d.b8[0]) << 56) |
			(int64(d.b8[1]) << 48) |
			(int64(d.b8[2]) << 40) |
			(int64(d.b8[3]) << 32) |
			(int64(d.b8[4]) << 24) |
			(int64(d.b8[5]) << 16) |
			(int64(d.b8[6]) << 8) |
			int64(d.b8[7]))
		return nil
	}
	return fmt.Errorf("msgpack: invalid code %x decoding int64", c)
}

func (d *Decoder) DecodeInt8(v *int8) error {
	c, err := d.R.ReadByte()
	if err != nil {
		return err
	}
	if c <= posFixNumHighCode || c >= negFixNumLowCode {
		*v = int8(c)
		return nil
	}
	switch c {
	case int8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return err
		}
		*v = int8(c)
		return nil
	}
	return fmt.Errorf("msgpack: invalid code %x decoding int8", c)
}

func (d *Decoder) DecodeInt16(v *int16) error {
	c, err := d.R.ReadByte()
	if err != nil {
		return err
	}
	if c <= posFixNumHighCode || c >= negFixNumLowCode {
		*v = int16(int8(c))
		return nil
	}
	switch c {
	case int8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return err
		}
		*v = int16(int8(c))
		return nil
	case int16Code:
		if err := d.read(d.b2); err != nil {
			return err
		}
		*v = (int16(d.b2[0]) << 8) | int16(d.b2[1])
		return nil
	}
	return fmt.Errorf("msgpack: invalid code %x decoding int16", c)
}

func (d *Decoder) DecodeInt32(v *int32) error {
	c, err := d.R.ReadByte()
	if err != nil {
		return err
	}
	if c <= posFixNumHighCode || c >= negFixNumLowCode {
		*v = int32(int8(c))
		return nil
	}
	switch c {
	case int8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return err
		}
		*v = int32(int8(c))
		return nil
	case int16Code:
		if err := d.read(d.b2); err != nil {
			return err
		}
		*v = int32((int16(d.b2[0]) << 8) | int16(d.b2[1]))
		return nil
	case int32Code:
		if err := d.read(d.b4); err != nil {
			return err
		}
		*v = (int32(d.b4[0]) << 24) |
			(int32(d.b4[1]) << 16) |
			(int32(d.b4[2]) << 8) |
			int32(d.b4[3])
		return nil
	}
	return fmt.Errorf("msgpack: invalid code %x decoding int32", c)
}
