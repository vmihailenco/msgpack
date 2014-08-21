package msgpack

import (
	"fmt"
	"math"
	"reflect"
)

func (d *Decoder) uint16() (uint16, error) {
	b, err := d.R.ReadN(2)
	if err != nil {
		return 0, err
	}
	return (uint16(b[0]) << 8) | uint16(b[1]), nil
}

func (d *Decoder) uint32() (uint32, error) {
	b, err := d.R.ReadN(4)
	if err != nil {
		return 0, err
	}
	n := (uint32(b[0]) << 24) |
		(uint32(b[1]) << 16) |
		(uint32(b[2]) << 8) |
		uint32(b[3])
	return n, nil
}

func (d *Decoder) uint64() (uint64, error) {
	b, err := d.R.ReadN(8)
	if err != nil {
		return 0, err
	}
	n := (uint64(b[0]) << 56) |
		(uint64(b[1]) << 48) |
		(uint64(b[2]) << 40) |
		(uint64(b[3]) << 32) |
		(uint64(b[4]) << 24) |
		(uint64(b[5]) << 16) |
		(uint64(b[6]) << 8) |
		uint64(b[7])
	return n, nil
}

func (d *Decoder) DecodeUint64() (uint64, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode {
		return uint64(c), nil
	}
	switch c {
	case uint8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return uint64(c), nil
	case uint16Code:
		b, err := d.R.ReadN(2)
		if err != nil {
			return 0, err
		}
		return (uint64(b[0]) << 8) | uint64(b[1]), nil
	case uint32Code:
		b, err := d.R.ReadN(4)
		if err != nil {
			return 0, err
		}
		v := (uint64(b[0]) << 24) |
			(uint64(b[1]) << 16) |
			(uint64(b[2]) << 8) |
			uint64(b[3])
		return v, nil
	case uint64Code:
		b, err := d.R.ReadN(8)
		if err != nil {
			return 0, err
		}
		v := (uint64(b[0]) << 56) |
			(uint64(b[1]) << 48) |
			(uint64(b[2]) << 40) |
			(uint64(b[3]) << 32) |
			(uint64(b[4]) << 24) |
			(uint64(b[5]) << 16) |
			(uint64(b[6]) << 8) |
			uint64(b[7])
		return v, nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding uint64", c)
}

func (d *Decoder) uint64Value(value reflect.Value) error {
	v, err := d.DecodeUint64()
	if err != nil {
		return err
	}
	value.SetUint(v)
	return nil
}

func (d *Decoder) DecodeInt64() (int64, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode || c >= negFixNumLowCode {
		return int64(int8(c)), nil
	}
	switch c {
	case int8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return int64(int8(c)), nil
	case int16Code:
		b, err := d.R.ReadN(2)
		if err != nil {
			return 0, err
		}
		return int64((int16(b[0]) << 8) | int16(b[1])), nil
	case int32Code:
		b, err := d.R.ReadN(4)
		if err != nil {
			return 0, err
		}
		v := int64((int32(b[0]) << 24) |
			(int32(b[1]) << 16) |
			(int32(b[2]) << 8) |
			int32(b[3]))
		return v, nil
	case int64Code:
		b, err := d.R.ReadN(8)
		if err != nil {
			return 0, err
		}
		v := (int64(b[0]) << 56) |
			(int64(b[1]) << 48) |
			(int64(b[2]) << 40) |
			(int64(b[3]) << 32) |
			(int64(b[4]) << 24) |
			(int64(b[5]) << 16) |
			(int64(b[6]) << 8) |
			int64(b[7])
		return v, nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding int64", c)
}

func (d *Decoder) int64Value(value reflect.Value) error {
	v, err := d.DecodeInt64()
	if err != nil {
		return err
	}
	value.SetInt(v)
	return nil
}

func (d *Decoder) DecodeFloat32() (float32, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c != floatCode {
		return 0, fmt.Errorf("msgpack: invalid code %x decoding float32", c)
	}
	b, err := d.uint32()
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(b), nil
}

func (d *Decoder) float32Value(value reflect.Value) error {
	v, err := d.DecodeFloat32()
	if err != nil {
		return err
	}
	value.SetFloat(float64(v))
	return nil
}

func (d *Decoder) DecodeFloat64() (float64, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c != doubleCode {
		return 0, fmt.Errorf("msgpack: invalid code %x decoding float64", c)
	}
	b, err := d.uint64()
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(b), nil
}

func (d *Decoder) float64Value(value reflect.Value) error {
	v, err := d.DecodeFloat64()
	if err != nil {
		return err
	}
	value.SetFloat(v)
	return nil
}

func (d *Decoder) DecodeInt() (int, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode || c >= negFixNumLowCode {
		return int(int8(c)), nil
	}
	switch c {
	case int8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return int(int8(c)), nil
	case int16Code:
		b, err := d.R.ReadN(2)
		if err != nil {
			return 0, err
		}
		return int((int16(b[0]) << 8) | int16(b[1])), nil
	case int32Code:
		b, err := d.R.ReadN(4)
		if err != nil {
			return 0, err
		}
		v := int((int32(b[0]) << 24) |
			(int32(b[1]) << 16) |
			(int32(b[2]) << 8) |
			int32(b[3]))
		return v, nil
	case int64Code:
		b, err := d.R.ReadN(8)
		if err != nil {
			return 0, err
		}
		v := int((int64(b[0]) << 56) |
			(int64(b[1]) << 48) |
			(int64(b[2]) << 40) |
			(int64(b[3]) << 32) |
			(int64(b[4]) << 24) |
			(int64(b[5]) << 16) |
			(int64(b[6]) << 8) |
			int64(b[7]))
		return v, nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding int64", c)
}

func (d *Decoder) DecodeInt8() (int8, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode || c >= negFixNumLowCode {
		return int8(c), nil
	}
	switch c {
	case int8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return int8(c), nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding int8", c)
}

func (d *Decoder) DecodeInt16() (int16, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode || c >= negFixNumLowCode {
		return int16(int8(c)), nil
	}
	switch c {
	case int8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return int16(int8(c)), nil
	case int16Code:
		b, err := d.R.ReadN(2)
		if err != nil {
			return 0, err
		}
		return (int16(b[0]) << 8) | int16(b[1]), nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding int16", c)
}

func (d *Decoder) DecodeInt32() (int32, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode || c >= negFixNumLowCode {
		return int32(int8(c)), nil
	}
	switch c {
	case int8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return int32(int8(c)), nil
	case int16Code:
		b, err := d.R.ReadN(2)
		if err != nil {
			return 0, err
		}
		return int32((int16(b[0]) << 8) | int16(b[1])), nil
	case int32Code:
		b, err := d.R.ReadN(4)
		if err != nil {
			return 0, err
		}
		v := (int32(b[0]) << 24) |
			(int32(b[1]) << 16) |
			(int32(b[2]) << 8) |
			int32(b[3])
		return v, nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding int32", c)
}

func (d *Decoder) DecodeUint() (uint, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode {
		return uint(c), nil
	}
	switch c {
	case uint8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return uint(c), nil
	case uint16Code:
		b, err := d.R.ReadN(2)
		if err != nil {
			return 0, err
		}
		return (uint(b[0]) << 8) | uint(b[1]), nil
	case uint32Code:
		b, err := d.R.ReadN(4)
		if err != nil {
			return 0, err
		}
		v := (uint(b[0]) << 24) |
			(uint(b[1]) << 16) |
			(uint(b[2]) << 8) |
			uint(b[3])
		return v, nil
	case uint64Code:
		b, err := d.R.ReadN(8)
		if err != nil {
			return 0, err
		}
		v := (uint(b[0]) << 56) |
			(uint(b[1]) << 48) |
			(uint(b[2]) << 40) |
			(uint(b[3]) << 32) |
			(uint(b[4]) << 24) |
			(uint(b[5]) << 16) |
			(uint(b[6]) << 8) |
			uint(b[7])
		return v, nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding uint", c)
}

func (d *Decoder) DecodeUint8() (uint8, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode {
		return uint8(c), nil
	}
	switch c {
	case uint8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return uint8(c), nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding uint8", c)
}

func (d *Decoder) DecodeUint16() (uint16, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode {
		return uint16(c), nil
	}
	switch c {
	case uint8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return uint16(c), nil
	case uint16Code:
		b, err := d.R.ReadN(2)
		if err != nil {
			return 0, err
		}
		return (uint16(b[0]) << 8) | uint16(b[1]), nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding uint16", c)
}

func (d *Decoder) DecodeUint32() (uint32, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c <= posFixNumHighCode {
		return uint32(c), nil
	}
	switch c {
	case uint8Code:
		c, err := d.R.ReadByte()
		if err != nil {
			return 0, err
		}
		return uint32(c), nil
	case uint16Code:
		b, err := d.R.ReadN(2)
		if err != nil {
			return 0, err
		}
		return (uint32(b[0]) << 8) | uint32(b[1]), nil
	case uint32Code:
		b, err := d.R.ReadN(4)
		if err != nil {
			return 0, err
		}
		v := (uint32(b[0]) << 24) |
			(uint32(b[1]) << 16) |
			(uint32(b[2]) << 8) |
			uint32(b[3])
		return v, nil
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding uint32", c)
}
