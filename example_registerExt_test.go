package msgpack_test

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

// https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1#eventtime-ext-format
type EventTime struct {
	time.Time
}

type OneMoreSecondEventTime struct {
	EventTime
}

var (
	_ msgpack.Marshaler   = (*EventTime)(nil)
	_ msgpack.Unmarshaler = (*EventTime)(nil)
	_ msgpack.Marshaler   = (*OneMoreSecondEventTime)(nil)
	_ msgpack.Unmarshaler = (*OneMoreSecondEventTime)(nil)
)

func (tm *EventTime) MarshalMsgpack() ([]byte, error) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint32(b, uint32(tm.Unix()))
	binary.BigEndian.PutUint32(b[4:], uint32(tm.Nanosecond()))
	return b, nil
}

func (tm *EventTime) UnmarshalMsgpack(b []byte) error {
	if len(b) != 8 {
		return fmt.Errorf("invalid data length: got %d, wanted 8", len(b))
	}
	sec := binary.BigEndian.Uint32(b)
	usec := binary.BigEndian.Uint32(b[4:])
	tm.Time = time.Unix(int64(sec), int64(usec))
	return nil
}

func (tm *OneMoreSecondEventTime) MarshalMsgpack() ([]byte, error) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint32(b, uint32(tm.Unix()+1))
	binary.BigEndian.PutUint32(b[4:], uint32(tm.Nanosecond()))
	return b, nil
}

func (tm *OneMoreSecondEventTime) UnmarshalMsgpack(b []byte) error {
	if len(b) != 8 {
		return fmt.Errorf("invalid data length: got %d, wanted 8", len(b))
	}
	sec := binary.BigEndian.Uint32(b)
	usec := binary.BigEndian.Uint32(b[4:])
	tm.Time = time.Unix(int64(sec+1), int64(usec))
	return nil
}

func ExampleRegisterExt() {
	t := time.Unix(123456789, 123)

	{
		msgpack.RegisterExt(1, (*EventTime)(nil))
		b, err := msgpack.Marshal(&EventTime{t})
		if err != nil {
			panic(err)
		}

		var v interface{}
		err = msgpack.Unmarshal(b, &v)
		if err != nil {
			panic(err)
		}
		fmt.Println(v.(*EventTime).UTC())

		tm := new(EventTime)
		err = msgpack.Unmarshal(b, &tm)
		if err != nil {
			panic(err)
		}
		fmt.Println(tm.UTC())
	}

	{
		msgpack.RegisterExt(1, (*EventTime)(nil))
		b, err := msgpack.Marshal(&EventTime{t})
		if err != nil {
			panic(err)
		}

		// override ext
		msgpack.RegisterExt(1, (*OneMoreSecondEventTime)(nil))

		var v interface{}
		err = msgpack.Unmarshal(b, &v)
		if err != nil {
			panic(err)
		}
		fmt.Println(v.(*OneMoreSecondEventTime).UTC())
	}

	{
		msgpack.RegisterExt(1, (*OneMoreSecondEventTime)(nil))
		b, err := msgpack.Marshal(&OneMoreSecondEventTime{
			EventTime{t},
		})
		if err != nil {
			panic(err)
		}

		// override ext
		msgpack.RegisterExt(1, (*EventTime)(nil))
		var v interface{}
		err = msgpack.Unmarshal(b, &v)
		if err != nil {
			panic(err)
		}
		fmt.Println(v.(*EventTime).UTC())
	}

	// Output: 1973-11-29 21:33:09.000000123 +0000 UTC
	// 1973-11-29 21:33:09.000000123 +0000 UTC
	// 1973-11-29 21:33:10.000000123 +0000 UTC
	// 1973-11-29 21:33:10.000000123 +0000 UTC
}

func ExampleUnregisterExt() {
	t := time.Unix(123456789, 123)

	{
		msgpack.RegisterExt(1, (*EventTime)(nil))
		b, err := msgpack.Marshal(&EventTime{t})
		if err != nil {
			panic(err)
		}

		msgpack.UnregisterExt(1)

		var v interface{}
		err = msgpack.Unmarshal(b, &v)
		wanted := "msgpack: unknown ext id=1"
		if err.Error() != wanted {
			panic(err)
		}

		msgpack.RegisterExt(1, (*OneMoreSecondEventTime)(nil))
		err = msgpack.Unmarshal(b, &v)
		if err != nil {
			panic(err)
		}
		fmt.Println(v.(*OneMoreSecondEventTime).UTC())
	}

	{
		msgpack.RegisterExt(1, (*OneMoreSecondEventTime)(nil))
		b, err := msgpack.Marshal(&OneMoreSecondEventTime{
			EventTime{t},
		})
		if err != nil {
			panic(err)
		}

		msgpack.UnregisterExt(1)
		var v interface{}
		err = msgpack.Unmarshal(b, &v)
		wanted := "msgpack: unknown ext id=1"
		if err.Error() != wanted {
			panic(err)
		}

		msgpack.RegisterExt(1, (*EventTime)(nil))
		err = msgpack.Unmarshal(b, &v)
		if err != nil {
			panic(err)
		}
		fmt.Println(v.(*EventTime).UTC())
	}

	// Output: 1973-11-29 21:33:10.000000123 +0000 UTC
	// 1973-11-29 21:33:10.000000123 +0000 UTC
}
