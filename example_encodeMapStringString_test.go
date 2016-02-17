package msgpack_test

import (
	"bytes"
	"fmt"

	"gopkg.in/vmihailenco/msgpack.v2"
)

func Example_encodeMapStringString() {
	buf := &bytes.Buffer{}

	m := map[string]string{"foo1": "bar1", "foo2": "bar2", "foo3": "bar3"}
	keys := []string{"foo1", "foo3"}

	encodedMap, err := encodeMap(m, keys...)
	if err != nil {
		panic(err)
	}

	_, err = buf.Write(encodedMap)
	if err != nil {
		panic(err)
	}

	decoder := msgpack.NewDecoder(buf)
	value, err := decoder.DecodeMap()
	if err != nil {
		panic(err)
	}

	decodedMapValue := value.(map[interface{}]interface{})

	for _, key := range keys {
		fmt.Printf("%#v: %#v, ", key, decodedMapValue[key])
	}

	// Output: "foo1": "bar1", "foo3": "bar3",
}

func Example_decodeMapStringString() {
	decodedMap := make(map[string]string)
	buf := &bytes.Buffer{}

	m := map[string]string{"foo1": "bar1", "foo2": "bar2", "foo3": "bar3"}
	keys := []string{"foo1", "foo3", "foo2"}

	encodedMap, err := encodeMap(m, keys...)
	if err != nil {
		panic(err)
	}

	_, err = buf.Write(encodedMap)
	if err != nil {
		panic(err)
	}

	decoder := msgpack.NewDecoder(buf)

	n, err := decoder.DecodeMapLen()
	if err != nil {
		panic(err)
	}

	for i := 0; i < n; i++ {
		key, err := decoder.DecodeString()
		if err != nil {
			panic(err)
		}
		value, err := decoder.DecodeString()
		if err != nil {
			panic(err)
		}
		decodedMap[key] = value
	}

	for _, key := range keys {
		fmt.Printf("%#v: %#v, ", key, decodedMap[key])
	}
	// Output: "foo1": "bar1", "foo3": "bar3", "foo2": "bar2",
}

func encodeMap(m map[string]string, keys ...string) ([]byte, error) {
	buf := &bytes.Buffer{}
	encoder := msgpack.NewEncoder(buf)

	if err := encoder.EncodeMapLen(len(keys)); err != nil {
		return nil, err
	}

	for _, key := range keys {
		if err := encoder.EncodeString(key); err != nil {
			return nil, err
		}
		if err := encoder.EncodeString(m[key]); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}
