package msgpack

const (
	posFixNumHighCode = 0x7f
	negFixNumLowCode  = 0xe0

	nilCode = 0xc0

	falseCode = 0xc2
	trueCode  = 0xc3

	floatCode  = 0xca
	doubleCode = 0xcb

	uint8Code  = 0xcc
	uint16Code = 0xcd
	uint32Code = 0xce
	uint64Code = 0xcf

	int8Code  = 0xd0
	int16Code = 0xd1
	int32Code = 0xd2
	int64Code = 0xd3

	fixRawLowCode  = 0xa0
	fixRawHighCode = 0xbf
	fixRawMask     = 0x1f
	raw16Code      = 0xda
	raw32Code      = 0xdb

	fixArrayLowCode  = 0x90
	fixArrayHighCode = 0x9f
	fixArrayMask     = 0xf
	array16Code      = 0xdc
	array32Code      = 0xdd

	fixMapLowCode  = 0x80
	fixMapHighCode = 0x8f
	fixMapMask     = 0xf
	map16Code      = 0xde
	map32Code      = 0xdf
)
