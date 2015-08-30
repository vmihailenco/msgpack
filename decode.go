package msgpack

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"

	"gopkg.in/vmihailenco/msgpack.v2/codes"
)

type bufReader interface {
	Read([]byte) (int, error)
	ReadByte() (byte, error)
	UnreadByte() error
}

func Unmarshal(b []byte, v ...interface{}) error {
	if len(v) == 1 && v[0] != nil {
		unmarshaler, ok := v[0].(Unmarshaler)
		if ok {
			return unmarshaler.UnmarshalMsgpack(b)
		}
	}
	return NewDecoder(bytes.NewReader(b)).Decode(v...)
}

type Decoder struct {
	DecodeMapFunc func(*Decoder) (interface{}, error)

	r   bufReader
	buf []byte
}

func NewDecoder(r io.Reader) *Decoder {
	br, ok := r.(bufReader)
	if !ok {
		br = bufio.NewReader(r)
	}
	return &Decoder{
		DecodeMapFunc: decodeMap,

		r:   br,
		buf: make([]byte, 64),
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
	decode := getDecoder(v.Type())
	return decode(d, v)
}

func (d *Decoder) DecodeNil() error {
	c, err := d.r.ReadByte()
	if err != nil {
		return err
	}
	if c != codes.Nil {
		return fmt.Errorf("msgpack: invalid code %x decoding nil", c)
	}
	return nil
}

func (d *Decoder) DecodeBool() (bool, error) {
	c, err := d.r.ReadByte()
	if err != nil {
		return false, err
	}
	switch c {
	case codes.False:
		return false, nil
	case codes.True:
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

func (d *Decoder) interfaceValue(v reflect.Value) error {
	iface, err := d.DecodeInterface()
	if err != nil {
		return err
	}
	if iface != nil {
		v.Set(reflect.ValueOf(iface))
	}
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
	c, err := d.peekCode()
	if err != nil {
		return nil, err
	}

	if codes.IsFixedNum(c) {
		return d.DecodeInt64()
	} else if codes.IsFixedMap(c) {
		return d.DecodeMap()
	} else if codes.IsFixedArray(c) {
		return d.DecodeSlice()
	} else if codes.IsFixedString(c) {
		return d.DecodeString()
	}

	switch c {
	case codes.Nil:
		_, err := d.r.ReadByte()
		return nil, err
	case codes.False, codes.True:
		return d.DecodeBool()
	case codes.Float:
		return d.DecodeFloat32()
	case codes.Double:
		return d.DecodeFloat64()
	case codes.Uint8, codes.Uint16, codes.Uint32, codes.Uint64:
		return d.DecodeUint64()
	case codes.Int8, codes.Int16, codes.Int32, codes.Int64:
		return d.DecodeInt64()
	case codes.Bin8, codes.Bin16, codes.Bin32:
		return d.DecodeBytes()
	case codes.Str8, codes.Str16, codes.Str32:
		return d.DecodeString()
	case codes.Array16, codes.Array32:
		return d.DecodeSlice()
	case codes.Map16, codes.Map32:
		return d.DecodeMap()
	case codes.FixExt1, codes.FixExt2, codes.FixExt4, codes.FixExt8, codes.FixExt16, codes.Ext8, codes.Ext16, codes.Ext32:
		return d.decodeExt()
	}

	return 0, fmt.Errorf("msgpack: invalid code %x decoding interface{}", c)
}

// PeekCode returns the next Msgpack code. See
// https://github.com/msgpack/msgpack/blob/master/spec.md#formats for details.
func (d *Decoder) peekCode() (code byte, err error) {
	code, err = d.r.ReadByte()
	if err != nil {
		return 0, err
	}
	return code, d.r.UnreadByte()
}

func (d *Decoder) hasNilCode() bool {
	code, err := d.peekCode()
	return err == nil && code == codes.Nil
}

func (d *Decoder) readN(n int) ([]byte, error) {
	var b []byte
	if n <= cap(d.buf) {
		b = d.buf[:n]
	} else {
		b = make([]byte, n)
	}
	_, err := io.ReadFull(d.r, b)
	return b, err
}
