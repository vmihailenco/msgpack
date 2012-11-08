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

func EncodeTime(e *Encoder, tm time.Time) error {
	if err := e.EncodeInt64(tm.Unix()); err != nil {
		return err
	}
	return e.EncodeInt(tm.Nanosecond())
}

func encodeTime(e *Encoder, v reflect.Value) error {
	tm := v.Interface().(time.Time)
	return EncodeTime(e, tm)
}

func DecodeTime(d *Decoder, tm *time.Time) error {
	var sec, nsec int64
	if err := d.DecodeInt64(&sec); err != nil {
		return err
	}
	if err := d.DecodeInt64(&nsec); err != nil {
		return err
	}
	*tm = time.Unix(sec, nsec)
	return nil
}

func decodeTime(d *Decoder, v reflect.Value) error {
	var tm time.Time
	if err := DecodeTime(d, &tm); err != nil {
		return err
	}
	v.Set(reflect.ValueOf(tm))
	return nil
}
