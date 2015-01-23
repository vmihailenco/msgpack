package msgpack

import (
	"bytes"
	"io"
	"reflect"
	"time"
)

type writer interface {
	io.Writer
	WriteByte(byte) error
	WriteString(string) (int, error)
}

type byteWriter struct {
	io.Writer
}

func (w *byteWriter) WriteByte(b byte) error {
	n, err := w.Write([]byte{b})
	if err != nil {
		return err
	}
	if n != 1 {
		return io.ErrShortWrite
	}
	return nil
}

func (w *byteWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func Marshal(v ...interface{}) ([]byte, error) {
	if len(v) == 1 {
		marshaler, ok := v[0].(Marshaler)
		if ok {
			return marshaler.MarshalMsgpack()
		}
	}
	buf := &bytes.Buffer{}
	err := NewEncoder(buf).Encode(v...)
	return buf.Bytes(), err
}

type Encoder struct {
	w   writer
	buf []byte
	m   *structCache
}

func NewEncoder(w io.Writer) *Encoder {
	ww, ok := w.(writer)
	if !ok {
		ww = &byteWriter{Writer: w}
	}
	return &Encoder{
		w:   ww,
		buf: make([]byte, 9),
		m:   newStructCache(),
	}
}

func (e *Encoder) Encode(v ...interface{}) error {
	for _, vv := range v {
		if err := e.encode(vv); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) encode(iv interface{}) error {
	if iv == nil {
		return e.EncodeNil()
	}

	switch v := iv.(type) {
	case string:
		return e.EncodeString(v)
	case []byte:
		return e.EncodeBytes(v)
	case int:
		return e.EncodeInt64(int64(v))
	case int64:
		return e.EncodeInt64(v)
	case uint:
		return e.EncodeUint64(uint64(v))
	case uint64:
		return e.EncodeUint64(v)
	case bool:
		return e.EncodeBool(v)
	case float32:
		return e.EncodeFloat32(v)
	case float64:
		return e.EncodeFloat64(v)
	case []string:
		return e.encodeStringSlice(v)
	case map[string]string:
		return e.encodeMapStringString(v)
	case time.Duration:
		return e.EncodeInt64(int64(v))
	case time.Time:
		return e.EncodeTime(v)
	case Marshaler:
		b, err := v.MarshalMsgpack()
		if err != nil {
			return err
		}
		_, err = e.w.Write(b)
		return err
	}
	return e.EncodeValue(reflect.ValueOf(iv))
}

func (e *Encoder) EncodeValue(v reflect.Value) error {
	encode := e.m.getEncoder(v.Type())
	return encode(e, v)
}

func (e *Encoder) EncodeNil() error {
	return e.w.WriteByte(nilCode)
}

func (e *Encoder) EncodeBool(value bool) error {
	if value {
		return e.w.WriteByte(trueCode)
	}
	return e.w.WriteByte(falseCode)
}

func (e *Encoder) write(b []byte) error {
	n, err := e.w.Write(b)
	if err != nil {
		return err
	}
	if n < len(b) {
		return io.ErrShortWrite
	}
	return nil
}

func (e *Encoder) writeString(s string) error {
	n, err := e.w.WriteString(s)
	if err != nil {
		return err
	}
	if n < len(s) {
		return io.ErrShortWrite
	}
	return nil
}
