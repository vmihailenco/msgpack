package msgpack

import (
	"io"
)

type Coder interface {
	EncodeMsgpack(io.Writer) error
	DecodeMsgpack(io.Reader) error
}
