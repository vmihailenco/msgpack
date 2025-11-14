package msgpack

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/gostudentorg/msgpack/v5/msgpcode"

	"gitlab.gostudent.cloud/pkg/log/errors"
)

const millisec = 1000000

var timeExtID int8 = 13
var timeExtID2 int8 = -1

func init() {
	RegisterExtEncoder(timeExtID, time.Time{}, timeEncoder)
	RegisterExtDecoder(timeExtID, time.Time{}, timeDecoder)
	RegisterExtDecoder(timeExtID2, time.Time{}, timeDecoder)
}

func timeEncoder(_ *Encoder, v reflect.Value) ([]byte, error) {
	t := v.Interface().(time.Time)
	if t.IsZero() {
		t = time.Unix(0, 0)
	}

	b := bytes.Buffer{}
	e := NewEncoder(&b)
	if r := e.EncodeFloat64(float64(t.Unix()*1000 + int64(t.Nanosecond())/millisec)); r != nil {
		return nil, r
	}

	return b.Bytes(), nil
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
	if err := e.encodeExtLen(9); err != nil {
		return err
	}
	if err := e.w.WriteByte(byte(timeExtID)); err != nil {
		return err
	}
	return e.write8(msgpcode.Double, math.Float64bits(float64(tm.Unix()*1000+int64(tm.Nanosecond())/millisec)))
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
		return stringToTime(s)
	}
	if msgpcode.IsFloat(c) {
		f, err := d.float64(c)
		if err != nil {
			return time.Time{}, err
		}

		return time.Unix(0, int64(f*millisec)), nil
	}

	extID, extLen, err := d.extHeader(c)
	if err != nil {
		return time.Time{}, err
	}

	if extID != timeExtID && extID != timeExtID2 {
		return time.Time{}, errors.Errorf("msgpack: invalid time ext id=%d", extID)
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
		t := time.Time{}
		if r := unmarshalTime(b, &t); r != nil {
			return time.Time{}, r
		}
		return t, nil
	}
}

func unmarshalTime(data []byte, d *time.Time) error {
	if len(data) == 0 {
		return nil
	}

	var v interface{}
	if err := Unmarshal(data, &v); err != nil {
		return err
	}

	switch val := v.(type) {
	case float64:
		if val == 0 {
			return nil
		}
		*d = time.Unix(0, int64(val*millisec))
	case int64:
		*d = time.Unix(val/1000, (val%1000)*millisec)
	case string:
		t, err := stringToTime(val)
		if err != nil {
			return errors.Wrap(err, errors.Msg("unmarshalTime: string layout not implemented"),
				errors.Fields{"value": val})
		}
		*d = t
	case *time.Time:
		*d = *val
	case uint64:
	case uint32:
		*d = time.Unix(int64(val), 0)
	default:
		return errors.New("unmarshalTime: unimplemented type", errors.Fields{
			"type":      reflect.TypeOf(v).String(),
			"value":     string(data),
			"value_hex": fmt.Sprintf("%x", data),
		})
	}

	return nil
}
