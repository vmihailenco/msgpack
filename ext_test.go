package msgpack_test

import (
	"bytes"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/vmihailenco/msgpack/v5/msgpcode"
)

func init() {
	msgpack.RegisterExt(9, (*ExtTest)(nil))
}

type ExtTest struct {
	S string
}

var (
	_ msgpack.Marshaler   = (*ExtTest)(nil)
	_ msgpack.Unmarshaler = (*ExtTest)(nil)
)

func (ext ExtTest) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal("hello " + ext.S)
}

func (ext *ExtTest) UnmarshalMsgpack(b []byte) error {
	return msgpack.Unmarshal(b, &ext.S)
}

func TestEncodeDecodeExtHeader(t *testing.T) {
	v := &ExtTest{"world"}

	payload, err := v.MarshalMsgpack()
	require.Nil(t, err)

	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	err = enc.EncodeExtHeader(9, len(payload))
	require.Nil(t, err)

	_, err = buf.Write(payload)
	require.Nil(t, err)

	var dst interface{}
	err = msgpack.Unmarshal(buf.Bytes(), &dst)
	require.Nil(t, err)

	v = dst.(*ExtTest)
	wanted := "hello world"
	require.Equal(t, v.S, wanted)

	dec := msgpack.NewDecoder(&buf)
	extID, extLen, err := dec.DecodeExtHeader()
	require.Nil(t, err)
	require.Equal(t, int8(9), extID)
	require.Equal(t, len(payload), extLen)

	data := make([]byte, extLen)
	err = dec.ReadFull(data)
	require.Nil(t, err)

	v = &ExtTest{}
	err = v.UnmarshalMsgpack(data)
	require.Nil(t, err)
	require.Equal(t, wanted, v.S)
}

func TestExt(t *testing.T) {
	v := &ExtTest{"world"}
	b, err := msgpack.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	var dst interface{}
	err = msgpack.Unmarshal(b, &dst)
	if err != nil {
		t.Fatal(err)
	}

	v, ok := dst.(*ExtTest)
	if !ok {
		t.Fatalf("got %#v, wanted ExtTest", dst)
	}

	wanted := "hello world"
	if v.S != wanted {
		t.Fatalf("got %q, wanted %q", v.S, wanted)
	}

	ext := new(ExtTest)
	err = msgpack.Unmarshal(b, &ext)
	if err != nil {
		t.Fatal(err)
	}
	if ext.S != wanted {
		t.Fatalf("got %q, wanted %q", ext.S, wanted)
	}
}

func TestUnknownExt(t *testing.T) {
	b := []byte{byte(msgpcode.FixExt1), 2, 0}

	var dst interface{}
	err := msgpack.Unmarshal(b, &dst)
	if err == nil {
		t.Fatalf("got nil, wanted error")
	}
	got := err.Error()
	wanted := "msgpack: unknown ext id=2"
	if got != wanted {
		t.Fatalf("got %q, wanted %q", got, wanted)
	}
}

func TestSliceOfTime(t *testing.T) {
	in := []interface{}{time.Now()}
	b, err := msgpack.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	var out []interface{}
	err = msgpack.Unmarshal(b, &out)
	if err != nil {
		t.Fatal(err)
	}

	outTime := out[0].(time.Time)
	inTime := in[0].(time.Time)
	if outTime.Unix() != inTime.Unix() {
		t.Fatalf("got %v, wanted %v", outTime, inTime)
	}
}

type customPayload struct {
	payload []byte
}

func (cp *customPayload) MarshalMsgpack() ([]byte, error) {
	return cp.payload, nil
}

func (cp *customPayload) UnmarshalMsgpack(b []byte) error {
	cp.payload = b
	return nil
}

func TestDecodeCustomPayload(t *testing.T) {
	b, err := hex.DecodeString("c70500c09eec3100")
	if err != nil {
		t.Fatal(err)
	}

	msgpack.RegisterExt(0, (*customPayload)(nil))

	var cp *customPayload
	err = msgpack.Unmarshal(b, &cp)
	if err != nil {
		t.Fatal(err)
	}

	payload := hex.EncodeToString(cp.payload)
	wanted := "c09eec3100"
	if payload != wanted {
		t.Fatalf("got %q, wanted %q", payload, wanted)
	}
}
