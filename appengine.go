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
	Register(keyPtrType, encodeDatastoreKeyValue, decodeDatastoreKeyValue)
	Register(cursorType, encodeDatastoreCursor, decodeDatastoreCursor)
}

func EncodeDatastoreKey(e *Encoder, key *ds.Key) error {
	return e.EncodeString(key.Encode())
}

func encodeDatastoreKeyValue(e *Encoder, v reflect.Value) error {
	key := v.Interface().(*ds.Key)
	return EncodeDatastoreKey(e, key)
}

func DecodeDatastoreKey(d *Decoder) (*ds.Key, error) {
	v, err := d.DecodeString()
	if err != nil {
		return nil, err
	}
	return ds.DecodeKey(v)
}

func decodeDatastoreKeyValue(d *Decoder, v reflect.Value) error {
	key, err := DecodeDatastoreKey(d)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(key))
	return nil
}

func encodeDatastoreCursor(e *Encoder, v reflect.Value) error {
	cursor := v.Interface().(ds.Cursor)
	return e.Encode(cursor.String())
}

func decodeDatastoreCursor(d *Decoder, v reflect.Value) error {
	s, err := d.DecodeString()
	if err != nil {
		return err
	}
	cursor, err := ds.DecodeCursor(s)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(cursor))
	return nil
}
