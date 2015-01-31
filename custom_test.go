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
	msgpack.Register(reflect.TypeOf(&customTestStruct{}), encodeCustomTestStruct, decodeCustomTestStruct)
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
	sln, err := d.DecodeSliceLen()
	if err != nil || sln < 0 {
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

type customTestStruct struct {
	s string
	n int
}

func encodeCustomTestStruct(e *msgpack.Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}

	o := v.Interface().(*customTestStruct)
	return e.Encode(o.s, o.n)
}

func decodeCustomTestStruct(d *msgpack.Decoder, v reflect.Value) error {
	ok, err := d.DecodeNil()
	if err != nil || ok {
		return err
	}

	var o customTestStruct
	if err = d.Decode(&o.s, &o.n); err == nil {
		v.Set(reflect.ValueOf(&o))
	}
	return err
}

// ------------------------------------------------------------------------

func TestCustom_Slices(t *testing.T) {
	tests := []struct {
		set customTestSet
		bin []byte
	}{
		{customTestSet{},
			[]byte{0x90}},
		{customTestSet{8: struct{}{}},
			[]byte{0x91, 0x8}},
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

func TestCustom_Structs(t *testing.T) {
	tests := []struct {
		sct *customTestStruct
		bin []byte
	}{
		{&customTestStruct{"n", 6},
			[]byte{0xa1, 'n', 0x6}},
		{&customTestStruct{},
			[]byte{0xa0, 0x0}},
		{nil,
			[]byte{0xc0}},
	}

	for _, test := range tests {
		bin, err := msgpack.Marshal(test.sct)
		if err != nil {
			t.Fatal("expected no errors to occur, but got %#v", err)
		} else if !bytes.Equal(bin, test.bin) {
			t.Fatalf("expected %#v, but got %#v", test.bin, bin)
		}

		var sct *customTestStruct
		err = msgpack.Unmarshal(test.bin, &sct)
		if err != nil {
			t.Fatal("expected no errors to occur, but got %#v", err)
		} else if !reflect.DeepEqual(sct, test.sct) {
			t.Fatalf("expected %#v, but got %#v", test.sct, sct)
		}
	}
}
