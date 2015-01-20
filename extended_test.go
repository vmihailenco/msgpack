package msgpack

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"
)

type SimpleStruct struct {
	b []byte
}

var testBytes = [][]byte{
	[]byte{1, 2, 4, 8, 16},
	[]byte{1, 3, 9, 27, 81},
	[]byte{4, 9, 16, 25, 64},
	[]byte{1, 4, 16, 64},
	[]byte{1, 1, 2, 3, 5, 8, 13, 21},
}

func TestExtendedSimple(t *testing.T) {
	table := []struct {
		s *SimpleStruct
		b []byte
	}{
		{&SimpleStruct{testBytes[0]}, append([]byte{0xc7, 0x05, 0x07}, testBytes[0]...)},
		{&SimpleStruct{testBytes[1]}, append([]byte{0xc7, 0x05, 0x07}, testBytes[1]...)},
		{&SimpleStruct{testBytes[2]}, append([]byte{0xc7, 0x05, 0x07}, testBytes[2]...)},
		{&SimpleStruct{testBytes[3]}, append([]byte{0xd6, 0x07}, testBytes[3]...)},
		{&SimpleStruct{testBytes[4]}, append([]byte{0xd7, 0x07}, testBytes[4]...)},
	}

	EncodeSimpleStruct := func(v reflect.Value) ([]byte, error) {
		switch iv := v.Interface().(type) {
		case *SimpleStruct:
			return iv.b, nil
		}
		return nil, errors.New("unsupported encode type")
	}

	DecodeSimpleStruct := func(v reflect.Value, b []byte) error {
		es, ok := v.Interface().(*SimpleStruct)
		if !ok {
			return fmt.Errorf("Unexpected value: %T", v.Interface())
		}
		if es == nil {
			es = new(SimpleStruct)
			v.Set(reflect.ValueOf(es))
		}

		es.b = b
		return nil
	}

	extensions := NewExtensions()
	extensions.AddExtension(7, reflect.TypeOf(&SimpleStruct{}), nil, EncodeSimpleStruct, DecodeSimpleStruct)

	buf := &bytes.Buffer{}
	encoder := NewEncoder(buf)
	encoder.AddExtensions(extensions)
	for _, test := range table {
		if err := encoder.Encode(test.s); err != nil {
			t.Fatalf("Error encoding struct")
		}
		if bytes.Compare(buf.Bytes(), test.b) != 0 {
			t.Errorf("Unexpected byte value\n\tExpected: %x\n\tActual: %x", test.b, buf.Bytes())
		}

		d := NewDecoder(bytes.NewBuffer(buf.Bytes()))
		d.AddExtensions(extensions)
		var es SimpleStruct
		if err := d.Decode(&es); err != nil {
			t.Fatalf("Error decoding bytes: %s", err)
		}
		if bytes.Compare(test.s.b, es.b) != 0 {
			t.Errorf("Unexpected byte value\n\tExpected; %x\n\tActual: %x", test.s.b, es.b)
		}

		// Decode as interface
		d = NewDecoder(bytes.NewBuffer(buf.Bytes()))
		d.AddExtensions(extensions)
		var iVal interface{}
		if err := d.Decode(&iVal); err != nil {
			t.Fatalf("Error decoding bytes: %s", err)
		}
		if es, ok := iVal.(*SimpleStruct); !ok {
			t.Fatalf("Unexpected interface value: %T", iVal)
		} else if bytes.Compare(test.s.b, es.b) != 0 {
			t.Errorf("Unexpected byte value\n\tExpected; %x\n\tActual: %x", test.s.b, es.b)
		}

		buf.Reset()
	}

}

func (s *SimpleStruct) Read(b []byte) (int, error) {
	return copy(b, s.b), nil
}

func TestExtendedInterfaces(t *testing.T) {
	table := []struct {
		s *SimpleStruct
		b []byte
	}{
		{&SimpleStruct{testBytes[0]}, []byte{0xd4, 0x11, 0x00}},
		{&SimpleStruct{testBytes[1]}, []byte{0xd4, 0x11, 0x01}},
		{&SimpleStruct{testBytes[2]}, []byte{0xd4, 0x11, 0x02}},
		{&SimpleStruct{testBytes[3]}, []byte{0xd4, 0x11, 0x03}},
		{&SimpleStruct{testBytes[4]}, []byte{0xd4, 0x11, 0x04}},
	}

	type ReaderWrapper struct {
		io.Reader
	}
	readers := make([]*ReaderWrapper, 0)

	EncodeReader := func(iv reflect.Value) ([]byte, error) {
		switch v := iv.Interface().(type) {
		case *ReaderWrapper:
			readers = append(readers, v)
			return []byte{byte(len(readers) - 1)}, nil
		case io.Reader:
			readers = append(readers, &ReaderWrapper{v})
			return []byte{byte(len(readers) - 1)}, nil
		}

		return nil, fmt.Errorf("unsupported type: %T", iv.Interface())
	}
	DecodeReader := func(v reflect.Value, b []byte) error {
		if len(b) > 1 {
			return fmt.Errorf("unexpected error length: %d", len(b))
		}
		i := int(b[0])
		if i >= len(readers) {
			return fmt.Errorf("unknown index: %d", i)
		}
		if !v.CanSet() {
			v.Elem().Set(reflect.ValueOf(*readers[i]))
		} else {
			v.Set(reflect.ValueOf(readers[i]))
		}

		return nil
	}
	extensions := NewExtensions()
	extensions.AddExtension(17, reflect.TypeOf(new(ReaderWrapper)), []reflect.Type{reflect.TypeOf(new(io.Reader))}, EncodeReader, DecodeReader)

	buf := &bytes.Buffer{}
	encoder := NewEncoder(buf)
	encoder.AddExtensions(extensions)
	for _, test := range table {
		if err := encoder.Encode(test.s); err != nil {
			t.Fatalf("Error encoding struct")
		}
		if bytes.Compare(buf.Bytes(), test.b) != 0 {
			t.Errorf("Unexpected byte value\n\tExpected: %x\n\tActual: %x", test.b, buf.Bytes())
		}

		d := NewDecoder(bytes.NewBuffer(buf.Bytes()))
		d.AddExtensions(extensions)
		var rw ReaderWrapper
		if err := d.Decode(&rw); err != nil {
			t.Fatalf("Error decoding bytes: %s", err)
		}
		if rw.Reader != test.s {
			t.Errorf("Unexpected reader value\n\tExpected: %p\n\tActual: %p", test.s, rw.Reader)
		}
		readBytes := make([]byte, len(test.s.b))
		if n, err := rw.Read(readBytes); err != nil {
			t.Errorf("Error reading bytes: %s", err)
		} else if n != len(test.s.b) {
			t.Errorf("Unexpected read length\n\tExpected: %d\n\tActual: %d", len(test.s.b), n)
		}
		if bytes.Compare(test.s.b, readBytes) != 0 {
			t.Errorf("Unexpected byte value\n\tExpected: %x\n\tActual: %x", test.s.b, readBytes)
		}

		// Decode as interface
		d = NewDecoder(bytes.NewBuffer(buf.Bytes()))
		d.AddExtensions(extensions)
		var iVal io.Reader
		if err := d.Decode(&iVal); err != nil {
			t.Fatalf("Error decoding bytes: %s", err)
		}
		if rw, ok := iVal.(*ReaderWrapper); !ok {
			t.Fatalf("Unexpected interface value: %T", iVal)
		} else {
			if rw.Reader != test.s {
				t.Errorf("Unexpected reader value\n\tExpected: %p\n\tActual: %p", test.s, rw.Reader)
			}
			readBytes := make([]byte, len(test.s.b))
			if n, err := rw.Read(readBytes); err != nil {
				t.Errorf("Error reading bytes: %s", err)
			} else if n != len(test.s.b) {
				t.Errorf("Unexpected read length\n\tExpected: %d\n\tActual: %d", len(test.s.b), n)
			}
			if bytes.Compare(test.s.b, readBytes) != 0 {
				t.Errorf("Unexpected byte value\n\tExpected: %x\n\tActual: %x", test.s.b, readBytes)
			}
		}

		buf.Reset()
	}
}
