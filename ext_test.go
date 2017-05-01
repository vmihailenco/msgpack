package msgpack_test

import (
	"reflect"
	"testing"

	"gopkg.in/vmihailenco/msgpack.v2"
	"gopkg.in/vmihailenco/msgpack.v2/codes"
)

func init() {
	msgpack.RegisterExt(9, extTest{})
}

func TestRegisterExtPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("panic expected")
		}
		got := r.(error).Error()
		wanted := "msgpack: ext with id=9 is already registered"
		if got != wanted {
			t.Fatalf("got %q, wanted %q", got, wanted)
		}
	}()
	msgpack.RegisterExt(9, extTest{})
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
	wanted := "msgpack: unregistered ext id=1"
	if got != wanted {
		t.Fatalf("got %q, wanted %q", got, wanted)
	}
}

func TestDecodeExtWithMap(t *testing.T) {
	type S struct {
		I int
	}
	msgpack.RegisterExt(2, S{})

	b, err := msgpack.Marshal(&S{I: 42})
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]interface{}
	if err := msgpack.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}

	wanted := map[string]interface{}{"I": uint64(42)}
	if !reflect.DeepEqual(got, wanted) {
		t.Fatalf("got %#v, but wanted %#v", got, wanted)
	}
}
