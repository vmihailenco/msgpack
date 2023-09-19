package msgpack

import (
	"fmt"
	"reflect"
	"time"

	"github.com/vmihailenco/msgpack/v6/msgpcode"
)

var timeExtID int8 = -1

func init() {
	RegisterExtEncoder(timeExtID, time.Time{}, timeEncoder)
	RegisterExtDecoder(timeExtID, time.Time{}, timeDecoder)
}

func timeEncoder(e *Encoder, v reflect.Value) ([]byte, error) {
	return e.encodeTime(v.Interface().(time.Time))
}

func timeDecoder(d *Decoder, v reflect.Value, extLen int) error {
	tm, err := d.decodeTime(extLen)
	if err != nil {
		return err
	}

	ptr := v.Addr().Interface().(*time.Time)
	*ptr = tm

	return nil
}

func (e *Encoder) EncodeTime(tm time.Time) error {
	b, err := e.encodeTime(tm)
	if err != nil {
		return err
	}
	if err = e.encodeExtLen(len(b)); err != nil {
		return err
	}
	if err = e.w.WriteByte(byte(timeExtID)); err != nil {
		return err
	}
	return e.write(b)
}

func (e *Encoder) encodeTime(tm time.Time) ([]byte, error) {
	return tm.MarshalBinary()
}

func (d *Decoder) DecodeTime() (time.Time, error) {
	c, err := d.readCode()
	if err != nil {
		return time.Time{}, err
	}

	// Legacy format.
	if c == msgpcode.FixedArrayLow|2 {
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

	if msgpcode.IsString(c) {
		s, err := d.string(c)
		if err != nil {
			return time.Time{}, err
		}
		return time.Parse(time.RFC3339Nano, s)
	}

	extID, extLen, err := d.extHeader(c)
	if err != nil {
		return time.Time{}, err
	}

	// NodeJS seems to use extID 13.
	if extID != timeExtID && extID != 13 {
		return time.Time{}, fmt.Errorf("msgpack: invalid time ext id=%d", extID)
	}

	tm, err := d.decodeTime(extLen)
	if err != nil {
		return tm, err
	}

	if tm.IsZero() {
		// Zero time does not have timezone information.
		return tm.UTC(), nil
	}
	return tm, nil
}

func (d *Decoder) decodeTime(extLen int) (time.Time, error) {
	b, err := d.readN(extLen)
	if err != nil {
		return time.Time{}, err
	}

	tm := time.Time{}
	err = tm.UnmarshalBinary(b)
	return tm, err
}
