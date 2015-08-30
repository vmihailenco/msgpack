package msgpack

import (
	"bytes"
	"fmt"
	"reflect"
	"sync"

	"gopkg.in/vmihailenco/msgpack.v2/codes"
)

var (
	extTypes [128]reflect.Type
)

var bufferPool = &sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

func RegisterExt(id int8, value interface{}) {
	if extTypes[id] != nil {
		panic(fmt.Errorf("ext with id %d is already registered", id))
	}
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
		buf := bufferPool.Get().(*bytes.Buffer)
		defer bufferPool.Put(buf)
		buf.Reset()

		oldw := e.w
		e.w = buf
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
		return e.w.WriteByte(codes.FixExt1)
	case l == 2:
		return e.w.WriteByte(codes.FixExt2)
	case l == 4:
		return e.w.WriteByte(codes.FixExt4)
	case l == 8:
		return e.w.WriteByte(codes.FixExt8)
	case l == 16:
		return e.w.WriteByte(codes.FixExt16)
	case l < 256:
		return e.write1(codes.Ext8, uint64(l))
	case l < 65536:
		return e.write2(codes.Ext16, uint64(l))
	default:
		return e.write4(codes.Ext32, uint64(l))
	}
}

func (d *Decoder) decodeExtLen() (int, error) {
	c, err := d.r.ReadByte()
	if err != nil {
		return 0, err
	}
	switch c {
	case codes.FixExt1:
		return 1, nil
	case codes.FixExt2:
		return 2, nil
	case codes.FixExt4:
		return 4, nil
	case codes.FixExt8:
		return 8, nil
	case codes.FixExt16:
		return 16, nil
	case codes.Ext8:
		n, err := d.uint8()
		return int(n), err
	case codes.Ext16:
		n, err := d.uint16()
		return int(n), err
	case codes.Ext32:
		n, err := d.uint32()
		return int(n), err
	default:
		return 0, fmt.Errorf("msgpack: invalid code %x decoding ext length", c)
	}
}

func (d *Decoder) decodeExt() (interface{}, error) {
	// TODO: use decoded length.
	_, err := d.decodeExtLen()
	if err != nil {
		return nil, err
	}
	extId, err := d.r.ReadByte()
	if err != nil {
		return nil, err
	}
	typ := extTypes[extId]
	if typ == nil {
		return nil, fmt.Errorf("msgpack: unregistered ext id %d", extId)
	}
	v := reflect.New(typ).Elem()
	if err := d.DecodeValue(v); err != nil {
		return nil, err
	}
	return v.Interface(), nil
}
