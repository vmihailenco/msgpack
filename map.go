package msgpack

import (
	"fmt"
	"reflect"

	"gopkg.in/vmihailenco/msgpack.v2/codes"
)

func (e *Encoder) EncodeMapLen(l int) error {
	if l < 16 {
		return e.w.WriteByte(codes.FixedMapLow | byte(l))
	}
	if l < 65536 {
		return e.write2(codes.Map16, uint64(l))
	}
	return e.write4(codes.Map32, uint64(l))
}

func (e *Encoder) encodeMapStringString(m map[string]string) error {
	if err := e.EncodeMapLen(len(m)); err != nil {
		return err
	}
	for mk, mv := range m {
		if err := e.EncodeString(mk); err != nil {
			return err
		}
		if err := e.EncodeString(mv); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) encodeMap(value reflect.Value) error {
	if err := e.EncodeMapLen(value.Len()); err != nil {
		return err
	}
	for _, key := range value.MapKeys() {
		if err := e.EncodeValue(key); err != nil {
			return err
		}
		if err := e.EncodeValue(value.MapIndex(key)); err != nil {
			return err
		}
	}
	return nil
}

func decodeMap(d *Decoder) (interface{}, error) {
	n, err := d.DecodeMapLen()
	if err != nil {
		return nil, err
	}
	if n == -1 {
		return nil, nil
	}

	m := make(map[interface{}]interface{}, n)
	for i := 0; i < n; i++ {
		mk, err := d.DecodeInterface()
		if err != nil {
			return nil, err
		}
		mv, err := d.DecodeInterface()
		if err != nil {
			return nil, err
		}
		m[mk] = mv
	}
	return m, nil
}

func (d *Decoder) DecodeMapLen() (int, error) {
	c, err := d.r.ReadByte()
	if err != nil {
		return 0, err
	}
	return d.mapLen(c)
}

func (d *Decoder) mapLen(c byte) (int, error) {
	if c == codes.Nil {
		return -1, nil
	}
	if c >= codes.FixedMapLow && c <= codes.FixedMapHigh {
		return int(c & codes.FixedMapMask), nil
	}
	if c == codes.Map16 {
		n, err := d.uint16()
		return int(n), err
	}
	if c == codes.Map32 {
		n, err := d.uint32()
		return int(n), err
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding map length", c)
}

func (d *Decoder) decodeIntoMapStringString(mp *map[string]string) error {
	n, err := d.DecodeMapLen()
	if err != nil {
		return err
	}
	if n == -1 {
		return nil
	}

	// TODO(vmihailenco): simpler way?
	m := *mp
	if m == nil {
		*mp = make(map[string]string, n)
		m = *mp
	}

	for i := 0; i < n; i++ {
		mk, err := d.DecodeString()
		if err != nil {
			return err
		}
		mv, err := d.DecodeString()
		if err != nil {
			return err
		}
		m[mk] = mv
	}

	return nil
}

func (d *Decoder) DecodeMap() (interface{}, error) {
	return d.DecodeMapFunc(d)
}

func (d *Decoder) skipMap(c byte) error {
	n, err := d.mapLen(c)
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		if err := d.Skip(); err != nil {
			return err
		}
		if err := d.Skip(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Decoder) mapValue(v reflect.Value) error {
	n, err := d.DecodeMapLen()
	if err != nil {
		return err
	}
	if n == -1 {
		return nil
	}

	typ := v.Type()
	if v.IsNil() {
		v.Set(reflect.MakeMap(typ))
	}
	keyType := typ.Key()
	valueType := typ.Elem()

	for i := 0; i < n; i++ {
		mk := reflect.New(keyType).Elem()
		if err := d.DecodeValue(mk); err != nil {
			return err
		}

		mv := reflect.New(valueType).Elem()
		if err := d.DecodeValue(mv); err != nil {
			return err
		}

		v.SetMapIndex(mk, mv)
	}

	return nil
}

func (e *Encoder) encodeStruct(strct reflect.Value) error {
	structFields := structs.Fields(strct.Type())
	fields := structFields.OmitEmpty(strct)

	if err := e.EncodeMapLen(len(fields)); err != nil {
		return err
	}

	for _, f := range fields {
		if err := e.EncodeString(f.name); err != nil {
			return err
		}
		if err := f.EncodeValue(e, strct); err != nil {
			return err
		}
	}

	return nil
}

func (d *Decoder) structValue(strct reflect.Value) error {
	n, err := d.DecodeMapLen()
	if err != nil {
		return err
	}

	fields := structs.Fields(strct.Type())

	for i := 0; i < n; i++ {
		name, err := d.DecodeString()
		if err != nil {
			return err
		}
		if f := fields.Table[name]; f != nil {
			if err := f.DecodeValue(d, strct); err != nil {
				return err
			}
		} else {
			if err := d.Skip(); err != nil {
				return err
			}
		}
	}

	return nil
}
