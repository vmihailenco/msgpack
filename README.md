Msgpack implementation for Golang
=================================

Supports:
- Primitives, arrays, maps, structs and interface{}.
- time.Time.
- Appengine *datastore.Key and datastore.Cursor.
- Extensions for user defined types.
- Tags.

API docs: http://godoc.org/github.com/vmihailenco/msgpack

Installation
------------

Install:

    go get github.com/vmihailenco/msgpack

Usage
-----

Example:

    import (
        "fmt"

        "github.com/vmihailenco/msgpack"
    )

    func main() {
        b, err := msgpack.Marshal(true)
        if err != nil {
            panic(err)
        }

        var out bool
        if err := msgpack.Unmarshal(b, &out); err != nil {
            panic(err)
        }
        fmt.Println(out)
    }

Extensions
----------

Look at [appengine.go](https://github.com/vmihailenco/msgpack/blob/master/appengine.go) for example.
