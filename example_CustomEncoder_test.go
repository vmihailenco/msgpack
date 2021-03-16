package msgpack_test

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
)

type customStruct struct {
	S string
	N int
}

var _ msgpack.CustomEncoder = (*customStruct)(nil)
var _ msgpack.CustomDecoder = (*customStruct)(nil)

func (s *customStruct) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeMulti(s.S, s.N)
}

func (s *customStruct) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.DecodeMulti(&s.S, &s.N)
}

func ExampleCustomEncoder() {
	b, err := msgpack.Marshal(&customStruct{S: "hello", N: 42})
	if err != nil {
		panic(err)
	}

	var v customStruct
	err = msgpack.Unmarshal(b, &v)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", v)
	// Output: msgpack_test.customStruct{S:"hello", N:42}
}

type NullInt struct {
	Valid bool
	Int   int
}

func (i *NullInt) Set(j int) {
	i.Int = j
	i.Valid = true
}

func (i NullInt) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal(i.Int)
}

func (i *NullInt) UnmarshalMsgpack(b []byte) error {
	if err := msgpack.Unmarshal(b, &i.Int); err != nil {
		return err
	}
	i.Valid = true
	return nil
}

type Secretive struct {
	Visible bool
	hidden  bool
}

func ExampleSimpleStructWithZeroValuesIsIgnoredWhenTaggedWithOmitempty() {
	type T struct {
		I NullInt `msgpack:",omitempty"`
		J NullInt
		// Secretive is not a "simple" struct because it has an hidden field.
		S Secretive `msgpack:",omitempty"`
	}

	var t1 T
	raw, err := msgpack.Marshal(t1)
	if err != nil {
		panic(err)
	}
	var t2 T
	if err = msgpack.Unmarshal(raw, &t2); err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", t2)

	t2.I.Set(42)
	t2.S.hidden = true // won't be included because it is a hidden field
	raw, err = msgpack.Marshal(t2)
	if err != nil {
		panic(err)
	}
	var t3 T
	if err = msgpack.Unmarshal(raw, &t3); err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", t3)
	// Output: msgpack_test.T{I:msgpack_test.NullInt{Valid:false, Int:0}, J:msgpack_test.NullInt{Valid:true, Int:0}, S:msgpack_test.Secretive{Visible:false, hidden:false}}
	// msgpack_test.T{I:msgpack_test.NullInt{Valid:true, Int:42}, J:msgpack_test.NullInt{Valid:true, Int:0}, S:msgpack_test.Secretive{Visible:false, hidden:false}}
}
