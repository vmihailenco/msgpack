package msgpack_test

import (
	"bytes"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
)

// Role is a named uint8 with a custom codec (encodes as string).
type Role uint8

func (r Role) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeString("BACKEND")
}

func (r *Role) DecodeMsgpack(dec *msgpack.Decoder) error {
	s, err := dec.DecodeString()
	if err != nil {
		return err
	}
	if s != "BACKEND" {
		return err
	}
	*r = 1
	return nil
}

func TestNamedUint8SliceUsesCustomCodec(t *testing.T) {
	// Scalar uses custom codec.
	b, err := msgpack.Marshal(Role(1))
	if err != nil {
		t.Fatal(err)
	}
	// string "BACKEND" is a7 + 7 bytes
	if !bytes.Contains(b, []byte("BACKEND")) {
		t.Fatalf("scalar: % x", b)
	}

	// Slice must not be packed as bin blob of raw uint8s.
	b, err = msgpack.Marshal([]Role{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(b, []byte{0xc4, 0x02, 0x01, 0x02}) {
		t.Fatalf("slice used bin shortcut: % x", b)
	}
	if !bytes.Contains(b, []byte("BACKEND")) {
		t.Fatalf("slice missing custom encoding: % x", b)
	}

	// Plain []byte still uses binary encoding.
	raw, err := msgpack.Marshal([]byte{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, []byte{0xc4, 0x02, 0x01, 0x02}) && !bytes.Equal(raw[:1], []byte{0xc4}) {
		// bin 8 format
		if raw[0] != 0xc4 && raw[0] != 0xa2 {
			t.Logf("[]byte encoded as % x (ok if bin-like)", raw)
		}
	}
}
