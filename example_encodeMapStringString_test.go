package msgpack_test

import (
	"bytes"
	"fmt"

	"gopkg.in/vmihailenco/msgpack.v2"
)

func Example_encodeMapStringString() {
	m := map[string]string{"foo1": "bar1", "foo2": "bar2", "foo3": "bar3"}
	keys := []string{"foo1", "foo3"}

	buf := &bytes.Buffer{}
	encoder := msgpack.NewEncoder(buf)

	if err := encoder.EncodeMapLen(len(keys)); err != nil {
		panic(err)
	}

	for _, key := range keys {
		if err := encoder.EncodeString(key); err != nil {
			panic(err)
		}
		if err := encoder.EncodeString(m[key]); err != nil {
			panic(err)
		}
	}

	decoder := msgpack.NewDecoder(buf)
	decodedMap, err := decoder.DecodeMap()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", decodedMap)
	// Output: map[interface {}]interface {}{"foo1":"bar1", "foo3":"bar3"}
}
