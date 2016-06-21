package msgpack

import (
	"fmt"
	"io"
	"reflect"

	"gopkg.in/vmihailenco/msgpack.v2/codes"
)

var sliceStringPtrType = reflect.TypeOf((*[]string)(nil))

func (d *Decoder) DecodeSliceLen() (int, error) {
	c, err := d.r.ReadByte()
	if err != nil {
		return 0, err
	}
	return d.sliceLen(c)
}

func (d *Decoder) sliceLen(c byte) (int, error) {
	if c == codes.Nil {
		return -1, nil
	} else if c >= codes.FixedArrayLow && c <= codes.FixedArrayHigh {
		return int(c & codes.FixedArrayMask), nil
	}
	switch c {
	case codes.Array16:
		n, err := d.uint16()
		return int(n), err
	case codes.Array32:
		n, err := d.uint32()
		return int(n), err
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding array length", c)
}

func decodeByteSliceValue(d *Decoder, v reflect.Value) error {
	c, err := d.r.ReadByte()
	if err != nil {
		return err
	}

	b, err := d.bytes(c, v.Bytes())
	if err != nil {
		return err
	}
	v.SetBytes(b)

	return nil
}

func decodeByteArrayValue(d *Decoder, v reflect.Value) error {
	c, err := d.r.ReadByte()
	if err != nil {
		return err
	}

	n, err := d.bytesLen(c)
	if err != nil {
		return err
	}
	if n == -1 {
		return nil
	}
	if n > v.Cap() {
		n = v.Cap()
	}

	b := v.Slice(0, n).Bytes()
	_, err = io.ReadFull(d.r, b)
	return err
}

func (d *Decoder) DecodeBytesLen() (int, error) {
	c, err := d.r.ReadByte()
	if err != nil {
		return 0, err
	}
	return d.bytesLen(c)
}

func (d *Decoder) bytesLen(c byte) (int, error) {
	if c == codes.Nil {
		return -1, nil
	} else if codes.IsFixedString(c) {
		return int(c & codes.FixedStrMask), nil
	}
	switch c {
	case codes.Str8, codes.Bin8:
		n, err := d.uint8()
		return int(n), err
	case codes.Str16, codes.Bin16:
		n, err := d.uint16()
		return int(n), err
	case codes.Str32, codes.Bin32:
		n, err := d.uint32()
		return int(n), err
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding bytes length", c)
}

func (d *Decoder) DecodeBytes() ([]byte, error) {
	c, err := d.r.ReadByte()
	if err != nil {
		return nil, err
	}
	return d.bytes(c, nil)
}

func (d *Decoder) bytes(c byte, b []byte) ([]byte, error) {
	n, err := d.bytesLen(c)
	if err != nil {
		return nil, err
	}
	if n == -1 {
		return nil, nil
	}

	if b == nil {
		b = make([]byte, n)
	} else if len(b) != n {
		b = setBytesLen(b, n)
	}

	_, err = io.ReadFull(d.r, b)
	return b, err
}

func (d *Decoder) decodeBytesPtr(ptr *[]byte) error {
	c, err := d.r.ReadByte()
	if err != nil {
		return err
	}
	return d.bytesPtr(c, ptr)
}

func (d *Decoder) bytesPtr(c byte, ptr *[]byte) error {
	n, err := d.bytesLen(c)
	if err != nil {
		return err
	}
	if n == -1 {
		*ptr = nil
		return nil
	}

	b := *ptr
	if b == nil {
		*ptr = make([]byte, n)
		b = *ptr
	} else if len(b) != n {
		*ptr = setBytesLen(b, n)
		b = *ptr
	}

	_, err = io.ReadFull(d.r, b)
	return err
}

func (d *Decoder) skipBytes(c byte) error {
	n, err := d.bytesLen(c)
	if err != nil {
		return err
	}
	if n == -1 {
		return nil
	}
	return d.skipN(n)
}

func (d *Decoder) DecodeString() (string, error) {
	c, err := d.r.ReadByte()
	if err != nil {
		return "", err
	}
	return d.string(c)
}

func (d *Decoder) string(c byte) (string, error) {
	n, err := d.bytesLen(c)
	if err != nil {
		return "", err
	}
	if n == -1 {
		return "", nil
	}
	b, err := d.readN(n)
	return string(b), err
}

func decodeStringValue(d *Decoder, v reflect.Value) error {
	s, err := d.DecodeString()
	if err != nil {
		return err
	}
	v.SetString(s)
	return nil
}

func decodeStringSliceValue(d *Decoder, v reflect.Value) error {
	ptr := v.Addr().Convert(sliceStringPtrType).Interface().(*[]string)
	return d.decodeStringSlicePtr(ptr)
}

func (d *Decoder) decodeStringSlicePtr(ptr *[]string) error {
	n, err := d.DecodeSliceLen()
	if err != nil {
		return err
	}
	if n == -1 {
		return nil
	}

	s := *ptr
	if s == nil {
		*ptr = make([]string, n)
		s = *ptr
	} else if len(s) != n {
		*ptr = setStringsLen(s, n)
		s = *ptr
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

func decodeSliceValue(d *Decoder, v reflect.Value) error {
	n, err := d.DecodeSliceLen()
	if err != nil {
		return err
	}

	if n == -1 {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}

	if v.IsNil() {
		v.Set(reflect.MakeSlice(v.Type(), n, n))
	} else if v.Len() != n {
		v.Set(setSliceValueLen(v, n))
	}

	for i := 0; i < n; i++ {
		sv := v.Index(i)
		if err := d.DecodeValue(sv); err != nil {
			return err
		}
	}

	return nil
}

func decodeArrayValue(d *Decoder, v reflect.Value) error {
	n, err := d.DecodeSliceLen()
	if err != nil {
		return err
	}

	if n == -1 {
		return nil
	}

	for i := 0; i < n && i < v.Len(); i++ {
		sv := v.Index(i)
		if err := d.DecodeValue(sv); err != nil {
			return err
		}
	}

	return nil
}

func (d *Decoder) DecodeSlice() ([]interface{}, error) {
	c, err := d.r.ReadByte()
	if err != nil {
		return nil, err
	}
	return d.decodeSlice(c)
}

func (d *Decoder) decodeSlice(c byte) ([]interface{}, error) {
	n, err := d.sliceLen(c)
	if err != nil {
		return nil, err
	}
	if n == -1 {
		return nil, nil
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

func (d *Decoder) skipSlice(c byte) error {
	n, err := d.sliceLen(c)
	if err != nil {
		return err
	}

	for i := 0; i < n; i++ {
		if err := d.Skip(); err != nil {
			return err
		}
	}

	return nil
}
