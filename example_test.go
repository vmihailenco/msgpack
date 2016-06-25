package msgpack_test

import (
	"bytes"
	"fmt"

	"gopkg.in/vmihailenco/msgpack.v2"
)

func ExampleMarshal() {
	type Item struct {
		Foo string
	}

	b, err := msgpack.Marshal(&Item{Foo: "bar"})
	if err != nil {
		panic(err)
	}

	var item Item
	err = msgpack.Unmarshal(b, &item)
	if err != nil {
		panic(err)
	}
	fmt.Println(item.Foo)
	// Output: bar
}

func ExampleRegisterExt() {
	type Item struct {
		S string
	}

	msgpack.RegisterExt(1, Item{})

	b, err := msgpack.Marshal(&Item{S: "string"})
	if err != nil {
		panic(err)
	}

	var v interface{}
	err = msgpack.Unmarshal(b, &v)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", v)
	// Output: msgpack_test.Item{S:"string"}
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
	buf := new(bytes.Buffer)

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

func ExampleDecoder_Query() {
	b, err := msgpack.Marshal([]map[string]interface{}{
		{"id": 1, "attrs": map[string]interface{}{"phone": 12345}},
		{"id": 2, "attrs": map[string]interface{}{"phone": 54321}},
	})
	if err != nil {
		panic(err)
	}

	dec := msgpack.NewDecoder(bytes.NewBuffer(b))
	values, err := dec.Query("*.attrs.phone")
	if err != nil {
		panic(err)
	}
	fmt.Println("phones are", values)

	dec.Reset(bytes.NewBuffer(b))
	values, err = dec.Query("1.attrs.phone")
	if err != nil {
		panic(err)
	}
	fmt.Println("2nd phone is", values[0])
	// Output: phones are [12345 54321]
	// 2nd phone is 54321
}
