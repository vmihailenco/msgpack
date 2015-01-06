package msgpack // import "gopkg.in/vmihailenco/msgpack.v2"

type Marshaler interface {
	MarshalMsgpack() ([]byte, error)
}

type Unmarshaler interface {
	UnmarshalMsgpack([]byte) error
}
