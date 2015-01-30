package msgpack

const (
	PosFixNumHighCode = 0x7f
	NegFixNumLowCode  = 0xe0

	NilCode = 0xc0

	FalseCode = 0xc2
	TrueCode  = 0xc3

	FloatCode  = 0xca
	DoubleCode = 0xcb

	Uint8Code  = 0xcc
	Uint16Code = 0xcd
	Uint32Code = 0xce
	Uint64Code = 0xcf

	Int8Code  = 0xd0
	Int16Code = 0xd1
	Int32Code = 0xd2
	Int64Code = 0xd3

	FixStrLowCode  = 0xa0
	FixStrHighCode = 0xbf
	FixStrMask     = 0x1f
	Str8Code       = 0xd9
	Str16Code      = 0xda
	Str32Code      = 0xdb

	Bin8Code  = 0xc4
	Bin16Code = 0xc5
	Bin32Code = 0xc6

	FixArrayLowCode  = 0x90
	FixArrayHighCode = 0x9f
	FixArrayMask     = 0xf
	Array16Code      = 0xdc
	Array32Code      = 0xdd

	FixMapLowCode  = 0x80
	FixMapHighCode = 0x8f
	FixMapMask     = 0xf
	Map16Code      = 0xde
	Map32Code      = 0xdf
)

type Code byte

func (c Code) IsFixNum() bool    { return c <= PosFixNumHighCode || c >= NegFixNumLowCode }
func (c Code) IsFixMap() bool    { return c >= FixMapLowCode && c <= FixMapHighCode }
func (c Code) IsFixSlice() bool  { return c >= FixArrayLowCode && c <= FixArrayHighCode }
func (c Code) IsFixString() bool { return c >= FixStrLowCode && c <= FixStrHighCode }
