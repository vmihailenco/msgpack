package msgpack

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"time"

	"github.com/vmihailenco/msgpack/codes"
)

var timeExtId int8 = -1

func init() {
	timeType := reflect.TypeOf((*time.Time)(nil)).Elem()
	registerExt(timeExtId, timeType, encodeTimeValue, decodeTimeValue)
}

func (e *Encoder) EncodeTime(tm time.Time) error {
	b := e.encodeTime(tm)
	if err := e.encodeExtLen(len(b)); err != nil {
		return err
	}
	if err := e.w.WriteByte(byte(timeExtId)); err != nil {
		return err
	}
	return e.write(b)
}

func (e *Encoder) encodeTime(tm time.Time) []byte {
	secs := uint64(tm.Unix())
	if secs>>34 == 0 {
		data := uint64(tm.Nanosecond())<<34 | secs
		if data&0xffffffff00000000 == 0 {
			b := make([]byte, 4)
			binary.BigEndian.PutUint32(b, uint32(data))
			return b
		} else {
			b := make([]byte, 8)
			binary.BigEndian.PutUint64(b, data)
			return b
		}
	}

	b := make([]byte, 12)
	binary.BigEndian.PutUint32(b, uint32(tm.Nanosecond()))
	binary.BigEndian.PutUint64(b[4:], uint64(secs))
	return b
}

func (d *Decoder) DecodeTime() (time.Time, error) {
	c, err := d.readByte()
	if err != nil {
		return time.Time{}, err
	}

	// Legacy format.
	if c == codes.FixedArrayLow|2 {
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

	extLen, err := d.parseExtLen(c)
	if err != nil {
		return time.Time{}, err
	}

	_, err = d.r.ReadByte()
	if err != nil {
		return time.Time{}, nil
	}

	b, err := d.readN(extLen)
	if err != nil {
		return time.Time{}, err
	}

	switch len(b) {
	case 4:
		sec := binary.BigEndian.Uint32(b)
		return time.Unix(int64(sec), 0), nil
	case 8:
		sec := binary.BigEndian.Uint64(b)
		nsec := int64(sec >> 34)
		sec &= 0x00000003ffffffff
		return time.Unix(int64(sec), nsec), nil
	case 12:
		nsec := binary.BigEndian.Uint32(b)
		sec := binary.BigEndian.Uint64(b[4:])
		return time.Unix(int64(sec), int64(nsec)), nil
	default:
		return time.Time{}, fmt.Errorf("msgpack: invalid ext len=%d decoding time", extLen)
	}
}

func encodeTimeValue(e *Encoder, v reflect.Value) error {
	tm := v.Interface().(time.Time)
	b := e.encodeTime(tm)
	return e.write(b)
}

func decodeTimeValue(d *Decoder, v reflect.Value) error {
	tm, err := d.DecodeTime()
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(tm))
	return nil
}
