package msgpack

import (
	"reflect"
	"time"
)

var (
	timeType = reflect.TypeOf((*time.Time)(nil)).Elem()
)

func init() {
	Register(timeType, EncodeTime, DecodeTime)
}

func EncodeTime(e *Encoder, v reflect.Value) error {
	tm := v.Interface().(time.Time)
	if err := e.EncodeInt64(tm.Unix()); err != nil {
		return err
	}
	return e.EncodeInt64(int64(tm.Nanosecond()))
}

func DecodeTime(d *Decoder, v reflect.Value, c byte) error {
	sec, err := d.DecodeInt64(c)
	if err != nil {
		return err
	}

	c, err = d.readByte()
	if err != nil {
		return err
	}

	nsec, err := d.DecodeInt64(c)
	if err != nil {
		return err
	}

	tm := time.Unix(sec, nsec)
	v.Set(reflect.ValueOf(tm))

	return nil
}
