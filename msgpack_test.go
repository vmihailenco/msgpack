package msgpack_test

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"encoding/gob"
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	gomsgpack "github.com/ugorji/go-msgpack"
	gocodec "github.com/ugorji/go/codec"
	. "gopkg.in/check.v1"

	"gopkg.in/vmihailenco/msgpack.v2"
	"gopkg.in/vmihailenco/msgpack.v2/codes"
)

type nameStruct struct {
	Name string
}

func Test(t *testing.T) { TestingT(t) }

type MsgpackTest struct {
	buf *bytes.Buffer
	enc *msgpack.Encoder
	dec *msgpack.Decoder
}

var _ = Suite(&MsgpackTest{})

func (t *MsgpackTest) SetUpTest(c *C) {
	t.buf = &bytes.Buffer{}
	t.enc = msgpack.NewEncoder(t.buf)
	t.dec = msgpack.NewDecoder(bufio.NewReader(t.buf))
}

func (t *MsgpackTest) TestUint64(c *C) {
	table := []struct {
		v uint64
		b []byte
	}{
		{0, []byte{0x00}},
		{1, []byte{0x01}},
		{math.MaxInt8 - 1, []byte{0x7e}},
		{math.MaxInt8, []byte{0x7f}},
		{math.MaxInt8 + 1, []byte{0xcc, 0x80}},
		{math.MaxUint8 - 1, []byte{0xcc, 0xfe}},
		{math.MaxUint8, []byte{0xcc, 0xff}},
		{math.MaxUint8 + 1, []byte{0xcd, 0x1, 0x0}},
		{math.MaxUint16 - 1, []byte{0xcd, 0xff, 0xfe}},
		{math.MaxUint16, []byte{0xcd, 0xff, 0xff}},
		{math.MaxUint16 + 1, []byte{0xce, 0x0, 0x1, 0x0, 0x0}},
		{math.MaxUint32 - 1, []byte{0xce, 0xff, 0xff, 0xff, 0xfe}},
		{math.MaxUint32, []byte{0xce, 0xff, 0xff, 0xff, 0xff}},
		{math.MaxUint32 + 1, []byte{0xcf, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0}},
		{math.MaxInt64 - 1, []byte{0xcf, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe}},
		{math.MaxInt64, []byte{0xcf, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	}
	for _, r := range table {
		var int64v int64
		c.Assert(t.enc.Encode(r.v), IsNil)
		c.Assert(t.buf.Bytes(), DeepEquals, r.b, Commentf("n=%d", r.v))
		c.Assert(t.dec.Decode(&int64v), IsNil, Commentf("n=%d", r.v))
		c.Assert(int64v, Equals, int64(r.v), Commentf("n=%d", r.v))

		var uint64v uint64
		c.Assert(t.enc.Encode(r.v), IsNil)
		c.Assert(t.buf.Bytes(), DeepEquals, r.b, Commentf("n=%d", r.v))
		c.Assert(t.dec.Decode(&uint64v), IsNil, Commentf("n=%d", r.v))
		c.Assert(uint64v, Equals, uint64(r.v), Commentf("n=%d", r.v))

		c.Assert(t.enc.Encode(r.v), IsNil)
		iface, err := t.dec.DecodeInterface()
		c.Assert(err, IsNil)
		switch iface.(type) {
		case int64:
			c.Assert(iface, Equals, int64(r.v))
		case uint64:
			c.Assert(iface, Equals, r.v)
		default:
			panic("not reached")
		}
	}
}

func (t *MsgpackTest) TestInt64(c *C) {
	table := []struct {
		v int64
		b []byte
	}{
		{math.MinInt64, []byte{0xd3, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{math.MinInt32 - 1, []byte{0xd3, 0xff, 0xff, 0xff, 0xff, 0x7f, 0xff, 0xff, 0xff}},
		{math.MinInt32, []byte{0xd2, 0x80, 0x00, 0x00, 0x00}},
		{math.MinInt32 + 1, []byte{0xd2, 0x80, 0x00, 0x00, 0x01}},
		{math.MinInt16 - 1, []byte{0xd2, 0xff, 0xff, 0x7f, 0xff}},
		{math.MinInt16, []byte{0xd1, 0x80, 0x00}},
		{math.MinInt16 + 1, []byte{0xd1, 0x80, 0x01}},
		{math.MinInt8 - 1, []byte{0xd1, 0xff, 0x7f}},
		{math.MinInt8, []byte{0xd0, 0x80}},
		{math.MinInt8 + 1, []byte{0xd0, 0x81}},
		{-33, []byte{0xd0, 0xdf}},
		{-32, []byte{0xe0}},
		{-31, []byte{0xe1}},
		{-1, []byte{0xff}},
		{0, []byte{0x00}},
		{1, []byte{0x01}},
		{math.MaxInt8 - 1, []byte{0x7e}},
		{math.MaxInt8, []byte{0x7f}},
		{math.MaxInt8 + 1, []byte{0xcc, 0x80}},
		{math.MaxUint8 - 1, []byte{0xcc, 0xfe}},
		{math.MaxUint8, []byte{0xcc, 0xff}},
		{math.MaxUint8 + 1, []byte{0xcd, 0x1, 0x0}},
		{math.MaxUint16 - 1, []byte{0xcd, 0xff, 0xfe}},
		{math.MaxUint16, []byte{0xcd, 0xff, 0xff}},
		{math.MaxUint16 + 1, []byte{0xce, 0x0, 0x1, 0x0, 0x0}},
		{math.MaxUint32 - 1, []byte{0xce, 0xff, 0xff, 0xff, 0xfe}},
		{math.MaxUint32, []byte{0xce, 0xff, 0xff, 0xff, 0xff}},
		{math.MaxUint32 + 1, []byte{0xcf, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0}},
		{math.MaxInt64 - 1, []byte{0xcf, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe}},
		{math.MaxInt64, []byte{0xcf, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	}
	for _, r := range table {
		var int64v int64
		c.Assert(t.enc.Encode(r.v), IsNil)
		c.Assert(t.buf.Bytes(), DeepEquals, r.b, Commentf("n=%d", r.v))
		c.Assert(t.dec.Decode(&int64v), IsNil, Commentf("n=%d", r.v))
		c.Assert(int64v, Equals, int64(r.v), Commentf("n=%d", r.v))

		var uint64v uint64
		c.Assert(t.enc.Encode(r.v), IsNil)
		c.Assert(t.buf.Bytes(), DeepEquals, r.b, Commentf("n=%d", r.v))
		c.Assert(t.dec.Decode(&uint64v), IsNil, Commentf("n=%d", r.v))
		c.Assert(uint64v, Equals, uint64(r.v), Commentf("n=%d", r.v))

		c.Assert(t.enc.Encode(r.v), IsNil)
		iface, err := t.dec.DecodeInterface()
		c.Assert(err, IsNil)
		switch iface.(type) {
		case int64:
			c.Assert(iface, Equals, r.v)
		case uint64:
			c.Assert(iface, Equals, uint64(r.v))
		default:
			panic("not reached")
		}
	}
}

func (t *MsgpackTest) TestFloat32(c *C) {
	table := []struct {
		v float32
		b []byte
	}{
		{0.1, []byte{codes.Float, 0x3d, 0xcc, 0xcc, 0xcd}},
		{0.2, []byte{codes.Float, 0x3e, 0x4c, 0xcc, 0xcd}},
		{-0.1, []byte{codes.Float, 0xbd, 0xcc, 0xcc, 0xcd}},
		{-0.2, []byte{codes.Float, 0xbe, 0x4c, 0xcc, 0xcd}},
		{float32(math.Inf(1)), []byte{codes.Float, 0x7f, 0x80, 0x00, 0x00}},
		{float32(math.Inf(-1)), []byte{codes.Float, 0xff, 0x80, 0x00, 0x00}},
		{math.MaxFloat32, []byte{codes.Float, 0x7f, 0x7f, 0xff, 0xff}},
		{math.SmallestNonzeroFloat32, []byte{codes.Float, 0x0, 0x0, 0x0, 0x1}},
	}
	for _, r := range table {
		c.Assert(t.enc.Encode(r.v), IsNil)
		c.Assert(t.buf.Bytes(), DeepEquals, r.b, Commentf("err encoding %v", r.v))

		var f32 float32
		c.Assert(t.dec.Decode(&f32), IsNil)
		c.Assert(f32, Equals, r.v)

		// Pass pointer to skip fast-path and trigger reflect.
		c.Assert(t.enc.Encode(&r.v), IsNil)
		c.Assert(t.buf.Bytes(), DeepEquals, r.b, Commentf("err encoding %v", r.v))

		var f64 float64
		c.Assert(t.dec.Decode(&f64), IsNil)
		c.Assert(float32(f64), Equals, r.v)

		c.Assert(t.enc.Encode(r.v), IsNil)
		iface, err := t.dec.DecodeInterface()
		c.Assert(err, IsNil)
		c.Assert(iface, Equals, r.v)
	}

	in := float32(math.NaN())
	c.Assert(t.enc.Encode(in), IsNil)
	var out float32
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(math.IsNaN(float64(out)), Equals, true)
}

func (t *MsgpackTest) TestFloat64(c *C) {
	table := []struct {
		v float64
		b []byte
	}{
		{.1, []byte{0xcb, 0x3f, 0xb9, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9a}},
		{.2, []byte{0xcb, 0x3f, 0xc9, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9a}},
		{-.1, []byte{0xcb, 0xbf, 0xb9, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9a}},
		{-.2, []byte{0xcb, 0xbf, 0xc9, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9a}},
		{math.Inf(1), []byte{0xcb, 0x7f, 0xf0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
		{math.Inf(-1), []byte{0xcb, 0xff, 0xf0, 0x00, 0x00, 0x0, 0x0, 0x0, 0x0}},
		{math.MaxFloat64, []byte{0xcb, 0x7f, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{math.SmallestNonzeroFloat64, []byte{0xcb, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}},
	}
	for _, r := range table {
		c.Assert(t.enc.Encode(r.v), IsNil)
		c.Assert(t.buf.Bytes(), DeepEquals, r.b, Commentf("err encoding %v", r.v))

		var v float64
		c.Assert(t.dec.Decode(&v), IsNil)
		c.Assert(v, Equals, r.v)

		c.Assert(t.enc.Encode(r.v), IsNil)
		iface, err := t.dec.DecodeInterface()
		c.Assert(err, IsNil)
		c.Assert(iface, Equals, r.v)
	}

	in := math.NaN()
	c.Assert(t.enc.Encode(in), IsNil)
	var out float64
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(math.IsNaN(out), Equals, true)
}

func (t *MsgpackTest) TestBool(c *C) {
	table := []struct {
		v bool
		b []byte
	}{
		{false, []byte{0xc2}},
		{true, []byte{0xc3}},
	}
	for _, r := range table {
		c.Assert(t.enc.Encode(r.v), IsNil)
		c.Assert(t.buf.Bytes(), DeepEquals, r.b, Commentf("err encoding %v", r.v))

		var v bool
		c.Assert(t.dec.Decode(&v), IsNil)
		c.Assert(v, Equals, r.v)

		c.Assert(t.enc.Encode(r.v), IsNil)
		iface, err := t.dec.DecodeInterface()
		c.Assert(err, IsNil)
		c.Assert(iface, Equals, r.v)
	}
}

func (t *MsgpackTest) TestDecodeNil(c *C) {
	c.Assert(t.dec.Decode(nil), NotNil)
}

func (t *MsgpackTest) TestTime(c *C) {
	in := time.Now()
	var out time.Time
	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(out.Equal(in), Equals, true)

	var zero time.Time
	c.Assert(t.enc.Encode(zero), IsNil)
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(out.Equal(zero), Equals, true)
	c.Assert(out.IsZero(), Equals, true)
}

func (t *MsgpackTest) TestSliceOfInts(c *C) {
	for _, test := range []struct {
		v []int64
		b []byte
	}{
		{nil, []byte{0xc0}},
		{[]int64{}, []byte{0x90}},
		{[]int64{0}, []byte{0x91, 0x0}},
	} {
		c.Assert(t.enc.Encode(test.v), IsNil)
		c.Assert(t.buf.Bytes(), DeepEquals, test.b)
		var dst []int64
		c.Assert(t.dec.Decode(&dst), IsNil)
		c.Assert(dst, DeepEquals, test.v)
	}
}

func (t *MsgpackTest) TestArrayOfInts(c *C) {
	src := [3]int{1, 2, 3}
	c.Assert(t.enc.Encode(src), IsNil)
	var dst [3]int
	c.Assert(t.dec.Decode(&dst), IsNil)
	c.Assert(dst, DeepEquals, src)
}

func (t *MsgpackTest) TestSliceOfStrings(c *C) {
	for _, i := range []struct {
		src []string
		b   []byte
	}{
		{nil, []byte{0xc0}},
		{[]string{}, []byte{0x90}},
		{[]string{"foo", "bar"}, []byte{0x92, 0xa3, 'f', 'o', 'o', 0xa3, 'b', 'a', 'r'}},
	} {
		c.Assert(t.enc.Encode(i.src), IsNil)
		c.Assert(t.buf.Bytes(), DeepEquals, i.b)
		var dst []string
		c.Assert(t.dec.Decode(&dst), IsNil)
		c.Assert(dst, DeepEquals, i.src)
	}
}

func (t *MsgpackTest) TestArrayOfStrings(c *C) {
	src := [2]string{"hello", "world"}
	c.Assert(t.enc.Encode(src), IsNil)
	var dst [2]string
	c.Assert(t.dec.Decode(&dst), IsNil)
	c.Assert(dst, DeepEquals, src)
}

func (t *MsgpackTest) TestBin(c *C) {
	lowBin8 := []byte(strings.Repeat("w", 32))
	highBin8 := []byte(strings.Repeat("w", 255))
	lowBin16 := []byte(strings.Repeat("w", 256))
	highBin16 := []byte(strings.Repeat("w", 65535))
	lowBin32 := []byte(strings.Repeat("w", 65536))
	for _, i := range []struct {
		src []byte
		b   []byte
	}{
		{
			lowBin8,
			append([]byte{0xc4, byte(len(lowBin8))}, lowBin8...),
		},
		{
			highBin8,
			append([]byte{0xc4, byte(len(highBin8))}, highBin8...),
		},
		{
			lowBin16,
			append([]byte{0xc5, 1, 0}, lowBin16...),
		},
		{
			highBin16,
			append([]byte{0xc5, 255, 255}, highBin16...),
		},
		{
			lowBin32,
			append([]byte{0xc6, 0, 1, 0, 0}, lowBin32...),
		},
	} {
		c.Assert(t.enc.Encode(i.src), IsNil)
		c.Assert(t.buf.Bytes(), DeepEquals, i.b)

		var dst []byte
		c.Assert(t.dec.Decode(&dst), IsNil)
		c.Assert(dst, DeepEquals, i.src)

		c.Assert(t.enc.Encode(i.src), IsNil)
		iface, err := t.dec.DecodeInterface()
		c.Assert(err, IsNil)
		c.Assert(iface, DeepEquals, i.src)
	}
}

func (t *MsgpackTest) TestString(c *C) {
	highFixStr := strings.Repeat("w", 31)
	lowStr8 := strings.Repeat("w", 32)
	highStr8 := strings.Repeat("w", 255)
	lowStr16 := strings.Repeat("w", 256)
	highStr16 := strings.Repeat("w", 65535)
	lowStr32 := strings.Repeat("w", 65536)
	for _, i := range []struct {
		src string
		b   []byte
	}{
		{"", []byte{0xa0}},                          // fixstr
		{"a", []byte{0xa1, 'a'}},                    // fixstr
		{"hello", append([]byte{0xa5}, "hello"...)}, // fixstr
		{
			highFixStr,
			append([]byte{0xbf}, highFixStr...),
		},
		{
			lowStr8,
			append([]byte{0xd9, byte(len(lowStr8))}, lowStr8...),
		},
		{
			highStr8,
			append([]byte{0xd9, byte(len(highStr8))}, highStr8...),
		},
		{
			lowStr16,
			append([]byte{0xda, 1, 0}, lowStr16...),
		},
		{
			highStr16,
			append([]byte{0xda, 255, 255}, highStr16...),
		},
		{
			lowStr32,
			append([]byte{0xdb, 0, 1, 0, 0}, lowStr32...),
		},
	} {
		c.Assert(t.enc.Encode(i.src), IsNil)
		c.Assert(t.buf.Bytes(), DeepEquals, i.b)

		var dst string
		c.Assert(t.dec.Decode(&dst), IsNil)
		c.Assert(dst, Equals, i.src)

		c.Assert(t.enc.Encode(i.src), IsNil)
		iface, err := t.dec.DecodeInterface()
		c.Assert(err, IsNil)
		c.Assert(iface, DeepEquals, i.src)
	}
}

func (t *MsgpackTest) TestLargeBytes(c *C) {
	N := int(1e6)

	src := bytes.Repeat([]byte{'1'}, N)
	c.Assert(t.enc.Encode(src), IsNil)
	var dst []byte
	c.Assert(t.dec.Decode(&dst), IsNil)
	c.Assert(dst, DeepEquals, src)
}

func (t *MsgpackTest) TestLargeString(c *C) {
	N := int(1e6)

	src := string(bytes.Repeat([]byte{'1'}, N))
	c.Assert(t.enc.Encode(src), IsNil)
	var dst string
	c.Assert(t.dec.Decode(&dst), IsNil)
	c.Assert(dst, Equals, src)
}

func (t *MsgpackTest) TestSliceOfStructs(c *C) {
	in := []*nameStruct{&nameStruct{"hello"}}
	var out []*nameStruct
	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(out, DeepEquals, in)
}

func (t *MsgpackTest) TestMap(c *C) {
	for _, i := range []struct {
		m map[string]string
		b []byte
	}{
		{map[string]string{}, []byte{0x80}},
		{map[string]string{"hello": "world"}, []byte{0x81, 0xa5, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0xa5, 0x77, 0x6f, 0x72, 0x6c, 0x64}},
	} {
		c.Assert(t.enc.Encode(i.m), IsNil)
		c.Assert(t.buf.Bytes(), DeepEquals, i.b, Commentf("err encoding %v", i.m))
		var m map[string]string
		c.Assert(t.dec.Decode(&m), IsNil)
		c.Assert(m, DeepEquals, i.m)
	}
}

func (t *MsgpackTest) TestStructNil(c *C) {
	var dst *nameStruct

	c.Assert(t.enc.Encode(nameStruct{Name: "foo"}), IsNil)
	c.Assert(t.dec.Decode(&dst), IsNil)
	c.Assert(dst, Not(IsNil))
	c.Assert(dst.Name, Equals, "foo")
}

func (t *MsgpackTest) TestStructUnknownField(c *C) {
	in := struct {
		Field1 string
		Field2 string
		Field3 string
	}{
		Field1: "value1",
		Field2: "value2",
		Field3: "value3",
	}
	c.Assert(t.enc.Encode(in), IsNil)

	out := struct {
		Field2 string
	}{}
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(out.Field2, Equals, "value2")
}

//------------------------------------------------------------------------------

type coderStruct struct {
	name string
}

type wrapperStruct struct {
	coderStruct `msgpack:",inline"`
}

var (
	_ msgpack.CustomEncoder = &coderStruct{}
	_ msgpack.CustomDecoder = &coderStruct{}
)

func (s *coderStruct) Name() string {
	return s.name
}

func (s *coderStruct) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.Encode(s.name)
}

func (s *coderStruct) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.Decode(&s.name)
}

func (t *MsgpackTest) TestCoder(c *C) {
	in := &coderStruct{name: "hello"}
	var out coderStruct
	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(out.Name(), Equals, "hello")
}

func (t *MsgpackTest) TestNilCoder(c *C) {
	in := &coderStruct{name: "hello"}
	var out *coderStruct
	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(out.Name(), Equals, "hello")
}

func (t *MsgpackTest) TestNilCoderValue(c *C) {
	c.Skip("TODO")

	in := &coderStruct{name: "hello"}
	var out *coderStruct
	v := reflect.ValueOf(out)
	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.DecodeValue(v), IsNil)
	c.Assert(out.Name(), Equals, "hello")
}

func (t *MsgpackTest) TestPtrToCoder(c *C) {
	in := &coderStruct{name: "hello"}
	var out coderStruct
	out2 := &out
	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.Decode(&out2), IsNil)
	c.Assert(out.Name(), Equals, "hello")
}

func (t *MsgpackTest) TestWrappedCoder(c *C) {
	in := &wrapperStruct{coderStruct: coderStruct{name: "hello"}}
	var out wrapperStruct
	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(out.Name(), Equals, "hello")
}

//------------------------------------------------------------------------------

type struct2 struct {
	Name string
}

type struct1 struct {
	Name    string
	Struct2 struct2
}

func (t *MsgpackTest) TestNestedStructs(c *C) {
	in := &struct1{Name: "hello", Struct2: struct2{Name: "world"}}
	var out struct1
	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(out.Name, Equals, in.Name)
	c.Assert(out.Struct2.Name, Equals, in.Struct2.Name)
}

type Struct4 struct {
	Name2 string
}

type Struct3 struct {
	Struct4
	Name1 string
}

func TestEmbedding(t *testing.T) {
	in := &Struct3{
		Name1: "hello",
		Struct4: Struct4{
			Name2: "world",
		},
	}
	var out Struct3

	b, err := msgpack.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	err = msgpack.Unmarshal(b, &out)
	if err != nil {
		t.Fatal(err)
	}
	if out.Name1 != in.Name1 {
		t.Fatalf("")
	}
	if out.Name2 != in.Name2 {
		t.Fatalf("")
	}
}

func (t *MsgpackTest) TestSliceNil(c *C) {
	in := [][]*int{nil}
	var out [][]*int

	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(out, DeepEquals, in)
}

//------------------------------------------------------------------------------

func (t *MsgpackTest) TestMapStringInterface(c *C) {
	in := map[string]interface{}{
		"foo": "bar",
		"hello": map[string]interface{}{
			"foo": "bar",
		},
	}
	var out map[string]interface{}

	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.Decode(&out), IsNil)

	c.Assert(out["foo"], Equals, "bar")
	mm := out["hello"].(map[interface{}]interface{})
	c.Assert(mm["foo"], Equals, "bar")
}

func (t *MsgpackTest) TestMapStringInterface2(c *C) {
	buf := &bytes.Buffer{}
	enc := msgpack.NewEncoder(buf)
	dec := msgpack.NewDecoder(buf)
	dec.DecodeMapFunc = func(d *msgpack.Decoder) (interface{}, error) {
		n, err := d.DecodeMapLen()
		if err != nil {
			return nil, err
		}

		m := make(map[string]interface{}, n)
		for i := 0; i < n; i++ {
			mk, err := d.DecodeString()
			if err != nil {
				return nil, err
			}

			mv, err := d.DecodeInterface()
			if err != nil {
				return nil, err
			}

			m[mk] = mv
		}
		return m, nil
	}

	in := map[string]interface{}{
		"foo": "bar",
		"hello": map[string]interface{}{
			"foo": "bar",
		},
	}
	var out map[string]interface{}

	c.Assert(enc.Encode(in), IsNil)
	c.Assert(dec.Decode(&out), IsNil)

	c.Assert(out["foo"], Equals, "bar")
	mm := out["hello"].(map[string]interface{})
	c.Assert(mm["foo"], Equals, "bar")
}

//------------------------------------------------------------------------------

func benchmarkEncodeDecode(b *testing.B, src, dst interface{}) {
	buf := &bytes.Buffer{}
	enc := msgpack.NewEncoder(buf)
	dec := msgpack.NewDecoder(buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := enc.Encode(src); err != nil {
			b.Fatal(err)
		}
		if err := dec.Decode(dst); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBool(b *testing.B) {
	var dst bool
	benchmarkEncodeDecode(b, true, &dst)
}

func BenchmarkInt0(b *testing.B) {
	var dst int
	benchmarkEncodeDecode(b, 1, &dst)
}

func BenchmarkInt1(b *testing.B) {
	var dst int
	benchmarkEncodeDecode(b, -33, &dst)
}

func BenchmarkInt2(b *testing.B) {
	var dst int
	benchmarkEncodeDecode(b, 128, &dst)
}

func BenchmarkInt4(b *testing.B) {
	var dst int
	benchmarkEncodeDecode(b, 32768, &dst)
}

func BenchmarkInt8(b *testing.B) {
	var dst int
	benchmarkEncodeDecode(b, 2147483648, &dst)
}

func BenchmarkIntBinary(b *testing.B) {
	buf := &bytes.Buffer{}

	var out int32
	for i := 0; i < b.N; i++ {
		if err := binary.Write(buf, binary.BigEndian, int32(1)); err != nil {
			b.Fatal(err)
		}
		if err := binary.Read(buf, binary.BigEndian, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIntUgorjiGoMsgpack(b *testing.B) {
	buf := &bytes.Buffer{}
	dec := gomsgpack.NewDecoder(buf, nil)
	enc := gomsgpack.NewEncoder(buf)

	var out int
	for i := 0; i < b.N; i++ {
		if err := enc.Encode(1); err != nil {
			b.Fatal(err)
		}
		if err := dec.Decode(&out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIntUgorjiGoCodec(b *testing.B) {
	buf := &bytes.Buffer{}
	enc := gocodec.NewEncoder(buf, &gocodec.MsgpackHandle{})
	dec := gocodec.NewDecoder(buf, &gocodec.MsgpackHandle{})

	var out int
	for i := 0; i < b.N; i++ {
		if err := enc.Encode(1); err != nil {
			b.Fatal(err)
		}
		if err := dec.Decode(&out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTime(b *testing.B) {
	var dst time.Time
	benchmarkEncodeDecode(b, time.Now(), &dst)
}

func BenchmarkDuration(b *testing.B) {
	var dst time.Duration
	benchmarkEncodeDecode(b, time.Hour, &dst)
}

func BenchmarkBytes(b *testing.B) {
	src := make([]byte, 1024)
	var dst []byte
	benchmarkEncodeDecode(b, src, &dst)
}

func BenchmarkMapStringString(b *testing.B) {
	src := map[string]string{
		"hello": "world",
		"foo":   "bar",
	}
	var dst map[string]string
	benchmarkEncodeDecode(b, src, &dst)
}

func BenchmarkMapStringStringPtr(b *testing.B) {
	src := map[string]string{
		"hello": "world",
		"foo":   "bar",
	}
	var dst map[string]string
	dstptr := &dst
	benchmarkEncodeDecode(b, src, &dstptr)
}

func BenchmarkMapIntInt(b *testing.B) {
	src := map[int]int{
		1: 10,
		2: 20,
	}
	var dst map[int]int
	benchmarkEncodeDecode(b, src, &dst)
}

func BenchmarkStringSlice(b *testing.B) {
	src := []string{"hello", "world"}
	var dst []string
	benchmarkEncodeDecode(b, src, &dst)
}

func BenchmarkStringSlicePtr(b *testing.B) {
	src := []string{"hello", "world"}
	var dst []string
	dstptr := &dst
	benchmarkEncodeDecode(b, src, &dstptr)
}

type benchmarkStruct struct {
	Name      string
	Age       int
	Colors    []string
	Data      []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

type benchmarkStruct2 struct {
	Name      string
	Age       int
	Colors    []string
	Data      []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	_ msgpack.CustomEncoder = &benchmarkStruct2{}
	_ msgpack.CustomDecoder = &benchmarkStruct2{}
)

func (s *benchmarkStruct2) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.Encode(
		s.Name,
		s.Colors,
		s.Age,
		s.Data,
		s.CreatedAt,
		s.UpdatedAt,
	)
}

func (s *benchmarkStruct2) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.Decode(
		&s.Name,
		&s.Colors,
		&s.Age,
		&s.Data,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
}

func structForBenchmark() *benchmarkStruct {
	return &benchmarkStruct{
		Name:      "Hello World",
		Colors:    []string{"red", "orange", "yellow", "green", "blue", "violet"},
		Age:       math.MaxInt32,
		Data:      make([]byte, 1024),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func structForBenchmark2() *benchmarkStruct2 {
	return &benchmarkStruct2{
		Name:      "Hello World",
		Colors:    []string{"red", "orange", "yellow", "green", "blue", "violet"},
		Age:       math.MaxInt32,
		Data:      make([]byte, 1024),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func BenchmarkStructVmihailencoMsgpack(b *testing.B) {
	in := structForBenchmark()
	out := &benchmarkStruct{}
	for i := 0; i < b.N; i++ {
		buf, err := msgpack.Marshal(in)
		if err != nil {
			b.Fatal(err)
		}

		err = msgpack.Unmarshal(buf, out)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStructMarshal(b *testing.B) {
	in := structForBenchmark()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := msgpack.Marshal(in)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStructUnmarshal(b *testing.B) {
	in := structForBenchmark()
	buf, err := msgpack.Marshal(in)
	if err != nil {
		b.Fatal(err)
	}
	out := &benchmarkStruct{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = msgpack.Unmarshal(buf, out)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStructManual(b *testing.B) {
	in := structForBenchmark2()
	out := &benchmarkStruct2{}
	for i := 0; i < b.N; i++ {
		buf, err := msgpack.Marshal(in)
		if err != nil {
			b.Fatal(err)
		}

		err = msgpack.Unmarshal(buf, out)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStructUgorjiGoMsgpack(b *testing.B) {
	in := structForBenchmark()
	out := &benchmarkStruct{}
	for i := 0; i < b.N; i++ {
		buf, err := gomsgpack.Marshal(in)
		if err != nil {
			b.Fatal(err)
		}

		err = gomsgpack.Unmarshal(buf, out, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStructUgorjiGoCodec(b *testing.B) {
	in := structForBenchmark()
	out := &benchmarkStruct{}
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		enc := gocodec.NewEncoder(buf, &gocodec.MsgpackHandle{})
		dec := gocodec.NewDecoder(buf, &gocodec.MsgpackHandle{})

		if err := enc.Encode(in); err != nil {
			b.Fatal(err)
		}
		if err := dec.Decode(out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStructJSON(b *testing.B) {
	in := structForBenchmark()
	out := &benchmarkStruct{}
	for i := 0; i < b.N; i++ {
		buf, err := json.Marshal(in)
		if err != nil {
			b.Fatal(err)
		}

		err = json.Unmarshal(buf, out)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStructGOB(b *testing.B) {
	in := structForBenchmark()
	out := &benchmarkStruct{}
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		enc := gob.NewEncoder(buf)
		dec := gob.NewDecoder(buf)

		if err := enc.Encode(in); err != nil {
			b.Fatal(err)
		}

		if err := dec.Decode(out); err != nil {
			b.Fatal(err)
		}
	}
}

type benchmarkSubStruct struct {
	Name string
	Age  int
}

func BenchmarkStructUnmarshalPartially(b *testing.B) {
	in := structForBenchmark()
	buf, err := msgpack.Marshal(in)
	if err != nil {
		b.Fatal(err)
	}
	out := &benchmarkSubStruct{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = msgpack.Unmarshal(buf, out)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCSV(b *testing.B) {
	for i := 0; i < b.N; i++ {
		record := []string{"1", "hello", "world"}

		buf := &bytes.Buffer{}
		r := csv.NewReader(buf)
		w := csv.NewWriter(buf)

		if err := w.Write(record); err != nil {
			b.Fatal(err)
		}
		w.Flush()
		if _, err := r.Read(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCSVMsgpack(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var num int
		var hello, world string

		buf := &bytes.Buffer{}
		enc := msgpack.NewEncoder(buf)
		dec := msgpack.NewDecoder(buf)

		if err := enc.Encode(1, "hello", "world"); err != nil {
			b.Fatal(err)
		}
		if err := dec.Decode(&num, &hello, &world); err != nil {
			b.Fatal(err)
		}
	}
}
