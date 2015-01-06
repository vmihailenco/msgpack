all:
	go test gopkg.in/vmihailenco/msgpack.v2 -cpu=1
	go test gopkg.in/vmihailenco/msgpack.v2 -cpu=2
	go test gopkg.in/vmihailenco/msgpack.v2 -short -race
