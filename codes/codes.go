package codes

const (
	PosFixedNumHigh = 0x7f
	NegFixedNumLow  = 0xe0

	Nil = 0xc0

	False = 0xc2
	True  = 0xc3

	Float  = 0xca
	Double = 0xcb

	Uint8  = 0xcc
	Uint16 = 0xcd
	Uint32 = 0xce
	Uint64 = 0xcf

	Int8  = 0xd0
	Int16 = 0xd1
	Int32 = 0xd2
	Int64 = 0xd3

	FixedStrLow  = 0xa0
	FixedStrHigh = 0xbf
	FixedStrMask = 0x1f
	Str8         = 0xd9
	Str16        = 0xda
	Str32        = 0xdb

	Bin8  = 0xc4
	Bin16 = 0xc5
	Bin32 = 0xc6

	FixedArrayLow  = 0x90
	FixedArrayHigh = 0x9f
	FixedArrayMask = 0xf
	Array16        = 0xdc
	Array32        = 0xdd

	FixedMapLow  = 0x80
	FixedMapHigh = 0x8f
	FixedMapMask = 0xf
	Map16        = 0xde
	Map32        = 0xdf

	FixExt1  = 0xd4
	FixExt2  = 0xd5
	FixExt4  = 0xd6
	FixExt8  = 0xd7
	FixExt16 = 0xd8
	Ext8     = 0xc7
	Ext16    = 0xc8
	Ext32    = 0xc9
)

func IsFixedNum(c byte) bool    { return c <= PosFixedNumHigh || c >= NegFixedNumLow }
func IsFixedMap(c byte) bool    { return c >= FixedMapLow && c <= FixedMapHigh }
func IsFixedArray(c byte) bool  { return c >= FixedArrayLow && c <= FixedArrayHigh }
func IsFixedString(c byte) bool { return c >= FixedStrLow && c <= FixedStrHigh }
