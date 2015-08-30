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
	_ msgpack.CustomEncoder = &intSet{}
	_ msgpack.CustomDecoder = &intSet{}
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

type compactEncoding struct {
	str string
	ref *compactEncoding
	num int
}

var (
	_ msgpack.CustomEncoder = &compactEncoding{}
	_ msgpack.CustomDecoder = &compactEncoding{}
)

func (s *compactEncoding) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.Encode(s.str, s.ref, s.num)
}

func (s *compactEncoding) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.Decode(&s.str, &s.ref, &s.num)
}

type compactEncodingFieldValue struct {
	Field compactEncoding
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
	{OmitEmptyTest{}, []byte{codes.FixedMapLow}},

	{intSet{}, []byte{codes.FixedArrayLow}},
	{intSet{8: struct{}{}}, []byte{codes.FixedArrayLow | 1, 0x8}},

	{&compactEncoding{}, []byte{codes.FixedStrLow, codes.Nil, 0x0}},
	{
		&compactEncoding{"n", &compactEncoding{"o", nil, 7}, 6},
		[]byte{codes.FixedStrLow | 1, 'n', codes.FixedStrLow | 1, 'o', codes.Nil, 0x7, 0x6},
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
			t.Fatalf("% x != % x", b, test.wanted)
		}
	}
}

//------------------------------------------------------------------------------

type embeddedTime struct {
	time.Time
}

type (
	stringAlias string
	uint8Alias  uint8
	stringSlice []string
)

type testStruct struct {
	F1 stringSlice
	F2 []string
}

type typeTest struct {
	in  interface{}
	out interface{}
}

var (
	typeTests = []typeTest{
		{stringSlice{"foo", "bar"}, new(stringSlice)},
		{([]int)(nil), new([]int)},
		{make([]int, 0), new([]int)},
		{make([]int, 1000), new([]int)},
		{[]interface{}{int64(1), "hello"}, new([]interface{})},
		{map[string]interface{}{"foo": nil}, new(map[string]interface{})},

		{[]stringAlias{"hello"}, new([]stringAlias)},
		{[]uint8Alias{1}, new([]uint8Alias)},

		{intSet{}, new(intSet)},
		{intSet{8: struct{}{}}, new(intSet)},

		{testStruct{stringSlice{"foo", "bar"}, []string{"hello"}}, new(testStruct)},
		{&coderStruct{name: "hello"}, new(*coderStruct)},
		{&embeddedTime{Time: time.Now()}, new(*embeddedTime)},

		{&compactEncoding{}, new(compactEncoding)},
		{&compactEncoding{"a", &compactEncoding{"b", nil, 1}, 2}, new(compactEncoding)},
		{&compactEncodingFieldValue{Field: compactEncoding{"a", nil, 1}}, new(compactEncodingFieldValue)},
	}
)

func TestTypes(t *testing.T) {
	for _, test := range typeTests {
		b, err := msgpack.Marshal(test.in)
		if err != nil {
			t.Fatal(err)
		}

		err = msgpack.Unmarshal(b, test.out)
		if err != nil {
			t.Fatal(err)
		}

		in := deref(test.in)
		out := deref(test.out)
		if !reflect.DeepEqual(out, in) {
			t.Fatalf("%#v != %#v", out, in)
		}
	}
}

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
