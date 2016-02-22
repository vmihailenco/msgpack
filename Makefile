all:
	go test gopkg.in/vmihailenco/msgpack.v2 -cpu=1
	go test gopkg.in/vmihailenco/msgpack.v2 -cpu=2
	go test gopkg.in/vmihailenco/msgpack.v2 -short -race

dev:
	go build -tags=dev
	go test -tags=dev -cpu=1 .
	go test -tags=dev -cpu=2 .
	go test -tags=dev -short -race .
