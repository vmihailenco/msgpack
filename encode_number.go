package msgpack

import "math"

func (e *Encoder) EncodeUint(v uint) error {
	return e.EncodeUint64(uint64(v))
}

func (e *Encoder) EncodeUint8(v uint8) error {
	return e.EncodeUint64(uint64(v))
}

func (e *Encoder) EncodeUint16(v uint16) error {
	return e.EncodeUint64(uint64(v))
}

func (e *Encoder) EncodeUint32(v uint32) error {
	return e.EncodeUint64(uint64(v))
}

func (e *Encoder) EncodeUint64(v uint64) error {
	switch {
	case v < 128:
		return e.W.WriteByte(byte(v))
	case v < 256:
		return e.write([]byte{uint8Code, byte(v)})
	case v < 65536:
		return e.write([]byte{uint16Code, byte(v >> 8), byte(v)})
	case v < 4294967296:
		return e.write([]byte{
			uint32Code,
			byte(v >> 24),
			byte(v >> 16),
			byte(v >> 8),
			byte(v),
		})
	default:
		return e.write([]byte{
			uint64Code,
			byte(v >> 56),
			byte(v >> 48),
			byte(v >> 40),
			byte(v >> 32),
			byte(v >> 24),
			byte(v >> 16),
			byte(v >> 8),
			byte(v),
		})
	}
	panic("not reached")
}

func (e *Encoder) EncodeInt(v int) error {
	return e.EncodeInt64(int64(v))
}

func (e *Encoder) EncodeInt8(v int8) error {
	return e.EncodeInt64(int64(v))
}

func (e *Encoder) EncodeInt16(v int16) error {
	return e.EncodeInt64(int64(v))
}

func (e *Encoder) EncodeInt32(v int32) error {
	return e.EncodeInt64(int64(v))
}

func (e *Encoder) EncodeInt64(v int64) error {
	switch {
	case v < -2147483648 || v >= 2147483648:
		return e.write([]byte{
			int64Code,
			byte(v >> 56),
			byte(v >> 48),
			byte(v >> 40),
			byte(v >> 32),
			byte(v >> 24),
			byte(v >> 16),
			byte(v >> 8),
			byte(v),
		})
	case v < -32768 || v >= 32768:
		return e.write([]byte{
			int32Code,
			byte(v >> 24),
			byte(v >> 16),
			byte(v >> 8),
			byte(v),
		})
	case v < -128 || v >= 128:
		return e.write([]byte{int16Code, byte(v >> 8), byte(v)})
	case v < -32:
		return e.write([]byte{int8Code, byte(v)})
	default:
		return e.W.WriteByte(byte(v))
	}
	panic("not reached")
}

func (e *Encoder) EncodeFloat32(value float32) error {
	v := math.Float32bits(value)
	return e.write([]byte{
		floatCode,
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v),
	})
}

func (e *Encoder) EncodeFloat64(value float64) error {
	v := math.Float64bits(value)
	return e.write([]byte{
		doubleCode,
		byte(v >> 56),
		byte(v >> 48),
		byte(v >> 40),
		byte(v >> 32),
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v),
	})
}
