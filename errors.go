package msgpack

import (
	"fmt"
)

// A DupExtIdError represents an error when is requested to register an extended
// type and its identifier already exists.
type DupExtIdError byte

// Error returns string representation of current instance error.
func (e DupExtIdError) Error() string {
	return fmt.Sprintf("ext with id %d is already registered", byte(e))
}

// An InvalidCodeError represents an error when the code of encoded data doesn't
// match which the type of provided variable.
type InvalidCodeError struct {
	Code     byte
	TypeName string
}

// Error returns string representation of current instance error.
func (e InvalidCodeError) Error() string {
	return fmt.Sprintf("msgpack: invalid code %x decoding %s",
		e.Code, e.TypeName)
}

// A NonAddressableError represents an error when is unable to obtain the
// pointer of destination variable.
type NonAddressableError struct {
	Value interface{}
}

// Error returns string representation of current instance error.
func (e NonAddressableError) Error() string {
	return fmt.Sprintf("msgpack: Encode(non-addressable %T)", e.Value)
}

// A NotSettableError represents an error when trying to decode to a variable
// that is not a pointer.
type NotSettableError struct {
	Value interface{}
}

// Error returns string representation of current instance error.
func (e NotSettableError) Error() string {
	return fmt.Sprintf("msgpack: Decode(nonsettable %T)", e.Value)
}

// A NullDestError represents an error when trying to decode to a null variable.
type NullDestError int

// Error returns string representation of current instance error.
func (e NullDestError) Error() string {
	return "msgpack: Decode(nil)"
}

// An UnknownCodeError represents an error when the code of encoded data doesn't
// match which the type of provided variable.
type UnknownCodeError struct {
	Code     byte
	TypeName string
}

// Error returns string representation of current instance error.
func (e UnknownCodeError) Error() string {
	if len(e.TypeName) > 0 {
		return fmt.Sprintf("msgpack: unknown code %x decoding %s",
			e.Code, e.TypeName)
	} else {
		return fmt.Sprintf("msgpack: unknown code %x",
			e.Code)
	}
}

// An UnregisteredExtError represents an error when trying to decode an
// unregistered extended type.
type UnregisteredExtError byte

// Error returns string representation of current instance error.
func (e UnregisteredExtError) Error() string {
	return fmt.Sprintf("msgpack: unregistered ext id %d", byte(e))
}

// An UnsupportedTypeError represents an error when trying to encode or decode
// an unsupported type.
type UnsupportedTypeError struct {
	Encoding bool
	Value    interface{}
}

// Error returns string representation of current instance error.
func (e UnsupportedTypeError) Error() string {
	var op string
	if e.Encoding {
		op = "Encode"
	} else {
		op = "Decode"
	}

	return fmt.Sprintf("msgpack: %s(unsupported %T)", op, e.Value)
}
