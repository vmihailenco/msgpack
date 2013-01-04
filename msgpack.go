package msgpack

import (
	"io"
)

type encoder interface {
	EncodeMsgpack(io.Writer) error
}

type decoder interface {
	DecodeMsgpack(io.Reader) error
}

type Coder interface {
	encoder
	decoder
}
