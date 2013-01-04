package msgpack

import (
	"io"
	"reflect"
)

var (
	coderType = reflect.TypeOf((*Coder)(nil)).Elem()
)

type Coder interface {
	EncodeMsgpack(io.Writer) error
	DecodeMsgpack(io.Reader) error
}
