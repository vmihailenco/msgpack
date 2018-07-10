package msgpack

import (
	"reflect"
	"sort"

	"github.com/vmihailenco/msgpack/codes"
)

func encodeMapValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}

	if err := e.EncodeMapLen(v.Len()); err != nil {
		return err
	}

	for _, key := range v.MapKeys() {
		if err := e.EncodeValue(key); err != nil {
			return err
		}
		if err := e.EncodeValue(v.MapIndex(key)); err != nil {
			return err
		}
	}

	return nil
}

func encodeMapStringStringValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}

	if err := e.EncodeMapLen(v.Len()); err != nil {
		return err
	}

	m := v.Convert(mapStringStringType).Interface().(map[string]string)
	if e.sortMapKeys {
		return e.encodeSortedMapStringString(m)
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

func encodeMapStringInterfaceValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}

	if err := e.EncodeMapLen(v.Len()); err != nil {
		return err
	}

	m := v.Convert(mapStringInterfaceType).Interface().(map[string]interface{})
	if e.sortMapKeys {
		return e.encodeSortedMapStringInterface(m)
	}

	for mk, mv := range m {
		if err := e.EncodeString(mk); err != nil {
			return err
		}
		if err := e.Encode(mv); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) encodeSortedMapStringString(m map[string]string) error {
	keys := make([]string, 0, len(m))
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		err := e.EncodeString(k)
		if err != nil {
			return err
		}
		if err = e.EncodeString(m[k]); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) encodeSortedMapStringInterface(m map[string]interface{}) error {
	keys := make([]string, 0, len(m))
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		err := e.EncodeString(k)
		if err != nil {
			return err
		}
		if err = e.Encode(m[k]); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) EncodeMapLen(l int) error {
	if l < 16 {
		return e.writeCode(codes.FixedMapLow | codes.Code(l))
	}
	if l < 65536 {
		return e.write2(codes.Map16, uint64(l))
	}
	return e.write4(codes.Map32, uint32(l))
}

func encodeStructValue(e *Encoder, strct reflect.Value) error {
	var structFields *fields
	if e.useJSONTag {
		structFields = jsonStructs.Fields(strct.Type())
	} else {
		structFields = structs.Fields(strct.Type())
	}

	if e.structAsArray || structFields.AsArray {
		if e.sortStructFields {
			return encodeSortedStructValue(e, strct, structFields, true)
		}
		return encodeStructValueAsArray(e, strct, structFields.List)
	}
	fields := structFields.OmitEmpty(strct)
	if e.sortStructFields {
		return encodeSortedStructValue(e, strct, fields, false)
	}

	if err := e.EncodeMapLen(len(fields.List)); err != nil {
		return err
	}

	for _, f := range fields.List {
		if err := e.EncodeString(f.name); err != nil {
			return err
		}
		if err := f.EncodeValue(e, strct); err != nil {
			return err
		}
	}

	return nil
}

func encodeSortedStructValue(e *Encoder, strct reflect.Value, fields *fields, asArray bool) error {
	var err error
	if asArray {
		err = e.EncodeArrayLen(len(fields.Table))
	} else {
		err = e.EncodeMapLen(len(fields.Table))
	}
	if err != nil {
		return err
	}
	fns := make([]string, 0, len(fields.Table))
	for fn := range fields.Table {
		fns = append(fns, fn)
	}
	sort.Strings(fns)
	for _, fn := range fns {
		if !asArray {
			if err = e.EncodeString(fn); err != nil {
				return err
			}
		}
		if err = fields.Table[fn].EncodeValue(e, strct); err != nil {
			return err
		}
	}
	return nil
}

func encodeStructValueAsArray(e *Encoder, strct reflect.Value, fields []*field) error {
	if err := e.EncodeArrayLen(len(fields)); err != nil {
		return err
	}
	for _, f := range fields {
		if err := f.EncodeValue(e, strct); err != nil {
			return err
		}
	}
	return nil
}
