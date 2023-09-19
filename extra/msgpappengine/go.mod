module github.com/vmihailenco/msgpack/extra/appengine

go 1.15

replace github.com/vmihailenco/msgpack/v6 => ../..

require (
	github.com/vmihailenco/msgpack/v6 v6.0.0
	google.golang.org/appengine v1.6.7
)
