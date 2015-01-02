package msgpack_test

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

	fixStrLowCode  = 0xa0
	fixStrHighCode = 0xbf
	fixStrMask     = 0x1f
	str8Code       = 0xd9
	str16Code      = 0xda
	str32Code      = 0xdb

	bin8Code  = 0xc4
	bin16Code = 0xc5
	bin32Code = 0xc6

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
