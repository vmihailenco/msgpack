package msgpack

import (
	"fmt"
	"reflect"
)

func (e *Encoder) EncodeString(v string) error {
	return e.EncodeBytes([]byte(v))
}

func (e *Encoder) EncodeBytes(v []byte) error {
	switch l := len(v); {
	case l < 32:
		if err := e.W.WriteByte(fixRawLowCode | uint8(l)); err != nil {
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

func (e *Encoder) encodeSliceLen(l int) error {
	switch {
	case l < 16:
		if err := e.W.WriteByte(fixArrayLowCode | byte(l)); err != nil {
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
	return nil
}

func (e *Encoder) encodeStringSlice(s []string) error {
	if err := e.encodeSliceLen(len(s)); err != nil {
		return err
	}
	for _, v := range s {
		if err := e.EncodeString(v); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) encodeSlice(v reflect.Value) error {
	switch v.Type().Elem().Kind() {
	case reflect.Uint8:
		return e.EncodeBytes(v.Bytes())
	case reflect.String:
		return e.encodeStringSlice(v.Interface().([]string))
	}

	l := v.Len()
	if err := e.encodeSliceLen(l); err != nil {
		return err
	}
	for i := 0; i < l; i++ {
		if err := e.EncodeValue(v.Index(i)); err != nil {
			return err
		}
	}

	return nil
}

func (d *Decoder) DecodeBytesLen() (int, error) {
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

func (d *Decoder) DecodeBytes() ([]byte, error) {
	n, err := d.DecodeBytesLen()
	if err != nil {
		return nil, err
	}
	if n == -1 {
		return nil, nil
	}
	v := make([]byte, n)
	if err := d.read(v); err != nil {
		return nil, err
	}
	return v, nil
}

func (d *Decoder) bytesValue(value reflect.Value) error {
	v, err := d.DecodeBytes()
	if err != nil {
		return err
	}
	value.SetBytes(v)
	return nil
}

func (d *Decoder) DecodeString() (string, error) {
	b, err := d.DecodeBytes()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (d *Decoder) stringValue(value reflect.Value) error {
	v, err := d.DecodeString()
	if err != nil {
		return err
	}
	value.SetString(v)
	return nil
}

func (d *Decoder) DecodeSliceLen() (int, error) {
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

func (d *Decoder) decodeIntoStrings(sp *[]string) error {
	n, err := d.DecodeSliceLen()
	if err != nil {
		return err
	}
	s := *sp
	if len(s) < n {
		s = make([]string, n)
	}
	for i := 0; i < n; i++ {
		v, err := d.DecodeString()
		if err != nil {
			return err
		}
		s[i] = v
	}
	return nil
}

func (d *Decoder) DecodeSlice() ([]interface{}, error) {
	n, err := d.DecodeSliceLen()
	if err != nil {
		return nil, err
	}
	s := make([]interface{}, n)
	for i := 0; i < n; i++ {
		v, err := d.DecodeInterface()
		if err != nil {
			return nil, err
		}
		s[i] = v
	}
	return s, nil
}

func (d *Decoder) sliceValue(v reflect.Value) error {
	elemType := v.Type().Elem()
	switch elemType.Kind() {
	case reflect.Uint8:
		b, err := d.DecodeBytes()
		if err != nil {
			return err
		}
		v.SetBytes(b)
		return nil
	}

	n, err := d.DecodeSliceLen()
	if err != nil {
		return err
	}

	if v.Len() < n || (v.Kind() == reflect.Slice && v.IsNil()) {
		v.Set(reflect.MakeSlice(v.Type(), n, n))
	}

	for i := 0; i < n; i++ {
		sv := v.Index(i)
		if err := d.DecodeValue(sv); err != nil {
			return err
		}
	}

	return nil
}
