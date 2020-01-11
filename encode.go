package msgpack

import (
	"bytes"
	"io"
	"reflect"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack/v4/codes"
)

type writer interface {
	io.Writer
	WriteByte(byte) error
}

type byteWriter struct {
	io.Writer

	buf [1]byte
}

func newByteWriter(w io.Writer) *byteWriter {
	bw := new(byteWriter)
	bw.Reset(w)
	return bw
}

func (bw *byteWriter) Reset(w io.Writer) {
	bw.Writer = w
}

func (bw *byteWriter) WriteByte(c byte) error {
	bw.buf[0] = c
	_, err := bw.Write(bw.buf[:])
	return err
}

//------------------------------------------------------------------------------

var encPool = sync.Pool{
	New: func() interface{} {
		return NewEncoder(nil)
	},
}

// Marshal returns the MessagePack encoding of v.
func Marshal(v interface{}) ([]byte, error) {
	enc := encPool.Get().(*Encoder)

	var buf bytes.Buffer
	enc.Reset(&buf)

	err := NewEncoder(&buf).Encode(v)
	b := buf.Bytes()

	encPool.Put(enc)

	return b, err
}

type Encoder struct {
	w writer

	buf       []byte
	timeBuf   []byte
	bootstrap [9 + 12]byte

	intern map[string]int

	sortMapKeys   bool
	structAsArray bool
	useJSONTag    bool
	useCompact    bool
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	e := new(Encoder)
	e.buf = e.bootstrap[:9]
	e.timeBuf = e.bootstrap[9 : 9+12]
	e.Reset(w)
	return e
}

func (e *Encoder) Reset(w io.Writer) {
	if bw, ok := w.(writer); ok {
		e.w = bw
	} else if bw, ok := e.w.(*byteWriter); ok {
		bw.Reset(w)
	} else {
		e.w = newByteWriter(w)
	}

	for k := range e.intern {
		delete(e.intern, k)
	}
}

// SortMapKeys causes the Encoder to encode map keys in increasing order.
// Supported map types are:
//   - map[string]string
//   - map[string]interface{}
func (e *Encoder) SortMapKeys(flag bool) *Encoder {
	e.sortMapKeys = flag
	return e
}

// StructAsArray causes the Encoder to encode Go structs as msgpack arrays.
func (e *Encoder) StructAsArray(flag bool) *Encoder {
	e.structAsArray = flag
	return e
}

// UseJSONTag causes the Encoder to use json struct tag as fallback option
// if there is no msgpack tag.
func (e *Encoder) UseJSONTag(flag bool) *Encoder {
	e.useJSONTag = flag
	return e
}

// UseCompactEncoding causes the Encoder to chose the most compact encoding.
// For example, it allows to encode small Go int64 as msgpack int8 saving 7 bytes.
func (e *Encoder) UseCompactEncoding(flag bool) *Encoder {
	e.useCompact = flag
	return e
}

func (e *Encoder) Encode(v interface{}) error {
	switch v := v.(type) {
	case nil:
		return e.EncodeNil()
	case string:
		return e.EncodeString(v)
	case []byte:
		return e.EncodeBytes(v)
	case int:
		return e.encodeInt64Cond(int64(v))
	case int64:
		return e.encodeInt64Cond(v)
	case uint:
		return e.encodeUint64Cond(uint64(v))
	case uint64:
		return e.encodeUint64Cond(v)
	case bool:
		return e.EncodeBool(v)
	case float32:
		return e.EncodeFloat32(v)
	case float64:
		return e.EncodeFloat64(v)
	case time.Duration:
		return e.encodeInt64Cond(int64(v))
	case time.Time:
		return e.EncodeTime(v)
	}
	return e.EncodeValue(reflect.ValueOf(v))
}

func (e *Encoder) EncodeMulti(v ...interface{}) error {
	for _, vv := range v {
		if err := e.Encode(vv); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) EncodeValue(v reflect.Value) error {
	fn := getEncoder(v.Type())
	return fn(e, v)
}

func (e *Encoder) EncodeNil() error {
	return e.writeCode(codes.Nil)
}

func (e *Encoder) EncodeBool(value bool) error {
	if value {
		return e.writeCode(codes.True)
	}
	return e.writeCode(codes.False)
}

func (e *Encoder) writeCode(c codes.Code) error {
	return e.w.WriteByte(byte(c))
}

func (e *Encoder) write(b []byte) error {
	_, err := e.w.Write(b)
	return err
}

func (e *Encoder) writeString(s string) error {
	_, err := e.w.Write(stringToBytes(s))
	return err
}
