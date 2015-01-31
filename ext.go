package msgpack

import (
	"bytes"
	"fmt"
	"reflect"
)

var (
	extTypes [128]reflect.Type
)

func RegisterExt(id int8, value interface{}) {
	extTypes[id] = reflect.TypeOf(value)
}

func extTypeId(typ reflect.Type) int8 {
	for id, t := range extTypes {
		if t == typ {
			return int8(id)
		}
	}
	return -1
}

func makeExtEncoder(id int8, enc encoderFunc) encoderFunc {
	return func(e *Encoder, v reflect.Value) error {
		var buf bytes.Buffer
		oldw := e.w
		e.w = &buf
		err := enc(e, v)
		e.w = oldw
		if err != nil {
			return err
		}

		if err := e.encodeExtLen(buf.Len()); err != nil {
			return err
		}
		if err := e.w.WriteByte(byte(id)); err != nil {
			return err
		}
		return e.write(buf.Bytes())
	}
}

func (e *Encoder) encodeExtLen(l int) error {
	switch {
	case l == 1:
		return e.w.WriteByte(fixExt1Code)
	case l == 2:
		return e.w.WriteByte(fixExt2Code)
	case l == 4:
		return e.w.WriteByte(fixExt4Code)
	case l == 8:
		return e.w.WriteByte(fixExt8Code)
	case l == 16:
		return e.w.WriteByte(fixExt16Code)
	case l < 256:
		return e.write1(ext8Code, uint64(l))
	case l < 65536:
		return e.write2(ext16Code, uint64(l))
	default:
		return e.write4(ext32Code, uint64(l))
	}
}

func (d *Decoder) decodeExtLen() (int, error) {
	c, err := d.r.ReadByte()
	if err != nil {
		return 0, err
	}
	switch c {
	case fixExt1Code:
		return 1, nil
	case fixExt2Code:
		return 2, nil
	case fixExt4Code:
		return 4, nil
	case fixExt8Code:
		return 8, nil
	case fixExt16Code:
		return 16, nil
	case ext8Code:
		n, err := d.uint8()
		return int(n), err
	case ext16Code:
		n, err := d.uint16()
		return int(n), err
	case ext32Code:
		n, err := d.uint32()
		return int(n), err
	default:
		return 0, fmt.Errorf("msgpack: invalid code %x decoding ext length", c)
	}
}

func (d *Decoder) decodeExt() (interface{}, error) {
	_, err := d.decodeExtLen()
	if err != nil {
		return nil, err
	}
	typId, err := d.r.ReadByte()
	if err != nil {
		return nil, err
	}
	typ := extTypes[typId]
	if typ.Kind() == reflect.Invalid {
		return nil, fmt.Errorf("msgpack: unregistered extended type %d", typId)
	}
	v := reflect.New(typ).Elem()
	if err := d.DecodeValue(v); err != nil {
		return nil, err
	}
	return v.Interface(), nil
}
