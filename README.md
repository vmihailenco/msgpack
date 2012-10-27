Msgpack implementation for Golang
=================================

Supports:
- Primitives, arrays, maps and structs.
- time.Time.
- Appengine: *datastore.Key and datastore.Cursor.
- Extensions for user defined types.

Does not support:
- Interface unmarshalling.

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
