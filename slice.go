package msgpack

import (
	"fmt"
	"reflect"
)

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

func (d *Decoder) DecodeBytes() ([]byte, error) {
	n, err := d.bytesLen()
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

func (d *Decoder) sliceLen() (int, error) {
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
	n, err := d.sliceLen()
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
	n, err := d.sliceLen()
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
	switch vv := v.Interface().(type) {
	case *[]byte:
		b, err := d.DecodeBytes()
		if err != nil {
			return err
		}
		*vv = b
		return nil
	case *[]string:
		return d.decodeIntoStrings(vv)
	}

	n, err := d.sliceLen()
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
