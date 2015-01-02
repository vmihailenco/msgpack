package msgpack_test

import (
	"reflect"
	"testing"

	"github.com/vmihailenco/msgpack"
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

var (
	stringsv  stringst
	structv   structt
	typeTests = []typeTest{
		{stringst{"foo", "bar"}, &stringsv},
		{structt{stringst{"foo", "bar"}, []string{"hello"}}, &structv},
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

		out := deref(test.out)
		if !reflect.DeepEqual(out, test.in) {
			t.Fatalf("%#v != %#v", out, test.in)
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
