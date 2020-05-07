package msgpack

type Marshaler interface {
	MarshalMsgpack() ([]byte, error)
}

type Unmarshaler interface {
	UnmarshalMsgpack([]byte) error
}

type CustomEncoder interface {
	EncodeMsgpack(*Encoder) error
}

type CustomDecoder interface {
	DecodeMsgpack(*Decoder) error
}

//------------------------------------------------------------------------------

type RawMessage []byte

var _ CustomEncoder = (RawMessage)(nil)
var _ CustomDecoder = (*RawMessage)(nil)

func (m RawMessage) EncodeMsgpack(enc *Encoder) error {
	return enc.write(m)
}

func (m *RawMessage) DecodeMsgpack(dec *Decoder) error {
	dec.rec = make([]byte, 0)
	if err := dec.Skip(); err != nil {
		return err
	}
	*m = dec.rec
	dec.rec = nil
	return nil
}
