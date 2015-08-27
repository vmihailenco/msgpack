Msgpack implementation for Golang [![Build Status](https://travis-ci.org/vmihailenco/msgpack.svg)](https://travis-ci.org/vmihailenco/msgpack)
===

Supports:
- Primitives, arrays, maps, structs and interface{}.
- time.Time.
- Appengine *datastore.Key and datastore.Cursor.
- Marshaler/Unmarshaler interfaces for custom encoding.
- Struct field renaming, e.g. `msgpack:"my_field_name"`.
- Struct field inlining, e.g. `msgpack:",inline"`.
- Omitempty flag, e.g. `msgpack:",omitempty"`.

API docs: http://godoc.org/gopkg.in/vmihailenco/msgpack.v2

Installation
------------

Install:

    go get gopkg.in/vmihailenco/msgpack.v2

Usage
-----

Examples:

    func ExampleEncode() {
        b, err := msgpack.Marshal(true)
        fmt.Printf("%v %#v\n", err, b)
        // Output: <nil> []byte{0xc3}
    }

    func ExampleDecode() {
        var out bool
        err := msgpack.Unmarshal([]byte{0xc3}, &out)
        fmt.Println(err, out)
        // Output: <nil> true
    }

    func ExampleMapStringInterface() {
        in := map[string]interface{}{"foo": 1, "hello": "world"}
        b, err := msgpack.Marshal(in)
        _ = err

        var out map[string]interface{}
        err = msgpack.Unmarshal(b, &out)
        fmt.Printf("%v %#v\n", err, out)
        // Output: <nil> map[string]interface {}{"foo":1, "hello":"world"}
    }

    func ExampleRecursiveMapStringInterface() {
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
