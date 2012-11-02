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
	Register(keyPtrType, encodeAppengineKey, decodeAppengineKey)
	Register(cursorType, encodeAppengineCursor, decodeAppengineCursor)
}

func encodeAppengineKey(e *Encoder, v reflect.Value) error {
	key := v.Interface().(*ds.Key)
	return e.Encode(key.Encode())
}

func decodeAppengineKey(d *Decoder, v reflect.Value) error {
	var data string
	if err := d.Decode(&data); err != nil {
		return err
	}
	key, err := ds.DecodeKey(data)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(key))
	return nil
}

func encodeAppengineCursor(e *Encoder, v reflect.Value) error {
	cursor := v.Interface().(ds.Cursor)
	return e.Encode(cursor.String())
}

func decodeAppengineCursor(d *Decoder, v reflect.Value) error {
	var data string
	if err := d.Decode(&data); err != nil {
		return err
	}
	cursor, err := ds.DecodeCursor(data)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(cursor))
	return nil
}
