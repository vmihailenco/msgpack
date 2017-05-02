package msgpack_test

import (
	"bytes"
	"fmt"
	"math"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"gopkg.in/vmihailenco/msgpack.v2"
	"gopkg.in/vmihailenco/msgpack.v2/codes"
)

//------------------------------------------------------------------------------

type Object struct {
	n int
}

func (o *Object) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal(o.n)
}

func (o *Object) UnmarshalMsgpack(b []byte) error {
	return msgpack.Unmarshal(b, &o.n)
}

//------------------------------------------------------------------------------

type IntSet map[int]struct{}

var _ msgpack.CustomEncoder = (*IntSet)(nil)
var _ msgpack.CustomDecoder = (*IntSet)(nil)

func (set IntSet) EncodeMsgpack(enc *msgpack.Encoder) error {
	slice := make([]int, 0, len(set))
	for n, _ := range set {
		slice = append(slice, n)
	}
	return enc.Encode(slice)
}

func (setptr *IntSet) DecodeMsgpack(dec *msgpack.Decoder) error {
	n, err := dec.DecodeSliceLen()
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

var _ msgpack.CustomEncoder = (*CustomEncoder)(nil)
var _ msgpack.CustomDecoder = (*CustomEncoder)(nil)

func (s *CustomEncoder) EncodeMsgpack(enc *msgpack.Encoder) error {
	if s == nil {
		return enc.EncodeNil()
	}
	return enc.Encode(s.str, s.ref, s.num)
}

func (s *CustomEncoder) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.Decode(&s.str, &s.ref, &s.num)
}

type CustomEncoderField struct {
	Field CustomEncoder
}

//------------------------------------------------------------------------------

type OmitEmptyTest struct {
	Foo string `msgpack:",omitempty"`
	Bar string `msgpack:",omitempty"`
}

type InlineTest struct {
	OmitEmptyTest `msgpack:",inline"`
}

type AsArrayTest struct {
	_msgpack struct{} `msgpack:",asArray"`

	OmitEmptyTest `msgpack:",inline"`
}

//------------------------------------------------------------------------------

type encoderTest struct {
	in     interface{}
	wanted []byte
}

var encoderTests = []encoderTest{
	{nil, []byte{codes.Nil}},

	{[]byte(nil), []byte{codes.Nil}},
	{[]byte{1, 2, 3}, []byte{codes.Bin8, 0x3, 0x1, 0x2, 0x3}},
	{[3]byte{1, 2, 3}, []byte{codes.Bin8, 0x3, 0x1, 0x2, 0x3}},

	{IntSet{}, []byte{codes.FixedArrayLow}},
	{IntSet{8: struct{}{}}, []byte{codes.FixedArrayLow | 1, 0x8}},

	{map[string]string(nil), []byte{codes.Nil}},
	{map[string]string{"a": "", "b": "", "c": "", "d": "", "e": ""}, []byte{
		codes.FixedMapLow | 5,
		codes.FixedStrLow | 1, 'a', codes.FixedStrLow,
		codes.FixedStrLow | 1, 'b', codes.FixedStrLow,
		codes.FixedStrLow | 1, 'c', codes.FixedStrLow,
		codes.FixedStrLow | 1, 'd', codes.FixedStrLow,
		codes.FixedStrLow | 1, 'e', codes.FixedStrLow,
	}},

	{(*Object)(nil), []byte{codes.Nil}},
	{&Object{}, []byte{0}},
	{&Object{42}, []byte{42}},
	{[]*Object{nil, nil}, []byte{codes.FixedArrayLow | 2, codes.Nil, codes.Nil}},

	{&CustomEncoder{}, []byte{codes.FixedStrLow, codes.Nil, 0x0}},
	{
		&CustomEncoder{"a", &CustomEncoder{"b", nil, 7}, 6},
		[]byte{codes.FixedStrLow | 1, 'a', codes.FixedStrLow | 1, 'b', codes.Nil, 0x7, 0x6},
	},

	{OmitEmptyTest{}, []byte{codes.FixedMapLow}},
	{&OmitEmptyTest{Foo: "hello"}, []byte{
		codes.FixedMapLow | 1,
		codes.FixedStrLow | byte(len("Foo")), 'F', 'o', 'o',
		codes.FixedStrLow | byte(len("hello")), 'h', 'e', 'l', 'l', 'o',
	}},

	{&InlineTest{OmitEmptyTest: OmitEmptyTest{Bar: "world"}}, []byte{
		codes.FixedMapLow | 1,
		codes.FixedStrLow | byte(len("Bar")), 'B', 'a', 'r',
		codes.FixedStrLow | byte(len("world")), 'w', 'o', 'r', 'l', 'd',
	}},

	{&AsArrayTest{}, []byte{codes.FixedArrayLow | 2, codes.FixedStrLow, codes.FixedStrLow}},
}

func TestEncoder(t *testing.T) {
	for _, test := range encoderTests {
		var buf bytes.Buffer
		enc := msgpack.NewEncoder(&buf).SortMapKeys(true)
		if err := enc.Encode(test.in); err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(buf.Bytes(), test.wanted) {
			t.Fatalf("%q != %q (in=%#v)", buf.Bytes(), test.wanted, test.in)
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
	{b: []byte{codes.Bin32, 0x0f, 0xff, 0xff, 0xff}, out: new([]byte), err: "EOF"},
	{b: []byte{codes.Str32, 0x0f, 0xff, 0xff, 0xff}, out: new([]byte), err: "EOF"},
	{b: []byte{codes.Array32, 0x0f, 0xff, 0xff, 0xff}, out: new([]int), err: "EOF"},
	{b: []byte{codes.Map32, 0x0f, 0xff, 0xff, 0xff}, out: new(map[int]int), err: "EOF"},
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

type EmbedingTest struct {
	unexported
	Exported
}

//------------------------------------------------------------------------------

type TimeEmbedingTest struct {
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

	in      interface{}
	out     interface{}
	encErr  string
	decErr  string
	wantnil bool
	wanted  interface{}
}

func (t typeTest) String() string {
	return fmt.Sprintf("in=%#v, out=%#v", t.in, t.out)
}

func (t *typeTest) assertErr(err error, s string) {
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
		{in: nil, out: 0, decErr: "msgpack: Decode(nonsettable int)"},
		{in: nil, out: (*int)(nil), decErr: "msgpack: Decode(nonsettable *int)"},
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
		{in: []interface{}{uint64(1), "hello"}, out: new([]interface{})},

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
		{in: map[string]interface{}{"foo": nil}, out: new(map[string]interface{})},
		{in: mapStringString{"foo": "bar"}, out: new(mapStringString)},
		{in: map[stringAlias]stringAlias{"foo": "bar"}, out: new(map[stringAlias]stringAlias)},
		{in: mapStringInterface{"foo": "bar"}, out: new(mapStringInterface)},
		{in: map[stringAlias]interfaceAlias{"foo": "bar"}, out: new(map[stringAlias]interfaceAlias)},

		{in: (*Object)(nil), out: new(*Object)},
		{in: &Object{42}, out: new(Object)},
		{in: []*Object{new(Object), new(Object)}, out: new([]*Object)},

		{in: IntSet{}, out: new(IntSet)},
		{in: IntSet{42: struct{}{}}, out: new(IntSet)},
		{in: IntSet{42: struct{}{}}, out: new(*IntSet)},

		{in: StructTest{sliceString{"foo", "bar"}, []string{"hello"}}, out: new(StructTest)},
		{in: StructTest{sliceString{"foo", "bar"}, []string{"hello"}}, out: new(*StructTest)},

		{in: EmbedingTest{}, out: new(EmbedingTest)},
		{in: EmbedingTest{}, out: new(*EmbedingTest)},
		{
			in: EmbedingTest{
				unexported: unexported{Foo: "hello"},
				Exported:   Exported{Bar: "world"},
			},
			out: new(EmbedingTest),
		},

		{in: TimeEmbedingTest{Time: time.Now()}, out: new(TimeEmbedingTest)},
		{in: TimeEmbedingTest{Time: time.Now()}, out: new(*TimeEmbedingTest)},

		{in: new(CustomEncoder), out: new(CustomEncoder)},
		{in: new(CustomEncoder), out: new(*CustomEncoder)},
		{
			in:  &CustomEncoder{"a", &CustomEncoder{"b", nil, 1}, 2},
			out: new(CustomEncoder),
		},
		{
			in:  &CustomEncoderField{Field: CustomEncoder{"a", nil, 1}},
			out: new(CustomEncoderField),
		},

		{in: repoURL, out: new(url.URL)},
		{in: repoURL, out: new(*url.URL)},

		{in: AsArrayTest{}, out: new(AsArrayTest)},
		{in: AsArrayTest{}, out: new(*AsArrayTest)},
		{in: AsArrayTest{OmitEmptyTest: OmitEmptyTest{"foo", "bar"}}, out: new(AsArrayTest)},
		{
			in:     AsArrayTest{OmitEmptyTest: OmitEmptyTest{"foo", "bar"}},
			out:    new(unexported),
			wanted: unexported{Foo: "foo"},
		},

		{in: (*EventTime)(nil), out: new(*EventTime)},
		{in: &EventTime{time.Unix(0, 0)}, out: new(EventTime)},

		{in: (*ExtTest)(nil), out: new(*ExtTest)},
		{in: &ExtTest{"world"}, out: new(ExtTest), wanted: ExtTest{"hello world"}},
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
	for _, test := range typeTests {
		test.T = t

		var buf bytes.Buffer

		enc := msgpack.NewEncoder(&buf)
		err := enc.Encode(test.in)
		if test.encErr != "" {
			test.assertErr(err, test.encErr)
			continue
		}
		if err != nil {
			t.Fatalf("Marshal failed: %s (in=%#v)", err, test.in)
		}

		dec := msgpack.NewDecoder(&buf)
		err = dec.Decode(test.out)
		if test.decErr != "" {
			test.assertErr(err, test.decErr)
			continue
		}
		if err != nil {
			t.Fatalf("Unmarshal failed: %s (%s)", err, test)
		}

		if buf.Len() > 0 {
			t.Fatalf("unread data in the buffer: %q (%s)", buf.Bytes(), test)
		}

		if test.wantnil {
			v := reflect.Indirect(reflect.ValueOf(test.out))
			if v.IsNil() {
				continue
			}
			t.Fatalf("got %#v, wanted nil (%s)", test.out, test)
		}

		out := indirect(test.out)
		wanted := test.wanted
		if wanted == nil {
			wanted = indirect(test.in)
		}
		if !reflect.DeepEqual(out, wanted) {
			t.Fatalf("%#v != %#v (%s)", out, wanted, test)
		}
	}
}

func TestStrings(t *testing.T) {
	for _, n := range []int{0, 1, 31, 32, 255, 256, 65535, 65536} {
		in := strings.Repeat("x", n)
		b, err := msgpack.Marshal(in)
		if err != nil {
			t.Fatal(err)
		}

		var out string
		err = msgpack.Unmarshal(b, &out)
		if err != nil {
			t.Fatal(err)
		}

		if out != in {
			t.Fatalf("%q != %q", out, in)
		}
	}
}
