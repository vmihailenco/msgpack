package msgpack_test

import (
	"testing"

	"gopkg.in/vmihailenco/msgpack.v2"
)

func init() {
	msgpack.RegisterExt(1, extType{})
}

type extType struct {
	S string
}

func (et *extType) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal(et.S)
}

func (et *extType) UnmarshalMsgpack(b []byte) error {
	return msgpack.Unmarshal(b, &et.S)
}

func TestExtType(t *testing.T) {
	b, err := msgpack.Marshal(extType{"hello"})
	if err != nil {
		t.Fatal(err)
	}

	var dst interface{}
	err = msgpack.Unmarshal(b, &dst)
	if err != nil {
		t.Fatal(err)
	}

	v, ok := dst.(extType)
	if !ok {
		t.Fatalf("got %T, wanted extType", dst)
	}
	if v.S != "hello" {
		t.Fatalf("got %q, wanted hello", v.S)
	}
}
