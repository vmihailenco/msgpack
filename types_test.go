package msgpack_test

import (
	"bytes"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"gopkg.in/vmihailenco/msgpack.v2"
	"gopkg.in/vmihailenco/msgpack.v2/codes"
)

//------------------------------------------------------------------------------

type intSet map[int]struct{}

var (
	_ msgpack.CustomEncoder = (*intSet)(nil)
	_ msgpack.CustomDecoder = (*intSet)(nil)
)

func (set intSet) EncodeMsgpack(enc *msgpack.Encoder) error {
	slice := make([]int, 0, len(set))
	for n, _ := range set {
		slice = append(slice, n)
	}
	return enc.Encode(slice)
}

func (setptr *intSet) DecodeMsgpack(dec *msgpack.Decoder) error {
	n, err := dec.DecodeSliceLen()
	if err != nil {
		return err
	}

	set := make(intSet, n)
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

type CompactEncodingTest struct {
	str string
	ref *CompactEncodingTest
	num int
}

var (
	_ msgpack.CustomEncoder = (*CompactEncodingTest)(nil)
	_ msgpack.CustomDecoder = (*CompactEncodingTest)(nil)
)

func (s *CompactEncodingTest) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.Encode(s.str, s.ref, s.num)
}

func (s *CompactEncodingTest) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.Decode(&s.str, &s.ref, &s.num)
}

type CompactEncodingFieldTest struct {
	Field CompactEncodingTest
}

//------------------------------------------------------------------------------

type OmitEmptyTest struct {
	Foo string `msgpack:",omitempty"`
	Bar string `msgpack:",omitempty"`
}

type InlineTest struct {
	OmitEmptyTest `msgpack:",inline"`
}

//------------------------------------------------------------------------------

type binTest struct {
	in     interface{}
	wanted []byte
}

var binTests = []binTest{
	{nil, []byte{codes.Nil}},

	{[]byte(nil), []byte{codes.Nil}},
	{[]byte{1, 2, 3}, []byte{codes.Bin8, 0x3, 0x1, 0x2, 0x3}},
	{[3]byte{1, 2, 3}, []byte{codes.Bin8, 0x3, 0x1, 0x2, 0x3}},

	{OmitEmptyTest{}, []byte{codes.FixedMapLow}},

	{intSet{}, []byte{codes.FixedArrayLow}},
	{intSet{8: struct{}{}}, []byte{codes.FixedArrayLow | 1, 0x8}},

	{map[string]string(nil), []byte{codes.Nil}},
	{map[string]string{"a": "", "b": "", "c": "", "d": "", "e": ""}, []byte{
		codes.FixedMapLow | 5,
		codes.FixedStrLow | 1, 'a', codes.FixedStrLow,
		codes.FixedStrLow | 1, 'b', codes.FixedStrLow,
		codes.FixedStrLow | 1, 'c', codes.FixedStrLow,
		codes.FixedStrLow | 1, 'd', codes.FixedStrLow,
		codes.FixedStrLow | 1, 'e', codes.FixedStrLow,
	}},

	{&CompactEncodingTest{}, []byte{codes.FixedStrLow, codes.Nil, 0x0}},
	{
		&CompactEncodingTest{"a", &CompactEncodingTest{"b", nil, 7}, 6},
		[]byte{codes.FixedStrLow | 1, 'a', codes.FixedStrLow | 1, 'b', codes.Nil, 0x7, 0x6},
	},

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
}

func TestBin(t *testing.T) {
	for _, test := range binTests {
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
	stringAlias string
	uint8Alias  uint8
	stringSlice []string
)

type StructTest struct {
	F1 stringSlice
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
	uri, _    = url.Parse("https://github.com/vmihailenco/msgpack")
	typeTests = []typeTest{
		{in: make(chan bool), encErr: "msgpack: Encode(unsupported chan bool)"},

		{in: nil, out: nil, decErr: "msgpack: Decode(nil)"},
		{in: nil, out: 0, decErr: "msgpack: Decode(nonsettable int)"},
		{in: nil, out: (*int)(nil), decErr: "msgpack: Decode(nonsettable *int)"},
		{in: nil, out: new(chan bool), decErr: "msgpack: Decode(unsupported chan bool)"},

		{in: nil, out: new(int), wanted: int(0)},
		{in: nil, out: new(*int), wantnil: true},

		{in: nil, out: new([]int), wantnil: true},
		{in: nil, out: &[]int{1, 2}, wantnil: true},
		{in: make([]int, 0), out: new([]int)},
		{in: make([]int, 1000), out: new([]int)},

		{in: nil, out: new([]byte), wantnil: true},
		{in: []byte(nil), out: new([]byte), wantnil: true},
		{in: []byte(nil), out: &[]byte{}, wantnil: true},
		{in: []byte{1, 2, 3}, out: new([]byte)},

		{in: nil, out: new([3]byte), wanted: [3]byte{}},
		{in: [3]byte{1, 2, 3}, out: new([3]byte)},
		{in: [3]byte{1, 2, 3}, out: new([2]byte), wanted: [2]byte{1, 2}},

		{in: nil, out: new([]interface{}), wantnil: true},
		{in: nil, out: &[]interface{}{}, wantnil: true},
		{in: []interface{}{uint64(1), "hello"}, out: new([]interface{})},

		{in: nil, out: new(map[string]string), wantnil: true},
		{in: nil, out: new(map[int]int), wantnil: true},
		{in: nil, out: &map[string]string{"foo": "bar"}, wantnil: true},
		{in: nil, out: &map[int]int{1: 2}, wantnil: true},
		{in: map[string]interface{}{"foo": nil}, out: new(map[string]interface{})},

		{in: stringSlice{"foo", "bar"}, out: new(stringSlice)},
		{in: []stringAlias{"hello"}, out: new([]stringAlias)},
		{in: []uint8Alias{1}, out: new([]uint8Alias)},

		{in: intSet{}, out: new(intSet)},
		{in: intSet{8: struct{}{}}, out: new(intSet)},

		{in: StructTest{stringSlice{"foo", "bar"}, []string{"hello"}}, out: new(StructTest)},
		{in: TimeEmbedingTest{Time: time.Now()}, out: new(TimeEmbedingTest)},
		{
			in: EmbedingTest{
				unexported: unexported{Foo: "hello"},
				Exported:   Exported{Bar: "world"},
			},
			out: new(EmbedingTest),
		},

		{in: &CompactEncodingTest{}, out: new(CompactEncodingTest)},
		{
			in:  &CompactEncodingTest{"a", &CompactEncodingTest{"b", nil, 1}, 2},
			out: new(CompactEncodingTest),
		},
		{
			in:  &CompactEncodingFieldTest{Field: CompactEncodingTest{"a", nil, 1}},
			out: new(CompactEncodingFieldTest),
		},

		{in: uri, out: new(url.URL)},
	}
)

func indirect(viface interface{}) interface{} {
	v := reflect.Indirect(reflect.ValueOf(viface))
	if v.IsValid() {
		return v.Interface()
	}
	return nil
}

func TestTypes(t *testing.T) {
	for _, test := range typeTests {
		test.T = t

		b, err := msgpack.Marshal(test.in)
		if test.encErr != "" {
			test.assertErr(err, test.encErr)
			continue
		}
		if err != nil {
			t.Fatalf("Marshal failed: %s (in=%#v)", err, test.in)
		}

		err = msgpack.Unmarshal(b, test.out)
		if test.decErr != "" {
			test.assertErr(err, test.decErr)
			continue
		}
		if err != nil {
			t.Fatalf("Unmarshal failed: %s (%s)", err, test)
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
