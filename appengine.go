// +build appengine

package msgpack

import (
	"reflect"

	ds "appengine/datastore"
)

var (
	keyPtrType = reflect.TypeOf((*ds.Key)(nil))
	cursorType = reflect.TypeOf((*ds.Cursor)(nil)).Elem()
)

func init() {
	Register(keyPtrType, EncodeAppengineKey, DecodeAppengineKey)
	Register(cursorType, EncodeAppengineCursor, DecodeAppengineCursor)
}

func EncodeAppengineKey(e *Encoder, v reflect.Value) error {
	key := v.Interface().(*ds.Key)
	return e.EncodeBytes([]byte(key.Encode()))
}

func DecodeAppengineKey(d *Decoder, v reflect.Value, c byte) error {
	data, err := d.DecodeBytes(c)
	if err != nil {
		return err
	}
	key, err := ds.DecodeKey(string(data))
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(key))
	return nil
}

func EncodeAppengineCursor(e *Encoder, v reflect.Value) error {
	cursor := v.Interface().(ds.Cursor)
	return e.EncodeBytes([]byte(cursor.String()))
}

func DecodeAppengineCursor(d *Decoder, v reflect.Value, c byte) error {
	data, err := d.DecodeBytes(c)
	if err != nil {
		return err
	}
	cursor, err := ds.DecodeCursor(string(data))
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(cursor))
	return nil
}
