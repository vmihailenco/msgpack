package msgpack

import (
	"fmt"
	"reflect"

	"gopkg.in/vmihailenco/msgpack.v2/codes"
)

var sliceStringPtrType = reflect.TypeOf((*[]string)(nil))

// Deprecated. Use DecodeArrayLen instead.
func (d *Decoder) DecodeSliceLen() (int, error) {
	return d.DecodeArrayLen()
}

func (d *Decoder) DecodeArrayLen() (int, error) {
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
