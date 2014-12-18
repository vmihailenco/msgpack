package msgpack

type Marshaler interface {
	MarshalMsgpack() ([]byte, error)
}

type Unmarshaler interface {
	UnmarshalMsgpack([]byte) error
}
