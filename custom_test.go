package msgpack_test

import (
	"bytes"
	"reflect"
	"testing"

	"gopkg.in/vmihailenco/msgpack.v2"
)

// ------------------------------------------------------------------------

func init() {
	msgpack.Register(reflect.TypeOf(customTestSet{}), encodeCustomTestSet, decodeCustomTestSet)
}

type customTestSet map[int]struct{}

func encodeCustomTestSet(e *msgpack.Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}

	set := v.Interface().(customTestSet)
	slc := make([]int, 0, len(set))
	for n, _ := range set {
		slc = append(slc, n)
	}
	return e.Encode(slc)
}

func decodeCustomTestSet(d *msgpack.Decoder, v reflect.Value) error {
	if code, err := d.PeekCode(); err != nil || code == msgpack.NilCode {
		return err
	}

	sln, err := d.DecodeSliceLen()
	if err != nil {
		return err
	}
	set := make(customTestSet, sln)
	for i := 0; i < sln; i++ {
		n, err := d.DecodeInt()
		if err != nil {
			return err
		}
		set[n] = struct{}{}
	}
	v.Set(reflect.ValueOf(set))
	return nil
}

// ------------------------------------------------------------------------

func TestCustom_NilOrBlank(t *testing.T) {
	tests := []struct {
		set customTestSet
		bin []byte
	}{
		{customTestSet{},
			[]byte{0x90}},
		{customTestSet{1: struct{}{}, 3: struct{}{}},
			[]byte{0x92, 0x1, 0x3}},
		{nil,
			[]byte{0xc0}},
	}

	for _, test := range tests {
		bin, err := msgpack.Marshal(test.set)
		if err != nil {
			t.Fatal("expected no errors to occur, but got %#v", err)
		} else if !bytes.Equal(bin, test.bin) {
			t.Fatalf("expected %#v, but got %#v", test.bin, bin)
		}

		var set customTestSet
		err = msgpack.Unmarshal(test.bin, &set)
		if err != nil {
			t.Fatal("expected no errors to occur, but got %#v", err)
		} else if !reflect.DeepEqual(set, test.set) {
			t.Fatalf("expected %#v, but got %#v", test.set, set)
		}
	}
}
