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

func DecodeTime(d *Decoder) (time.Time, error) {
	sec, err := d.DecodeInt64()
	if err != nil {
		return time.Time{}, err
	}
	nsec, err := d.DecodeInt64()
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(sec, nsec), nil
}

func decodeTime(d *Decoder, v reflect.Value) error {
	tm, err := DecodeTime(d)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(tm))
	return nil
}
