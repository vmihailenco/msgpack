package msgpack

import (
	"reflect"
)

var (
	typEncMap = make(map[reflect.Type]encoder)
	typDecMap = make(map[reflect.Type]decoder)
)

type encoder func(*Encoder, reflect.Value) error

type decoder func(*Decoder, reflect.Value) error

func Register(typ reflect.Type, enc encoder, dec decoder) {
	typEncMap[typ] = enc
	typDecMap[typ] = dec
}
