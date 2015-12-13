# MessagePack encoding for Golang [![Build Status](https://travis-ci.org/vmihailenco/msgpack.svg)](https://travis-ci.org/vmihailenco/msgpack)

Supports:
- Primitives, arrays, maps, structs, time.Time and interface{}.
- Appengine *datastore.Key and datastore.Cursor.
- [CustomEncoder](http://godoc.org/gopkg.in/vmihailenco/msgpack.v2#example-CustomEncoder)/CustomDecoder interfaces for custom encoding.
- [Extensions](http://godoc.org/gopkg.in/vmihailenco/msgpack.v2#example-RegisterExt) to encode type information.
- Fields renaming, e.g. `msgpack:"my_field_name"`.
- Structs inlining, e.g. `msgpack:",inline"`.
- Omitempty flag, e.g. `msgpack:",omitempty"`.

API docs: http://godoc.org/gopkg.in/vmihailenco/msgpack.v2.
Examples: http://godoc.org/gopkg.in/vmihailenco/msgpack.v2#pkg-examples.

## Installation

Install:

    go get gopkg.in/vmihailenco/msgpack.v2

## Quickstart

```go
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
```

## Benchmark

```
BenchmarkStruct-4               	  200000	     11515 ns/op	    3296 B/op	      27 allocs/op
BenchmarkStructUgorjiGoMsgpack-4	  100000	     12234 ns/op	    3840 B/op	      70 allocs/op
BenchmarkStructUgorjiGoCodec-4  	  100000	     15251 ns/op	    7474 B/op	      29 allocs/op
BenchmarkStructJSON-4           	   30000	     50851 ns/op	    8088 B/op	      29 allocs/op
BenchmarkStructGOB-4            	   20000	     64993 ns/op	   15609 B/op	     299 allocs/op
```

## Howto

Please go through [examples](http://godoc.org/gopkg.in/vmihailenco/msgpack.v2#pkg-examples) to get an idea how to use this package.
