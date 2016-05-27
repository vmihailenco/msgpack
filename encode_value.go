package msgpack

import "reflect"

var valueEncoders []encoderFunc

func init() {
	valueEncoders = []encoderFunc{
		reflect.Bool:          encodeBoolValue,
		reflect.Int:           encodeInt64Value,
		reflect.Int8:          encodeInt64Value,
		reflect.Int16:         encodeInt64Value,
		reflect.Int32:         encodeInt64Value,
		reflect.Int64:         encodeInt64Value,
		reflect.Uint:          encodeUint64Value,
		reflect.Uint8:         encodeUint64Value,
		reflect.Uint16:        encodeUint64Value,
		reflect.Uint32:        encodeUint64Value,
		reflect.Uint64:        encodeUint64Value,
		reflect.Float32:       encodeFloat32Value,
		reflect.Float64:       encodeFloat64Value,
		reflect.Complex64:     encodeUnsupportedValue,
		reflect.Complex128:    encodeUnsupportedValue,
		reflect.Array:         encodeArrayValue,
		reflect.Chan:          encodeUnsupportedValue,
		reflect.Func:          encodeUnsupportedValue,
		reflect.Interface:     encodeInterfaceValue,
		reflect.Map:           encodeMapValue,
		reflect.Ptr:           encodeUnsupportedValue,
		reflect.Slice:         encodeSliceValue,
		reflect.String:        encodeStringValue,
		reflect.Struct:        encodeStructValue,
		reflect.UnsafePointer: encodeUnsupportedValue,
	}
}

func getEncoder(typ reflect.Type) encoderFunc {
	enc := _getEncoder(typ)
	if id := extTypeId(typ); id != -1 {
		return makeExtEncoder(id, enc)
	}
	return enc
}

func _getEncoder(typ reflect.Type) encoderFunc {
	kind := typ.Kind()

	if typ.Implements(encoderType) {
		return encodeCustomValue
	}

	// Addressable struct field value.
	if reflect.PtrTo(typ).Implements(encoderType) {
		return encodeCustomValuePtr
	}

	if typ.Implements(marshalerType) {
		return marshalValue
	}
	if encoder, ok := typEncMap[typ]; ok {
		return encoder
	}

	switch kind {
	case reflect.Ptr:
		return ptrEncoderFunc(typ)
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			return encodeByteSliceValue
		}
	case reflect.Array:
		if typ.Elem().Kind() == reflect.Uint8 {
			return encodeByteArrayValue
		}
	case reflect.Map:
		if typ.Key().Kind() == reflect.String {
			switch typ.Elem().Kind() {
			case reflect.String:
				return encodeMapStringStringValue
			case reflect.Interface:
				return encodeMapStringInterfaceValue
			}
		}
	}
	return valueEncoders[kind]
}
