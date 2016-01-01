package msgpack_test

import (
	"bytes"
	"fmt"

	"gopkg.in/vmihailenco/msgpack.v2"
)

func ExampleMarshal() {
	b, err := msgpack.Marshal(true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", b)
	// Output:

	var out bool
	err = msgpack.Unmarshal([]byte{0xc3}, &out)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
	// Output: []byte{0xc3}
	// true
}

type myStruct struct {
	S string
}

func ExampleRegisterExt() {
	msgpack.RegisterExt(1, myStruct{})

	b, err := msgpack.Marshal(&myStruct{S: "string"})
	if err != nil {
		panic(err)
	}

	var v interface{}
	err = msgpack.Unmarshal(b, &v)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", v)
	// Output: msgpack_test.myStruct{S:"string"}
}

func Example_mapStringInterface() {
	in := map[string]interface{}{"foo": 1, "hello": "world"}
	b, err := msgpack.Marshal(in)
	if err != nil {
		panic(err)
	}

	var out map[string]interface{}
	err = msgpack.Unmarshal(b, &out)
	if err != nil {
		panic(err)
	}

	fmt.Println("foo =", out["foo"])
	fmt.Println("hello =", out["hello"])

	// Output:
	// foo = 1
	// hello = world
}

func Example_recursiveMapStringInterface() {
	buf := &bytes.Buffer{}

	enc := msgpack.NewEncoder(buf)
	in := map[string]interface{}{"foo": map[string]interface{}{"hello": "world"}}
	_ = enc.Encode(in)

	dec := msgpack.NewDecoder(buf)
	dec.DecodeMapFunc = func(d *msgpack.Decoder) (interface{}, error) {
		n, err := d.DecodeMapLen()
		if err != nil {
			return nil, err
		}

		m := make(map[string]interface{}, n)
		for i := 0; i < n; i++ {
			mk, err := d.DecodeString()
			if err != nil {
				return nil, err
			}

			mv, err := d.DecodeInterface()
			if err != nil {
				return nil, err
			}

			m[mk] = mv
		}
		return m, nil
	}
	out, err := dec.DecodeInterface()
	fmt.Printf("%v %#v\n", err, out)
	// Output: <nil> map[string]interface {}{"foo":map[string]interface {}{"hello":"world"}}
}
