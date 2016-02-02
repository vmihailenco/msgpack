package msgpack_test

import (
	"testing"

	"gopkg.in/vmihailenco/msgpack.v2"
	"gopkg.in/vmihailenco/msgpack.v2/codes"
)

func init() {
	msgpack.RegisterExt(0, extTest{})
}

func TestRegisterExtPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("panic expected")
		}
		got := r.(error).Error()
		wanted := "ext with id 0 is already registered"
		if got != wanted {
			t.Fatalf("got %q, wanted %q", got, wanted)
		}
	}()
	msgpack.RegisterExt(0, extTest{})
}

type extTest struct {
	S string
}

type extTest2 struct {
	S string
}

func TestExt(t *testing.T) {
	for _, v := range []interface{}{extTest{"hello"}, &extTest{"hello"}} {
		b, err := msgpack.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}

		var dst interface{}
		err = msgpack.Unmarshal(b, &dst)
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
	err := msgpack.Unmarshal(b, &dst)
	if err == nil {
		t.Fatalf("got nil, wanted error")
	}
	got := err.Error()
	wanted := "msgpack: unregistered ext id 1"
	if got != wanted {
		t.Fatalf("got %q, wanted %q", got, wanted)
	}
}
