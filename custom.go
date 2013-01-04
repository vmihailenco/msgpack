package msgpack

import (
	"reflect"
)

var (
	typEncMap = make(map[reflect.Type]encoderFunc)
	typDecMap = make(map[reflect.Type]decoderFunc)
)

type encoderFunc func(*Encoder, reflect.Value) error

type decoderFunc func(*Decoder, reflect.Value) error

func Register(typ reflect.Type, enc encoderFunc, dec decoderFunc) {
	typEncMap[typ] = enc
	typDecMap[typ] = dec
}
