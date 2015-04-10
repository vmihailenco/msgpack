package msgpack_test

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/vmihailenco/msgpack/codes"
	"gopkg.in/vmihailenco/msgpack.v2"
)

//------------------------------------------------------------------------------

type intSet map[int]struct{}

func (set intSet) MarshalMsgpack() ([]byte, error) {
	var slice []int

	if set == nil {
		return msgpack.Marshal(slice)
	}

	slice = make([]int, 0, len(set))
	for n, _ := range set {
		slice = append(slice, n)
	}
	return msgpack.Marshal(slice)
}

func (setptr *intSet) UnmarshalMsgpack(b []byte) error {
	d := msgpack.NewDecoder(bytes.NewReader(b))

	n, err := d.DecodeSliceLen()
	if err != nil {
		return err
	}
	if n == -1 {
		return nil
	}

	set := make(intSet, n)
	for i := 0; i < n; i++ {
		n, err := d.DecodeInt()
		if err != nil {
			return err
		}
		set[n] = struct{}{}
	}
	*setptr = set

	return nil
}

type compactEncoding struct {
	str     string
	struct_ *compactEncoding
	num     int
}

func (s *compactEncoding) MarshalMsgpack() ([]byte, error) {
	if s == nil {
		return []byte{codes.Nil}, nil
	}
	return msgpack.Marshal(s.str, s.struct_, s.num)
}

func (s *compactEncoding) UnmarshalMsgpack(b []byte) error {
	if len(b) == 1 && b[0] == codes.Nil {
		return nil
	}
	return msgpack.Unmarshal(b, &s.str, &s.struct_, &s.num)
}

//------------------------------------------------------------------------------

type omitEmptyTest struct {
	Foo string `msgpack:",omitempty"`
	Bar string `msgpack:",omitempty"`
}

//------------------------------------------------------------------------------

type binTest struct {
	in     interface{}
	wanted []byte
}

var binTests = []binTest{
	{nil, []byte{codes.Nil}},
	{omitEmptyTest{}, []byte{codes.FixedMapLow}},

	{intSet{}, []byte{codes.FixedArrayLow}},
	{intSet{8: struct{}{}}, []byte{codes.FixedArrayLow | 1, 0x8}},

	{&compactEncoding{}, []byte{codes.FixedStrLow, codes.Nil, 0x0}},
	{
		&compactEncoding{"n", &compactEncoding{"o", nil, 7}, 6},
		[]byte{codes.FixedStrLow | 1, 'n', codes.FixedStrLow | 1, 'o', codes.Nil, 0x7, 0x6},
	},
}

func init() {
	test := binTest{in: &omitEmptyTest{Foo: "hello"}}
	test.wanted = append(test.wanted, codes.FixedMapLow|0x01)
	test.wanted = append(test.wanted, codes.FixedStrLow|byte(len("Foo")))
	test.wanted = append(test.wanted, "Foo"...)
	test.wanted = append(test.wanted, codes.FixedStrLow|byte(len("hello")))
	test.wanted = append(test.wanted, "hello"...)
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

type typeTest struct {
	in  interface{}
	out interface{}
}

type stringst []string

type structt struct {
	F1 stringst
	F2 []string
}

type embededTime struct {
	time.Time
}

type (
	stringAlias string
	uint8Alias  uint8
)

var (
	stringsv           stringst
	structv            structt
	ints               []int
	interfaces         []interface{}
	unmarshalerPtr     *coderStruct
	coders             []coderStruct
	stringInterfaceMap map[string]interface{}
	embededTimeValue   *embededTime

	stringAliasSliceValue []stringAlias
	uint8AliasSliceValue  []uint8Alias

	intSetValue          intSet
	compactEncodingValue *compactEncoding

	typeTests = []typeTest{
		{stringst{"foo", "bar"}, &stringsv},
		{structt{stringst{"foo", "bar"}, []string{"hello"}}, &structv},
		{([]int)(nil), &ints},
		{make([]int, 0), &ints},
		{make([]int, 1000), &ints},
		{[]interface{}{int64(1), "hello"}, &interfaces},
		{map[string]interface{}{"foo": nil}, &stringInterfaceMap},
		{&coderStruct{name: "hello"}, &unmarshalerPtr},
		{&embededTime{Time: time.Now()}, &embededTimeValue},

		{[]stringAlias{"hello"}, &stringAliasSliceValue},
		{[]uint8Alias{1}, &uint8AliasSliceValue},

		{intSet{}, &intSetValue},
		{intSet{8: struct{}{}}, &intSetValue},

		{&compactEncoding{}, &compactEncodingValue},
		{&compactEncoding{"n", &compactEncoding{"o", nil, 7}, 6}, &compactEncodingValue},
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
