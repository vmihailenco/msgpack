package codes

import "errors"

var (
	PosFixedNumHigh byte = 0x7f
	NegFixedNumLow  byte = 0xe0

	Nil byte = 0xc0

	False byte = 0xc2
	True  byte = 0xc3

	Float  byte = 0xca
	Double byte = 0xcb

	Uint8  byte = 0xcc
	Uint16 byte = 0xcd
	Uint32 byte = 0xce
	Uint64 byte = 0xcf

	Int8  byte = 0xd0
	Int16 byte = 0xd1
	Int32 byte = 0xd2
	Int64 byte = 0xd3

	FixedStrLow  byte = 0xa0
	FixedStrHigh byte = 0xbf
	FixedStrMask byte = 0x1f
	Str8         byte = 0xd9
	Str16        byte = 0xda
	Str32        byte = 0xdb

	Bin8  byte = 0xc4
	Bin16 byte = 0xc5
	Bin32 byte = 0xc6

	FixedArrayLow  byte = 0x90
	FixedArrayHigh byte = 0x9f
	FixedArrayMask byte = 0xf
	Array16        byte = 0xdc
	Array32        byte = 0xdd

	FixedMapLow  byte = 0x80
	FixedMapHigh byte = 0x8f
	FixedMapMask byte = 0xf
	Map16        byte = 0xde
	Map32        byte = 0xdf

	FixExt1  byte = 0xd4
	FixExt2  byte = 0xd5
	FixExt4  byte = 0xd6
	FixExt8  byte = 0xd7
	FixExt16 byte = 0xd8
	Ext8     byte = 0xc7
	Ext16    byte = 0xc8
	Ext32    byte = 0xc9

	ErrNotExtType = errors.New("not an ext type")
)

func IsFixedNum(c byte) bool {
	return c <= PosFixedNumHigh || c >= NegFixedNumLow
}

func IsFixedMap(c byte) bool {
	return c >= FixedMapLow && c <= FixedMapHigh
}

func IsFixedArray(c byte) bool {
	return c >= FixedArrayLow && c <= FixedArrayHigh
}

func IsFixedString(c byte) bool {
	return c >= FixedStrLow && c <= FixedStrHigh
}

func IsExt(c byte) bool {
	return (c >= FixExt1 && c <= FixExt16) || (c >= Ext8 && c <= Ext32)
}

// ExtType provides a fast way to extract the type field from an ext type without fully unmarshaling it
func ExtType(value []byte) (int8, error) {
	if len(value) >= 2 {
		switch value[0] {
		case FixExt1, FixExt2, FixExt4, FixExt8, FixExt16:
			return int8(value[1]), nil
		case Ext8:
			if len(value) >= 3 {
				return int8(value[2]), nil
			}
		case Ext16:
			if len(value) >= 4 {
				return int8(value[3]), nil
			}
		case Ext32:
			if len(value) >= 6 {
				return int8(value[5]), nil
			}
		}
	}
	return 0, ErrNotExtType
}
