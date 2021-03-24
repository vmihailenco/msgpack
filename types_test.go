package msgpack_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/vmihailenco/msgpack/v5/msgpcode"
)

//------------------------------------------------------------------------------

type Object struct {
	n int64
}

func (o *Object) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal(o.n)
}

func (o *Object) UnmarshalMsgpack(b []byte) error {
	return msgpack.Unmarshal(b, &o.n)
}

//------------------------------------------------------------------------------

type CustomTime time.Time

func (t CustomTime) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.Encode(time.Time(t))
}

func (t *CustomTime) DecodeMsgpack(dec *msgpack.Decoder) error {
	var tm time.Time
	err := dec.Decode(&tm)
	if err != nil {
		return err
	}
	*t = CustomTime(tm)
	return nil
}

//------------------------------------------------------------------------------

type IntSet map[int]struct{}

var (
	_ msgpack.CustomEncoder = (*IntSet)(nil)
	_ msgpack.CustomDecoder = (*IntSet)(nil)
)

func (set IntSet) EncodeMsgpack(enc *msgpack.Encoder) error {
	slice := make([]int, 0, len(set))
	for n := range set {
		slice = append(slice, n)
	}
	return enc.Encode(slice)
}

func (setptr *IntSet) DecodeMsgpack(dec *msgpack.Decoder) error {
	n, err := dec.DecodeArrayLen()
	if err != nil {
		return err
	}

	set := make(IntSet, n)
	for i := 0; i < n; i++ {
		n, err := dec.DecodeInt()
		if err != nil {
			return err
		}
		set[n] = struct{}{}
	}
	*setptr = set

	return nil
}

//------------------------------------------------------------------------------

type CustomEncoder struct {
	str string
	ref *CustomEncoder
	num int
}

var (
	_ msgpack.CustomEncoder = (*CustomEncoder)(nil)
	_ msgpack.CustomDecoder = (*CustomEncoder)(nil)
)

func (s *CustomEncoder) EncodeMsgpack(enc *msgpack.Encoder) error {
	if s == nil {
		return enc.EncodeNil()
	}
	return enc.EncodeMulti(s.str, s.ref, s.num)
}

func (s *CustomEncoder) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.DecodeMulti(&s.str, &s.ref, &s.num)
}

type CustomEncoderField struct {
	Field CustomEncoder
}

type CustomEncoderEmbeddedPtr struct {
	*CustomEncoder
}

func (s *CustomEncoderEmbeddedPtr) DecodeMsgpack(dec *msgpack.Decoder) error {
	if s.CustomEncoder == nil {
		s.CustomEncoder = new(CustomEncoder)
	}
	return s.CustomEncoder.DecodeMsgpack(dec)
}

//------------------------------------------------------------------------------

type JSONFallbackTest struct {
	Foo string `json:"foo,omitempty"`
	Bar string `json:",omitempty" msgpack:"bar"`
}

func TestUseJsonTag(t *testing.T) {
	var buf bytes.Buffer

	enc := msgpack.NewEncoder(&buf)
	enc.SetCustomStructTag("json")

	in := &JSONFallbackTest{Foo: "hello", Bar: "world"}
	err := enc.Encode(in)
	if err != nil {
		t.Fatal(err)
	}

	dec := msgpack.NewDecoder(&buf)
	dec.SetCustomStructTag("json")

	out := new(JSONFallbackTest)
	err = dec.Decode(out)
	if err != nil {
		t.Fatal(err)
	}

	if out.Foo != in.Foo {
		t.Fatalf("got %q, wanted %q", out.Foo, in.Foo)
	}
	if out.Bar != in.Bar {
		t.Fatalf("got %q, wanted %q", out.Foo, in.Foo)
	}
}

//------------------------------------------------------------------------------

type CustomFallbackTest struct {
	Foo string `custom:"foo,omitempty"`
	Bar string `custom:",omitempty" msgpack:"bar"`
}

func TestUseCustomTag(t *testing.T) {
	var buf bytes.Buffer

	enc := msgpack.NewEncoder(&buf)
	enc.SetCustomStructTag("custom")
	in := &CustomFallbackTest{Foo: "hello", Bar: "world"}
	err := enc.Encode(in)
	if err != nil {
		t.Fatal(err)
	}

	dec := msgpack.NewDecoder(&buf)
	dec.SetCustomStructTag("custom")
	out := new(CustomFallbackTest)
	err = dec.Decode(out)
	if err != nil {
		t.Fatal(err)
	}

	if out.Foo != in.Foo || out.Foo != "hello" {
		t.Fatalf("got %q, wanted %q", out.Foo, in.Foo)
	}
	if out.Bar != in.Bar || out.Bar != "world" {
		t.Fatalf("got %q, wanted %q", out.Foo, in.Foo)
	}
}

//------------------------------------------------------------------------------

type OmitTimeTest struct {
	Foo time.Time  `msgpack:",omitempty"`
	Bar *time.Time `msgpack:",omitempty"`
}

type OmitEmptyTest struct {
	Foo string `msgpack:",omitempty"`
	Bar string `msgpack:",omitempty"`
}

type InlineTest struct {
	OmitEmptyTest
}

type InlinePtrTest struct {
	*OmitEmptyTest
}

type FooTest struct {
	Foo string
}

type FooDupTest FooTest

type InlineDupTest struct {
	FooTest
	FooDupTest
}

type AsArrayTest struct {
	_msgpack struct{} `msgpack:",as_array"`

	OmitEmptyTest
}

type ExtTestField struct {
	ExtTest ExtTest
}

//------------------------------------------------------------------------------

type encoderTest struct {
	in     interface{}
	wanted string
}

var encoderTests = []encoderTest{
	{nil, "c0"},

	{[]byte(nil), "c0"},
	{[]byte{1, 2, 3}, "c403010203"},
	{[3]byte{1, 2, 3}, "c403010203"},

	{time.Unix(0, 0), "d6ff00000000"},
	{time.Unix(1, 1), "d7ff0000000400000001"},
	{time.Time{}, "c70cff00000000fffffff1886e0900"},

	{IntSet{}, "90"},
	{IntSet{8: struct{}{}}, "9108"},

	{map[string]string(nil), "c0"},
	{
		map[string]string{"a": "", "b": "", "c": "", "d": "", "e": ""},
		"85a161a0a162a0a163a0a164a0a165a0",
	},

	{(*Object)(nil), "c0"},
	{&Object{}, "d30000000000000000"},
	{&Object{42}, "d3000000000000002a"},
	{[]*Object{nil, nil}, "92c0c0"},

	{&CustomEncoder{}, "a0c000"},
	{
		&CustomEncoder{"a", &CustomEncoder{"b", nil, 7}, 6},
		"a161a162c00706",
	},

	{OmitEmptyTest{}, "80"},
	{&OmitEmptyTest{Foo: "hello"}, "81a3466f6fa568656c6c6f"},

	{&InlineTest{OmitEmptyTest: OmitEmptyTest{Bar: "world"}}, "81a3426172a5776f726c64"},
	{&InlinePtrTest{OmitEmptyTest: &OmitEmptyTest{Bar: "world"}}, "81a3426172a5776f726c64"},

	{&AsArrayTest{}, "92a0a0"},

	{&JSONFallbackTest{Foo: "hello"}, "82a3666f6fa568656c6c6fa3626172a0"},
	{&JSONFallbackTest{Bar: "world"}, "81a3626172a5776f726c64"},
	{&JSONFallbackTest{Foo: "hello", Bar: "world"}, "82a3666f6fa568656c6c6fa3626172a5776f726c64"},

	{&NoIntern{A: "foo", B: "foo", C: "foo"}, "83a141a3666f6fa142a3666f6fa143a3666f6f"},
	{&Intern{A: "foo", B: "foo", C: "foo"}, "83a141a3666f6fa142d48000a143d48000"},
}

func TestEncoder(t *testing.T) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.SetCustomStructTag("json")
	enc.SetSortMapKeys(true)
	enc.UseCompactInts(true)

	for i, test := range encoderTests {
		buf.Reset()

		err := enc.Encode(test.in)
		require.Nil(t, err)

		s := hex.EncodeToString(buf.Bytes())
		require.Equal(t, test.wanted, s, "#%d", i)
	}
}

type floatEncoderTest struct {
	in      interface{}
	wanted  string
	compact bool
}

var floatEncoderTests = []floatEncoderTest{
	{float32(3.0), "ca40400000", false},
	{float32(3.0), "03", true},

	{float64(3.0), "cb4008000000000000", false},
	{float64(3.0), "03", true},

	{float64(-3.0), "cbc008000000000000", false},
	{float64(-3.0), "fd", true},

	{math.NaN(), "cb7ff8000000000001", false},
	{math.NaN(), "cb7ff8000000000001", true},
	{math.Inf(1), "cb7ff0000000000000", false},
	{math.Inf(1), "cb7ff0000000000000", true},
}

func TestFloatEncoding(t *testing.T) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.UseCompactInts(true)

	for _, test := range floatEncoderTests {
		buf.Reset()

		enc.UseCompactFloats(test.compact)

		err := enc.Encode(test.in)
		if err != nil {
			t.Fatal(err)
		}

		s := hex.EncodeToString(buf.Bytes())
		if s != test.wanted {
			t.Fatalf("%s != %s (in=%#v)", s, test.wanted, test.in)
		}
	}
}

//------------------------------------------------------------------------------

type decoderTest struct {
	b   []byte
	out interface{}
	err string
}

var decoderTests = []decoderTest{
	{b: []byte{byte(msgpcode.Bin32), 0x0f, 0xff, 0xff, 0xff}, out: new([]byte), err: "EOF"},
	{b: []byte{byte(msgpcode.Str32), 0x0f, 0xff, 0xff, 0xff}, out: new([]byte), err: "EOF"},
	{b: []byte{byte(msgpcode.Array32), 0x0f, 0xff, 0xff, 0xff}, out: new([]int), err: "EOF"},
	{b: []byte{byte(msgpcode.Map32), 0x0f, 0xff, 0xff, 0xff}, out: new(map[int]int), err: "EOF"},
}

func TestDecoder(t *testing.T) {
	for i, test := range decoderTests {
		err := msgpack.Unmarshal(test.b, test.out)
		if err == nil {
			t.Fatalf("#%d err is nil, wanted %q", i, test.err)
		}
		if err.Error() != test.err {
			t.Fatalf("#%d err is %q, wanted %q", i, err.Error(), test.err)
		}
	}
}

//------------------------------------------------------------------------------

type unexported struct {
	Foo string
}

type Exported struct {
	Bar string
}

type EmbeddingTest struct {
	unexported
	Exported
}

type EmbeddingPtrTest struct {
	*Exported
}

type EmbeddedTime struct {
	time.Time
}

type (
	interfaceAlias     interface{}
	byteAlias          byte
	uint8Alias         uint8
	stringAlias        string
	sliceByte          []byte
	sliceString        []string
	mapStringString    map[string]string
	mapStringInterface map[string]interface{}
)

type StructTest struct {
	F1 sliceString
	F2 []string
}

type typeTest struct {
	*testing.T

	in       interface{}
	out      interface{}
	encErr   string
	decErr   string
	wantnil  bool
	wantzero bool
	wanted   interface{}
}

func (t typeTest) String() string {
	return fmt.Sprintf("in=%#v, out=%#v", t.in, t.out)
}

func (t *typeTest) requireErr(err error, s string) {
	if err == nil {
		t.Fatalf("got %v error, wanted %q", err, s)
	}
	if err.Error() != s {
		t.Fatalf("got %q error, wanted %q", err, s)
	}
}

var (
	intSlice   = make([]int, 0, 3)
	repoURL, _ = url.Parse("https://github.com/vmihailenco/msgpack")
	typeTests  = []typeTest{
		{in: make(chan bool), encErr: "msgpack: Encode(unsupported chan bool)"},

		{in: nil, out: nil, decErr: "msgpack: Decode(nil)"},
		{in: nil, out: 0, decErr: "msgpack: Decode(non-pointer int)"},
		{in: nil, out: (*int)(nil), decErr: "msgpack: Decode(non-settable *int)"},
		{in: nil, out: new(chan bool), decErr: "msgpack: Decode(unsupported chan bool)"},

		{in: true, out: new(bool)},
		{in: false, out: new(bool)},

		{in: nil, out: new(int), wanted: int(0)},
		{in: nil, out: new(*int), wantnil: true},

		{in: float32(3.14), out: new(float32)},
		{in: int8(-1), out: new(float32), wanted: float32(-1)},
		{in: int32(1), out: new(float32), wanted: float32(1)},
		{in: int32(999999999), out: new(float32), wanted: float32(999999999)},
		{in: int64(math.MaxInt64), out: new(float32), wanted: float32(math.MaxInt64)},

		{in: float64(3.14), out: new(float64)},
		{in: int8(-1), out: new(float64), wanted: float64(-1)},
		{in: int64(1), out: new(float64), wanted: float64(1)},
		{in: int64(999999999), out: new(float64), wanted: float64(999999999)},
		{in: int64(math.MaxInt64), out: new(float64), wanted: float64(math.MaxInt64)},

		{in: nil, out: new(*string), wantnil: true},
		{in: nil, out: new(string), wanted: ""},
		{in: "", out: new(string)},
		{in: "foo", out: new(string)},

		{in: nil, out: new([]byte), wantnil: true},
		{in: []byte(nil), out: new([]byte), wantnil: true},
		{in: []byte(nil), out: &[]byte{}, wantnil: true},
		{in: []byte{1, 2, 3}, out: new([]byte)},
		{in: []byte{1, 2, 3}, out: new([]byte)},
		{in: sliceByte{1, 2, 3}, out: new(sliceByte)},
		{in: []byteAlias{1, 2, 3}, out: new([]byteAlias)},
		{in: []uint8Alias{1, 2, 3}, out: new([]uint8Alias)},

		{in: nil, out: new([3]byte), wanted: [3]byte{}},
		{in: [3]byte{1, 2, 3}, out: new([3]byte)},
		{in: [3]byte{1, 2, 3}, out: new([2]byte), decErr: "[2]uint8 len is 2, but msgpack has 3 elements"},

		{in: nil, out: new([]interface{}), wantnil: true},
		{in: nil, out: new([]interface{}), wantnil: true},
		{in: []interface{}{int8(1), "hello"}, out: new([]interface{})},

		{in: nil, out: new([]int), wantnil: true},
		{in: nil, out: &[]int{1, 2}, wantnil: true},
		{in: []int(nil), out: new([]int), wantnil: true},
		{in: make([]int, 0), out: new([]int)},
		{in: []int{}, out: new([]int)},
		{in: []int{1, 2, 3}, out: new([]int)},
		{in: []int{1, 2, 3}, out: &intSlice},
		{in: [3]int{1, 2, 3}, out: new([3]int)},
		{in: [3]int{1, 2, 3}, out: new([2]int), decErr: "[2]int len is 2, but msgpack has 3 elements"},

		{in: []string(nil), out: new([]string), wantnil: true},
		{in: []string{}, out: new([]string)},
		{in: []string{"a", "b"}, out: new([]string)},
		{in: [2]string{"a", "b"}, out: new([2]string)},
		{in: sliceString{"foo", "bar"}, out: new(sliceString)},
		{in: []stringAlias{"hello"}, out: new([]stringAlias)},

		{in: nil, out: new(map[string]string), wantnil: true},
		{in: nil, out: new(map[int]int), wantnil: true},
		{in: nil, out: &map[string]string{"foo": "bar"}, wantnil: true},
		{in: nil, out: &map[int]int{1: 2}, wantnil: true},
		{in: map[string]string(nil), out: new(map[string]string)},
		{in: map[string]interface{}{"foo": nil}, out: new(map[string]interface{})},
		{in: mapStringString{"foo": "bar"}, out: new(mapStringString)},
		{in: map[stringAlias]stringAlias{"foo": "bar"}, out: new(map[stringAlias]stringAlias)},
		{in: mapStringInterface{"foo": "bar"}, out: new(mapStringInterface)},
		{in: map[stringAlias]interfaceAlias{"foo": "bar"}, out: new(map[stringAlias]interfaceAlias)},
		{in: map[int]string{1: "string"}, out: new(map[int]string)},

		{in: (*Object)(nil), out: new(*Object)},
		{in: &Object{42}, out: new(Object)},
		{in: []*Object{new(Object), new(Object)}, out: new([]*Object)},

		{in: IntSet{}, out: new(IntSet)},
		{in: IntSet{42: struct{}{}}, out: new(IntSet)},
		{in: IntSet{42: struct{}{}}, out: new(*IntSet)},

		{in: StructTest{sliceString{"foo", "bar"}, []string{"hello"}}, out: new(StructTest)},
		{in: StructTest{sliceString{"foo", "bar"}, []string{"hello"}}, out: new(*StructTest)},

		{in: EmbeddingTest{}, out: new(EmbeddingTest)},
		{
			in:     EmbeddingTest{},
			out:    new(EmbeddingPtrTest),
			wanted: EmbeddingPtrTest{Exported: new(Exported)},
		},
		{
			in:     EmbeddingPtrTest{},
			out:    new(EmbeddingPtrTest),
			wanted: EmbeddingPtrTest{Exported: new(Exported)},
		},
		{in: EmbeddingTest{}, out: new(*EmbeddingTest)},
		{
			in: EmbeddingTest{
				unexported: unexported{Foo: "hello"},
				Exported:   Exported{Bar: "world"},
			},
			out: new(EmbeddingTest),
		},

		{in: time.Unix(0, 0), out: new(time.Time)},
		{in: new(time.Time), out: new(time.Time)},
		{in: time.Unix(0, 1), out: new(time.Time)},
		{in: time.Unix(1, 0), out: new(time.Time)},
		{in: time.Unix(1, 1), out: new(time.Time)},
		{
			in:     time.Unix(0, 0).Format(time.RFC3339),
			out:    new(time.Time),
			wanted: mustParseTime(time.RFC3339, time.Unix(0, 0).Format(time.RFC3339)),
		},
		{in: EmbeddedTime{Time: time.Unix(1, 1)}, out: new(EmbeddedTime)},
		{in: EmbeddedTime{Time: time.Unix(1, 1)}, out: new(*EmbeddedTime)},
		{in: CustomTime(time.Unix(0, 0)), out: new(CustomTime)},

		{in: nil, out: new(*CustomEncoder), wantnil: true},
		{in: nil, out: &CustomEncoder{str: "a"}, wantzero: true},
		{
			in:  &CustomEncoder{"a", &CustomEncoder{"b", nil, 1}, 2},
			out: new(CustomEncoder),
		},
		{
			in:  &CustomEncoderField{Field: CustomEncoder{"a", nil, 1}},
			out: new(CustomEncoderField),
		},
		{
			in:  &CustomEncoderEmbeddedPtr{&CustomEncoder{"a", nil, 1}},
			out: new(CustomEncoderEmbeddedPtr),
		},

		{in: repoURL, out: new(url.URL)},
		{in: repoURL, out: new(*url.URL)},

		{in: OmitEmptyTest{}, out: new(OmitEmptyTest)},
		{in: OmitTimeTest{}, out: new(OmitTimeTest)},

		{in: nil, out: new(*AsArrayTest), wantnil: true},
		{in: nil, out: new(AsArrayTest), wantzero: true},
		{in: AsArrayTest{OmitEmptyTest: OmitEmptyTest{"foo", "bar"}}, out: new(AsArrayTest)},
		{
			in:     AsArrayTest{OmitEmptyTest: OmitEmptyTest{"foo", "bar"}},
			out:    new(unexported),
			decErr: "msgpack: number of fields in array-encoded struct has changed",
		},

		{in: (*EventTime)(nil), out: new(*EventTime)},
		{in: &EventTime{time.Unix(0, 0)}, out: new(*EventTime)},

		{in: (*ExtTest)(nil), out: new(*ExtTest)},
		{in: &ExtTest{"world"}, out: new(*ExtTest), wanted: ExtTest{"hello world"}},
		{
			in:     &ExtTestField{ExtTest{"world"}},
			out:    new(*ExtTestField),
			wanted: ExtTestField{ExtTest{"hello world"}},
		},

		{
			in:  &InlineTest{OmitEmptyTest: OmitEmptyTest{Bar: "world"}},
			out: new(InlineTest),
		},
		{
			in:  &InlinePtrTest{OmitEmptyTest: &OmitEmptyTest{Bar: "world"}},
			out: new(InlinePtrTest),
		},
		{
			in:  InlineDupTest{FooTest{"foo"}, FooDupTest{"foo"}},
			out: new(InlineDupTest),
		},

		{in: big.NewInt(123), out: new(big.Int)},
	}
)

func indirect(viface interface{}) interface{} {
	v := reflect.ValueOf(viface)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.IsValid() {
		return v.Interface()
	}
	return nil
}

func TestTypes(t *testing.T) {
	msgpack.RegisterExt(1, (*EventTime)(nil))

	for _, test := range typeTests {
		test.T = t

		var buf bytes.Buffer

		enc := msgpack.NewEncoder(&buf)
		err := enc.Encode(test.in)
		if test.encErr != "" {
			test.requireErr(err, test.encErr)
			continue
		}
		if err != nil {
			t.Fatalf("Encode failed: %s (in=%#v)", err, test.in)
		}

		dec := msgpack.NewDecoder(&buf)
		err = dec.Decode(test.out)
		if test.decErr != "" {
			test.requireErr(err, test.decErr)
			continue
		}
		if err != nil {
			t.Fatalf("Decode failed: %s (%s)", err, test)
		}

		if buf.Len() > 0 {
			t.Fatalf("unread data in the buffer: %q (%s)", buf.Bytes(), test)
		}

		if test.wantnil {
			v := reflect.Indirect(reflect.ValueOf(test.out))
			if !v.IsNil() {
				t.Fatalf("got %#v, wanted nil (%s)", test.out, test)
			}
			continue
		}

		out := indirect(test.out)
		var wanted interface{}
		if test.wantzero {
			typ := reflect.TypeOf(out)
			wanted = reflect.Zero(typ).Interface()
		} else {
			wanted = test.wanted
		}
		if wanted == nil {
			wanted = indirect(test.in)
		}
		require.Equal(t, wanted, out)
	}

	for _, test := range typeTests {
		if test.encErr != "" || test.decErr != "" {
			continue
		}

		b, err := msgpack.Marshal(test.in)
		if err != nil {
			t.Fatal(err)
		}

		var dst interface{}
		dec := msgpack.NewDecoder(bytes.NewReader(b))
		dec.SetMapDecoder(func(dec *msgpack.Decoder) (interface{}, error) {
			return dec.DecodeUntypedMap()
		})

		err = dec.Decode(&dst)
		if err != nil {
			t.Fatalf("Unmarshal into interface{} failed: %s (%s)", err, test)
		}

		dec = msgpack.NewDecoder(bytes.NewReader(b))
		dec.SetMapDecoder(func(dec *msgpack.Decoder) (interface{}, error) {
			return dec.DecodeUntypedMap()
		})

		_, err = dec.DecodeInterface()
		if err != nil {
			t.Fatalf("DecodeInterface failed: %s (%s)", err, test)
		}
	}
}

func TestStringsBin(t *testing.T) {
	tests := []struct {
		in     string
		wanted string
	}{
		{"", "a0"},
		{"a", "a161"},
		{"hello", "a568656c6c6f"},
		{
			strings.Repeat("x", 31),
			"bf" + strings.Repeat("78", 31),
		},
		{
			strings.Repeat("x", 32),
			"d920" + strings.Repeat("78", 32),
		},
		{
			strings.Repeat("x", 255),
			"d9ff" + strings.Repeat("78", 255),
		},
		{
			strings.Repeat("x", 256),
			"da0100" + strings.Repeat("78", 256),
		},
		{
			strings.Repeat("x", 65535),
			"daffff" + strings.Repeat("78", 65535),
		},
		{
			strings.Repeat("x", 65536),
			"db00010000" + strings.Repeat("78", 65536),
		},
	}

	for _, test := range tests {
		b, err := msgpack.Marshal(test.in)
		require.Nil(t, err)
		s := hex.EncodeToString(b)
		require.Equal(t, s, test.wanted)

		var out string
		err = msgpack.Unmarshal(b, &out)
		require.Nil(t, err)
		require.Equal(t, out, test.in)

		var msg msgpack.RawMessage
		err = msgpack.Unmarshal(b, &msg)
		require.Nil(t, err)
		require.Equal(t, []byte(msg), b)

		dec := msgpack.NewDecoder(bytes.NewReader(b))
		v, err := dec.DecodeInterface()
		require.Nil(t, err)
		require.Equal(t, v.(string), test.in)

		var dst interface{}
		dst = ""
		err = msgpack.Unmarshal(b, &dst)
		require.EqualError(t, err, "msgpack: Decode(non-pointer string)")
	}
}

func TestBin(t *testing.T) {
	tests := []struct {
		in     []byte
		wanted string
	}{
		{[]byte{}, "c400"},
		{[]byte{0}, "c40100"},
		{
			bytes.Repeat([]byte{'x'}, 31),
			"c41f" + strings.Repeat("78", 31),
		},
		{
			bytes.Repeat([]byte{'x'}, 32),
			"c420" + strings.Repeat("78", 32),
		},
		{
			bytes.Repeat([]byte{'x'}, 255),
			"c4ff" + strings.Repeat("78", 255),
		},
		{
			bytes.Repeat([]byte{'x'}, 256),
			"c50100" + strings.Repeat("78", 256),
		},
		{
			bytes.Repeat([]byte{'x'}, 65535),
			"c5ffff" + strings.Repeat("78", 65535),
		},
		{
			bytes.Repeat([]byte{'x'}, 65536),
			"c600010000" + strings.Repeat("78", 65536),
		},
	}

	for _, test := range tests {
		b, err := msgpack.Marshal(test.in)
		if err != nil {
			t.Fatal(err)
		}
		s := hex.EncodeToString(b)
		if s != test.wanted {
			t.Fatalf("%.32s != %.32s", s, test.wanted)
		}

		var out []byte
		err = msgpack.Unmarshal(b, &out)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(out, test.in) {
			t.Fatalf("%x != %x", out, test.in)
		}

		dec := msgpack.NewDecoder(bytes.NewReader(b))
		v, err := dec.DecodeInterface()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(v.([]byte), test.in) {
			t.Fatalf("%x != %x", v, test.in)
		}

		var dst interface{}
		dst = make([]byte, 0)
		err = msgpack.Unmarshal(b, &dst)
		if err.Error() != "msgpack: Decode(non-pointer []uint8)" {
			t.Fatal(err)
		}
	}
}

func TestUint64(t *testing.T) {
	tests := []struct {
		in     uint64
		wanted string
	}{
		{0, "00"},
		{1, "01"},
		{math.MaxInt8 - 1, "7e"},
		{math.MaxInt8, "7f"},
		{math.MaxInt8 + 1, "cc80"},
		{math.MaxUint8 - 1, "ccfe"},
		{math.MaxUint8, "ccff"},
		{math.MaxUint8 + 1, "cd0100"},
		{math.MaxUint16 - 1, "cdfffe"},
		{math.MaxUint16, "cdffff"},
		{math.MaxUint16 + 1, "ce00010000"},
		{math.MaxUint32 - 1, "cefffffffe"},
		{math.MaxUint32, "ceffffffff"},
		{math.MaxUint32 + 1, "cf0000000100000000"},
		{math.MaxInt64 - 1, "cf7ffffffffffffffe"},
		{math.MaxInt64, "cf7fffffffffffffff"},
	}

	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.UseCompactInts(true)

	for _, test := range tests {
		err := enc.Encode(test.in)
		if err != nil {
			t.Fatal(err)
		}
		s := hex.EncodeToString(buf.Bytes())
		if s != test.wanted {
			t.Fatalf("%.32s != %.32s", s, test.wanted)
		}

		var out uint64
		err = msgpack.Unmarshal(buf.Bytes(), &out)
		if err != nil {
			t.Fatal(err)
		}
		if out != test.in {
			t.Fatalf("%d != %d", out, test.in)
		}

		var out2 int64
		err = msgpack.Unmarshal(buf.Bytes(), &out2)
		if err != nil {
			t.Fatal(err)
		}
		if out2 != int64(test.in) {
			t.Fatalf("%d != %d", out2, int64(test.in))
		}

		var out3 interface{}
		out3 = uint64(0)
		err = msgpack.Unmarshal(buf.Bytes(), &out3)
		if err.Error() != "msgpack: Decode(non-pointer uint64)" {
			t.Fatal(err)
		}

		dec := msgpack.NewDecoder(&buf)
		_, err = dec.DecodeInterface()
		if err != nil {
			t.Fatal(err)
		}

		if buf.Len() != 0 {
			panic("buffer is not empty")
		}
	}
}

func TestInt64(t *testing.T) {
	tests := []struct {
		in     int64
		wanted string
	}{
		{math.MinInt64, "d38000000000000000"},
		{math.MinInt32 - 1, "d3ffffffff7fffffff"},
		{math.MinInt32, "d280000000"},
		{math.MinInt32 + 1, "d280000001"},
		{math.MinInt16 - 1, "d2ffff7fff"},
		{math.MinInt16, "d18000"},
		{math.MinInt16 + 1, "d18001"},
		{math.MinInt8 - 1, "d1ff7f"},
		{math.MinInt8, "d080"},
		{math.MinInt8 + 1, "d081"},
		{-33, "d0df"},
		{-32, "e0"},
		{-31, "e1"},
		{-1, "ff"},
		{0, "00"},
		{1, "01"},
		{math.MaxInt8 - 1, "7e"},
		{math.MaxInt8, "7f"},
		{math.MaxInt8 + 1, "cc80"},
		{math.MaxUint8 - 1, "ccfe"},
		{math.MaxUint8, "ccff"},
		{math.MaxUint8 + 1, "cd0100"},
		{math.MaxUint16 - 1, "cdfffe"},
		{math.MaxUint16, "cdffff"},
		{math.MaxUint16 + 1, "ce00010000"},
		{math.MaxUint32 - 1, "cefffffffe"},
		{math.MaxUint32, "ceffffffff"},
		{math.MaxUint32 + 1, "cf0000000100000000"},
		{math.MaxInt64 - 1, "cf7ffffffffffffffe"},
		{math.MaxInt64, "cf7fffffffffffffff"},
	}

	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.UseCompactInts(true)

	for _, test := range tests {
		err := enc.Encode(test.in)
		if err != nil {
			t.Fatal(err)
		}
		s := hex.EncodeToString(buf.Bytes())
		if s != test.wanted {
			t.Fatalf("%.32s != %.32s", s, test.wanted)
		}

		var out int64
		err = msgpack.Unmarshal(buf.Bytes(), &out)
		if err != nil {
			t.Fatal(err)
		}
		if out != test.in {
			t.Fatalf("%d != %d", out, test.in)
		}

		var out2 uint64
		err = msgpack.Unmarshal(buf.Bytes(), &out2)
		if err != nil {
			t.Fatal(err)
		}
		if out2 != uint64(test.in) {
			t.Fatalf("%d != %d", out2, uint64(test.in))
		}

		var out3 interface{}
		out3 = int64(0)
		err = msgpack.Unmarshal(buf.Bytes(), &out3)
		if err.Error() != "msgpack: Decode(non-pointer int64)" {
			t.Fatal(err)
		}

		dec := msgpack.NewDecoder(&buf)
		_, err = dec.DecodeInterface()
		if err != nil {
			t.Fatal(err)
		}

		if buf.Len() != 0 {
			panic("buffer is not empty")
		}
	}
}

func TestFloat32(t *testing.T) {
	tests := []struct {
		in     float32
		wanted string
	}{
		{0.1, "ca3dcccccd"},
		{0.2, "ca3e4ccccd"},
		{-0.1, "cabdcccccd"},
		{-0.2, "cabe4ccccd"},
		{float32(math.Inf(1)), "ca7f800000"},
		{float32(math.Inf(-1)), "caff800000"},
		{math.MaxFloat32, "ca7f7fffff"},
		{math.SmallestNonzeroFloat32, "ca00000001"},
	}
	for _, test := range tests {
		b, err := msgpack.Marshal(test.in)
		if err != nil {
			t.Fatal(err)
		}
		s := hex.EncodeToString(b)
		if s != test.wanted {
			t.Fatalf("%.32s != %.32s", s, test.wanted)
		}

		var out float32
		err = msgpack.Unmarshal(b, &out)
		if err != nil {
			t.Fatal(err)
		}
		if out != test.in {
			t.Fatalf("%f != %f", out, test.in)
		}

		var out2 float64
		err = msgpack.Unmarshal(b, &out2)
		if err != nil {
			t.Fatal(err)
		}
		if out2 != float64(test.in) {
			t.Fatalf("%f != %f", out2, float64(test.in))
		}

		dec := msgpack.NewDecoder(bytes.NewReader(b))
		v, err := dec.DecodeInterface()
		if err != nil {
			t.Fatal(err)
		}
		if v.(float32) != test.in {
			t.Fatalf("%f != %f", v, test.in)
		}

		var dst interface{}
		dst = float32(0)
		err = msgpack.Unmarshal(b, &dst)
		if err.Error() != "msgpack: Decode(non-pointer float32)" {
			t.Fatal(err)
		}
	}

	in := float32(math.NaN())
	b, err := msgpack.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	var out float32
	err = msgpack.Unmarshal(b, &out)
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsNaN(float64(out)) {
		t.Fatal("not NaN")
	}
}

func TestFloat64(t *testing.T) {
	table := []struct {
		in     float64
		wanted string
	}{
		{0.1, "cb3fb999999999999a"},
		{0.2, "cb3fc999999999999a"},
		{-0.1, "cbbfb999999999999a"},
		{-0.2, "cbbfc999999999999a"},
		{math.Inf(1), "cb7ff0000000000000"},
		{math.Inf(-1), "cbfff0000000000000"},
		{math.MaxFloat64, "cb7fefffffffffffff"},
		{math.SmallestNonzeroFloat64, "cb0000000000000001"},
	}
	for _, test := range table {
		b, err := msgpack.Marshal(test.in)
		if err != nil {
			t.Fatal(err)
		}
		s := hex.EncodeToString(b)
		if s != test.wanted {
			t.Fatalf("%.32s != %.32s", s, test.wanted)
		}

		var out float64
		err = msgpack.Unmarshal(b, &out)
		if err != nil {
			t.Fatal(err)
		}
		if out != test.in {
			t.Fatalf("%f != %f", out, test.in)
		}

		dec := msgpack.NewDecoder(bytes.NewReader(b))
		v, err := dec.DecodeInterface()
		if err != nil {
			t.Fatal(err)
		}
		if v.(float64) != test.in {
			t.Fatalf("%f != %f", v, test.in)
		}

		var dst interface{}
		dst = float64(0)
		err = msgpack.Unmarshal(b, &dst)
		if err.Error() != "msgpack: Decode(non-pointer float64)" {
			t.Fatal(err)
		}
	}
}

func mustParseTime(format, s string) time.Time {
	tm, err := time.Parse(format, s)
	if err != nil {
		panic(err)
	}
	return tm
}
