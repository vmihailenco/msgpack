package msgpack

import (
	"fmt"
	"io"
	"reflect"

	"gopkg.in/vmihailenco/msgpack.v2/codes"
)

func (e *Encoder) encodeBytesLen(l int) error {
	if l < 256 {
		return e.write1(codes.Bin8, uint64(l))
	}
	if l < 65536 {
		return e.write2(codes.Bin16, uint64(l))
	}
	return e.write4(codes.Bin32, uint64(l))
}

func (e *Encoder) encodeStrLen(l int) error {
	if l < 32 {
		return e.w.WriteByte(codes.FixedStrLow | uint8(l))
	}
	if l < 256 {
		return e.write1(codes.Str8, uint64(l))
	}
	if l < 65536 {
		return e.write2(codes.Str16, uint64(l))
	}
	return e.write4(codes.Str32, uint64(l))
}

func (e *Encoder) EncodeString(v string) error {
	if err := e.encodeStrLen(len(v)); err != nil {
		return err
	}
	return e.writeString(v)
}

func (e *Encoder) EncodeBytes(v []byte) error {
	if v == nil {
		return e.EncodeNil()
	}
	if err := e.encodeBytesLen(len(v)); err != nil {
		return err
	}
	return e.write(v)
}

func (e *Encoder) EncodeSliceLen(l int) error {
	if l < 16 {
		return e.w.WriteByte(codes.FixedArrayLow | byte(l))
	}
	if l < 65536 {
		return e.write2(codes.Array16, uint64(l))
	}
	return e.write4(codes.Array32, uint64(l))
}

func (e *Encoder) encodeStringSlice(s []string) error {
	if s == nil {
		return e.EncodeNil()
	}
	if err := e.EncodeSliceLen(len(s)); err != nil {
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
	if v.IsNil() {
		return e.EncodeNil()
	}
	return e.encodeArray(v)
}

func (e *Encoder) encodeArray(v reflect.Value) error {
	l := v.Len()
	if err := e.EncodeSliceLen(l); err != nil {
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
	c, err := d.r.ReadByte()
	if err != nil {
		return 0, err
	}
	return d.bytesLen(c)
}

func (d *Decoder) bytesLen(c byte) (int, error) {
	if c == codes.Nil {
		return -1, nil
	} else if c >= codes.FixedStrLow && c <= codes.FixedStrHigh {
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
	return d.bytes(c)
}

func (d *Decoder) bytes(c byte) ([]byte, error) {
	n, err := d.bytesLen(c)
	if err != nil {
		return nil, err
	}
	if n == -1 {
		return nil, nil
	}
	b := make([]byte, n)
	_, err = io.ReadFull(d.r, b)
	if err != nil {
		return nil, err
	}
	return b, nil
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

func (d *Decoder) bytesValue(value reflect.Value) error {
	v, err := d.DecodeBytes()
	if err != nil {
		return err
	}
	value.SetBytes(v)
	return nil
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

func (d *Decoder) decodeIntoStrings(sp *[]string) error {
	n, err := d.DecodeSliceLen()
	if err != nil {
		return err
	}
	if n == -1 {
		return nil
	}
	s := *sp
	if s == nil || len(s) < n {
		s = make([]string, n)
	}
	for i := 0; i < n; i++ {
		v, err := d.DecodeString()
		if err != nil {
			return err
		}
		s[i] = v
	}
	*sp = s
	return nil
}

func (d *Decoder) DecodeSlice() ([]interface{}, error) {
	n, err := d.DecodeSliceLen()
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

func (d *Decoder) arrayValue(v reflect.Value) error {
	n, err := d.DecodeBytesLen()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		b, err := d.r.ReadByte()
		if err != nil {
			return err
		}
		if i < v.Len() {
			v.Index(i).SetUint(uint64(b))
		}
	}
	return nil
}

func (d *Decoder) sliceValue(v reflect.Value) error {
	n, err := d.DecodeSliceLen()
	if err != nil {
		return err
	}

	if n == -1 {
		v.Set(reflect.Zero(v.Type()))
		return nil
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

func (d *Decoder) strings() ([]string, error) {
	n, err := d.DecodeSliceLen()
	if err != nil {
		return nil, err
	}

	if n == -1 {
		return nil, nil
	}

	ss := make([]string, n)
	for i := 0; i < n; i++ {
		s, err := d.DecodeString()
		if err != nil {
			return nil, err
		}
		ss[i] = s
	}

	return ss, nil
}

func (d *Decoder) stringsValue(value reflect.Value) error {
	ss, err := d.strings()
	if err != nil {
		return err
	}
	value.Set(reflect.ValueOf(ss))
	return nil
}
