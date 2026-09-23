package msgpack_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
)

type codedByte uint8

func (v codedByte) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeString(fmt.Sprintf("role-%d", v))
}

func (v *codedByte) DecodeMsgpack(dec *msgpack.Decoder) error {
	var n uint8
	s, err := dec.DecodeString()
	if err != nil {
		return err
	}
	if _, err := fmt.Sscanf(s, "role-%d", &n); err != nil {
		return err
	}
	*v = codedByte(n)
	return nil
}

func TestByteElementCodecTakesPrecedenceOverBinaryShortcut(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{"slice", []codedByte{1, 2}, []codedByte{1, 2}},
		{"array", [2]codedByte{1, 2}, [2]codedByte{1, 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := msgpack.Marshal(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if encoded[0] != 0x92 {
				t.Fatalf("encoded %T as %x, want a 2-element MessagePack array", tc.input, encoded)
			}
			out := reflect.New(reflect.TypeOf(tc.input))
			if err := msgpack.Unmarshal(encoded, out.Interface()); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(out.Elem().Interface(), tc.want) {
				t.Fatalf("roundtrip got %#v, want %#v", out.Elem().Interface(), tc.want)
			}
		})
	}
	for _, input := range []interface{}{[]byte{1, 2}, []plainByte{1, 2}, [2]plainByte{1, 2}} {
		encoded, err := msgpack.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		if encoded[0] != 0xc4 {
			t.Errorf("encoded plain bytes %T as %x, want binary", input, encoded)
		}
		out := reflect.New(reflect.TypeOf(input))
		if err := msgpack.Unmarshal(encoded, out.Interface()); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(out.Elem().Interface(), input) {
			t.Errorf("plain byte roundtrip got %#v, want %#v", out.Elem().Interface(), input)
		}
	}
}

type plainByte uint8
