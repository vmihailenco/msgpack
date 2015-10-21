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
	*testing.T

	in     interface{}
	out    interface{}
	encErr string
	decErr string
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
		{in: nil, out: new(chan bool), decErr: "msgpack: Decode(unsupported chan bool)"},

		{in: []int(nil), out: new([]int)},
		{in: make([]int, 0), out: new([]int)},
		{in: make([]int, 1000), out: new([]int)},
		{in: []interface{}{int64(1), "hello"}, out: new([]interface{})},
		{in: map[string]interface{}{"foo": nil}, out: new(map[string]interface{})},

		{in: stringSlice{"foo", "bar"}, out: new(stringSlice)},
		{in: []stringAlias{"hello"}, out: new([]stringAlias)},
		{in: []uint8Alias{1}, out: new([]uint8Alias)},

		{in: intSet{}, out: new(intSet)},
		{in: intSet{8: struct{}{}}, out: new(intSet)},

		{in: testStruct{stringSlice{"foo", "bar"}, []string{"hello"}}, out: new(testStruct)},
		{in: &coderStruct{name: "hello"}, out: new(*coderStruct)},
		{in: &embeddedTime{Time: time.Now()}, out: new(*embeddedTime)},

		{in: &compactEncoding{}, out: new(compactEncoding)},
		{in: &compactEncoding{"a", &compactEncoding{"b", nil, 1}, 2}, out: new(compactEncoding)},
		{
			in:  &compactEncodingFieldValue{Field: compactEncoding{"a", nil, 1}},
			out: new(compactEncodingFieldValue),
		},
	}
)

func TestTypes(t *testing.T) {
	for _, test := range typeTests {
		test.T = t

		b, err := msgpack.Marshal(test.in)
		if test.encErr != "" {
			test.assertErr(err, test.encErr)
			continue
		}
		if err != nil {
			t.Fatal(err)
		}

		err = msgpack.Unmarshal(b, test.out)
		if test.decErr != "" {
			test.assertErr(err, test.decErr)
			continue
		}
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
