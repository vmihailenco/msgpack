package msgpack_test

import (
	"reflect"
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

	{[]byte{1, 2, 3}, []byte{codes.Bin8, 0x3, 0x1, 0x2, 0x3}},
	{[3]byte{1, 2, 3}, []byte{codes.Bin8, 0x3, 0x1, 0x2, 0x3}},

	{OmitEmptyTest{}, []byte{codes.FixedMapLow}},

	{intSet{}, []byte{codes.FixedArrayLow}},
	{intSet{8: struct{}{}}, []byte{codes.FixedArrayLow | 1, 0x8}},

	{&CompactEncodingTest{}, []byte{codes.FixedStrLow, codes.Nil, 0x0}},
	{
		&CompactEncodingTest{"a", &CompactEncodingTest{"b", nil, 7}, 6},
		[]byte{codes.FixedStrLow | 1, 'a', codes.FixedStrLow | 1, 'b', codes.Nil, 0x7, 0x6},
	},
}

func init() {
	test := binTest{in: &OmitEmptyTest{Foo: "hello"}}
	test.wanted = append(test.wanted, codes.FixedMapLow|1)
	test.wanted = append(test.wanted, codes.FixedStrLow|byte(len("Foo")))
	test.wanted = append(test.wanted, "Foo"...)
	test.wanted = append(test.wanted, codes.FixedStrLow|byte(len("hello")))
	test.wanted = append(test.wanted, "hello"...)
	binTests = append(binTests, test)

	test = binTest{in: &InlineTest{OmitEmptyTest: OmitEmptyTest{Bar: "world"}}}
	test.wanted = append(test.wanted, codes.FixedMapLow|1)
	test.wanted = append(test.wanted, codes.FixedStrLow|byte(len("Bar")))
	test.wanted = append(test.wanted, "Bar"...)
	test.wanted = append(test.wanted, codes.FixedStrLow|byte(len("world")))
	test.wanted = append(test.wanted, "world"...)
	binTests = append(binTests, test)
}

func TestBin(t *testing.T) {
	for _, test := range binTests {
		b, err := msgpack.Marshal(test.in)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(b, test.wanted) {
			t.Fatalf("%q != %q (in=%v)", b, test.wanted, test.in)
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
}

func (t *typeTest) assertErr(err error, s string) {
	if err == nil {
		t.Fatalf("got %v, wanted %q", err, s)
	}
	if err.Error() != s {
		t.Fatalf("got %q, wanted %q", err, s)
	}
}

var (
	typeTests = []typeTest{
		{in: make(chan bool), encErr: "msgpack: Encode(unsupported chan bool)"},

		{in: nil, out: nil, decErr: "msgpack: Decode(nil)"},
		{in: nil, out: 0, decErr: "msgpack: Decode(nonsettable int)"},
		{in: nil, out: (*int)(nil), decErr: "msgpack: Decode(nonsettable *int)"},
		{in: nil, out: new(chan bool), decErr: "msgpack: Decode(unsupported chan bool)"},

		{in: []int(nil), out: new([]int)},
		{in: make([]int, 0), out: new([]int)},
		{in: make([]int, 1000), out: new([]int)},
		{in: []interface{}{uint64(1), "hello"}, out: new([]interface{})},
		{in: map[string]interface{}{"foo": nil}, out: new(map[string]interface{})},

		{in: nil, out: new([]byte), wantnil: true},
		{in: []byte(nil), out: new([]byte)},
		{in: []byte{1, 2, 3}, out: new([]byte)},

		{in: nil, out: new([3]byte), wantnil: true},
		{in: [3]byte{1, 2, 3}, out: new([3]byte)},
		{in: [3]byte{1, 2, 3}, out: new([2]byte)},

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
	}
)

func deref(viface interface{}) interface{} {
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
			t.Fatalf("Unmarshal failed: %s (in=%#v out=%#v)", err, test.in, test.out)
		}

		out := deref(test.out)
		if test.wantnil {
			v := reflect.ValueOf(out)
			if v.IsNil() {
				return
			}
			t.Fatalf("got %#v, wanted nil", out)
			return
		}

		in := deref(test.in)
		if !reflect.DeepEqual(out, in) {
			t.Fatalf("%#v != %#v", out, in)
		}
	}
}
