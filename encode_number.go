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
		return e.w.WriteByte(byte(v))
	case v < 256:
		return e.write1(uint8Code, v)
	case v < 65536:
		return e.write2(uint16Code, v)
	case v < 4294967296:
		return e.write4(uint32Code, v)
	default:
		return e.write8(uint64Code, v)
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
		return e.write8(int64Code, uint64(v))
	case v < -32768 || v >= 32768:
		return e.write4(int32Code, uint64(v))
	case v < -128 || v >= 128:
		return e.write2(int16Code, uint64(v))
	case v < -32:
		return e.write1(int8Code, uint64(v))
	default:
		return e.w.WriteByte(byte(v))
	}
	panic("not reached")
}

func (e *Encoder) EncodeFloat32(n float32) error {
	return e.write4(floatCode, uint64(math.Float32bits(n)))
}

func (e *Encoder) EncodeFloat64(n float64) error {
	return e.write8(doubleCode, math.Float64bits(n))
}

func (e *Encoder) write1(code byte, n uint64) error {
	e.buf = e.buf[:2]
	e.buf[0] = code
	e.buf[1] = byte(n)
	return e.write(e.buf)
}

func (e *Encoder) write2(code byte, n uint64) error {
	e.buf = e.buf[:3]
	e.buf[0] = code
	e.buf[1] = byte(n >> 8)
	e.buf[2] = byte(n)
	return e.write(e.buf)
}

func (e *Encoder) write4(code byte, n uint64) error {
	e.buf = e.buf[:5]
	e.buf[0] = code
	e.buf[1] = byte(n >> 24)
	e.buf[2] = byte(n >> 16)
	e.buf[3] = byte(n >> 8)
	e.buf[4] = byte(n)
	return e.write(e.buf)
}

func (e *Encoder) write8(code byte, n uint64) error {
	e.buf = e.buf[:9]
	e.buf[0] = code
	e.buf[1] = byte(n >> 56)
	e.buf[2] = byte(n >> 48)
	e.buf[3] = byte(n >> 40)
	e.buf[4] = byte(n >> 32)
	e.buf[5] = byte(n >> 24)
	e.buf[6] = byte(n >> 16)
	e.buf[7] = byte(n >> 8)
	e.buf[8] = byte(n)
	return e.write(e.buf)
}
