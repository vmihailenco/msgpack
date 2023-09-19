package msgpappengine

import (
	"reflect"

	"github.com/vmihailenco/msgpack/v6"
	ds "google.golang.org/appengine/datastore"
)

func init() {
	msgpack.Register((*ds.Key)(nil), encodeDatastoreKeyValue, decodeDatastoreKeyValue)
	msgpack.Register((*ds.Cursor)(nil), encodeDatastoreCursorValue, decodeDatastoreCursorValue)
}

func EncodeDatastoreKey(e *msgpack.Encoder, key *ds.Key) error {
	if key == nil {
		return e.EncodeNil()
	}
	return e.EncodeString(key.Encode())
}

func encodeDatastoreKeyValue(e *msgpack.Encoder, v reflect.Value) error {
	key := v.Interface().(*ds.Key)
	return EncodeDatastoreKey(e, key)
}

func DecodeDatastoreKey(d *msgpack.Decoder) (*ds.Key, error) {
	v, err := d.DecodeString()
	if err != nil {
		return nil, err
	}
	if v == "" {
		return nil, nil
	}
	return ds.DecodeKey(v)
}

func decodeDatastoreKeyValue(d *msgpack.Decoder, v reflect.Value) error {
	key, err := DecodeDatastoreKey(d)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(key))
	return nil
}

func encodeDatastoreCursorValue(e *msgpack.Encoder, v reflect.Value) error {
	cursor := v.Interface().(ds.Cursor)
	return e.Encode(cursor.String())
}

func decodeDatastoreCursorValue(d *msgpack.Decoder, v reflect.Value) error {
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
