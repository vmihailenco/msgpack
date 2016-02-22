package msgpack

import (
	"testing"

	"gopkg.in/vmihailenco/msgpack.v2/codes"
)

func init() {
	RegisterExt(0, extTest{})
}

func TestRegisterExtPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("panic expected")
		}
		_, ok := r.(DupExtIdError)
		if !ok {
			t.Fatalf("Unexpected panic type: %T", r)
		}
	}()
	RegisterExt(0, extTest{})
}

type extTest struct {
	S string
}

type extTest2 struct {
	S string
}

func TestExt(t *testing.T) {
	for _, v := range []interface{}{extTest{"hello"}, &extTest{"hello"}} {
		b, err := Marshal(v)
		if err != nil {
			t.Fatal(err)
		}

		var dst interface{}
		err = Unmarshal(b, &dst)
		if err != nil {
			t.Fatal(err)
		}

		v, ok := dst.(extTest)
		if !ok {
			t.Fatalf("got %#v, wanted extTest", dst)
		}
		if v.S != "hello" {
			t.Fatalf("got %q, wanted hello", v.S)
		}
	}
}

func TestUnknownExt(t *testing.T) {
	b := []byte{codes.FixExt1, 1, 0}

	var dst interface{}
	err := Unmarshal(b, &dst)
	if err == nil {
		t.Fatalf("got nil, wanted error")
	}
	_, ok := err.(UnregisteredExtError)
	if !ok {
		t.Fatalf("Unexpected error type: %T", err)
	}
}
