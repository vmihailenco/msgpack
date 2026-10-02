package msgpack_test

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/vmihailenco/msgpack/v5/msgpcode"
)

func TestDecodeUint64AsFloat(t *testing.T) {
	cases := []uint64{
		0,
		42,
		uint64(math.MaxInt64),
		uint64(1) << 63,
		math.MaxUint64,
	}
	for _, n := range cases {
		var buf bytes.Buffer
		buf.WriteByte(msgpcode.Uint64)
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], n)
		buf.Write(b[:])
		raw := buf.Bytes()

		dec := msgpack.NewDecoder(bytes.NewReader(raw))
		got, err := dec.DecodeFloat64()
		if err != nil {
			t.Fatal(err)
		}
		if got != float64(n) {
			t.Fatalf("DecodeFloat64(%d) = %v, want %v", n, got, float64(n))
		}

		dec = msgpack.NewDecoder(bytes.NewReader(raw))
		got32, err := dec.DecodeFloat32()
		if err != nil {
			t.Fatal(err)
		}
		if got32 != float32(n) {
			t.Fatalf("DecodeFloat32(%d) = %v, want %v", n, got32, float32(n))
		}

		var out float64
		if err := msgpack.Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}
		if out != float64(n) {
			t.Fatalf("Unmarshal(%d) = %v, want %v", n, out, float64(n))
		}
	}

	raw, err := msgpack.Marshal(int64(-3))
	if err != nil {
		t.Fatal(err)
	}
	var neg float64
	if err := msgpack.Unmarshal(raw, &neg); err != nil {
		t.Fatal(err)
	}
	if neg != -3 {
		t.Fatalf("negative int64 decoded as %v", neg)
	}

	raw, err = msgpack.Marshal(uint64(42))
	if err != nil {
		t.Fatal(err)
	}
	var small float64
	if err := msgpack.Unmarshal(raw, &small); err != nil {
		t.Fatal(err)
	}
	if small != 42 {
		t.Fatalf("uint64 42 decoded as %v", small)
	}
}
