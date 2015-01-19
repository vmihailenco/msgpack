package msgpack_test

import (
	"reflect"
	"testing"
	"time"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type omitEmptyTest struct {
	Foo string `msgpack:",omitempty"`
	Bar string `msgpack:",omitempty"`
}

type binTest struct {
	in     interface{}
	wanted []byte
}

var binTests = []binTest{
	{nil, []byte{nilCode}},
	{omitEmptyTest{}, []byte{fixMapLowCode}},
}

func init() {
	test := binTest{in: &omitEmptyTest{Foo: "hello"}}
	test.wanted = append(test.wanted, fixMapLowCode|0x01)
	test.wanted = append(test.wanted, fixStrLowCode|byte(len("Foo")))
	test.wanted = append(test.wanted, "Foo"...)
	test.wanted = append(test.wanted, fixStrLowCode|byte(len("hello")))
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

var (
	stringsv           stringst
	structv            structt
	ints               []int
	interfaces         []interface{}
	unmarshalerPtr     *coderStruct
	coders             []coderStruct
	stringInterfaceMap map[string]interface{}
	embededTimeValue   *embededTime
	typeTests          = []typeTest{
		{stringst{"foo", "bar"}, &stringsv},
		{structt{stringst{"foo", "bar"}, []string{"hello"}}, &structv},
		{([]int)(nil), &ints},
		{make([]int, 0), &ints},
		{make([]int, 1000), &ints},
		{[]interface{}{int64(1), "hello"}, &interfaces},
		{map[string]interface{}{"foo": nil}, &stringInterfaceMap},
		{&coderStruct{name: "hello"}, &unmarshalerPtr},
		{&embededTime{Time: time.Now()}, &embededTimeValue},
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
