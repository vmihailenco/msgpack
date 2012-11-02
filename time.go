package msgpack

import (
	"reflect"
	"time"
)

var (
	timeType = reflect.TypeOf((*time.Time)(nil)).Elem()
)

func init() {
	Register(timeType, encodeTime, decodeTime)
}

func encodeTime(e *Encoder, v reflect.Value) error {
	tm := v.Interface().(time.Time)
	if err := e.EncodeInt64(tm.Unix()); err != nil {
		return err
	}
	return e.EncodeInt64(int64(tm.Nanosecond()))
}

func decodeTime(d *Decoder, v reflect.Value) error {
	var sec, nsec int64
	if err := d.DecodeInt64(&sec); err != nil {
		return err
	}
	if err := d.DecodeInt64(&nsec); err != nil {
		return err
	}
	tm := time.Unix(sec, nsec)
	v.Set(reflect.ValueOf(tm))
	return nil
}
