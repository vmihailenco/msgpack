package msgpack_test

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"encoding/gob"
	"encoding/json"
	"math"
	"testing"
	"time"

	gomsgpack "github.com/ugorji/go-msgpack"
	gocodec "github.com/ugorji/go/codec"

	"gopkg.in/vmihailenco/msgpack.v2"
)

func benchmarkEncodeDecode(b *testing.B, src, dst interface{}) {
	var buf bytes.Buffer
	dec := msgpack.NewDecoder(&buf)
	enc := msgpack.NewEncoder(&buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := enc.Encode(src); err != nil {
			b.Fatal(err)
		}
		if err := dec.Decode(dst); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkEncodeDecodeUgorjiGoCodec(b *testing.B, src, dst interface{}) {
	var buf bytes.Buffer
	h := &gocodec.MsgpackHandle{}
	enc := gocodec.NewEncoder(&buf, h)
	dec := gocodec.NewDecoder(&buf, h)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := enc.Encode(src); err != nil {
			b.Fatal(err)
		}
		if err := dec.Decode(dst); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBool(b *testing.B) {
	var dst bool
	benchmarkEncodeDecode(b, true, &dst)
}

func BenchmarkInt0(b *testing.B) {
	var dst int
	benchmarkEncodeDecode(b, 1, &dst)
}

func BenchmarkInt1(b *testing.B) {
	var dst int
	benchmarkEncodeDecode(b, -33, &dst)
}

func BenchmarkInt2(b *testing.B) {
	var dst int
	benchmarkEncodeDecode(b, 128, &dst)
}

func BenchmarkInt4(b *testing.B) {
	var dst int
	benchmarkEncodeDecode(b, 32768, &dst)
}

func BenchmarkInt8(b *testing.B) {
	var dst int
	benchmarkEncodeDecode(b, 2147483648, &dst)
}

func BenchmarkInt0Binary(b *testing.B) {
	var buf bytes.Buffer
	var out int32
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := binary.Write(&buf, binary.BigEndian, int32(1)); err != nil {
			b.Fatal(err)
		}
		if err := binary.Read(&buf, binary.BigEndian, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInt0UgorjiGoMsgpack(b *testing.B) {
	buf := &bytes.Buffer{}
	dec := gomsgpack.NewDecoder(buf, nil)
	enc := gomsgpack.NewEncoder(buf)
	var out int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := enc.Encode(1); err != nil {
			b.Fatal(err)
		}
		if err := dec.Decode(&out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInt0UgorjiGoCodec(b *testing.B) {
	var dst int
	benchmarkEncodeDecodeUgorjiGoCodec(b, 1, &dst)
}

func BenchmarkTime(b *testing.B) {
	var dst time.Time
	benchmarkEncodeDecode(b, time.Now(), &dst)
}

func BenchmarkDuration(b *testing.B) {
	var dst time.Duration
	benchmarkEncodeDecode(b, time.Hour, &dst)
}

func BenchmarkByteSlice(b *testing.B) {
	src := make([]byte, 1024)
	var dst []byte
	benchmarkEncodeDecode(b, src, &dst)
}

func BenchmarkByteArray(b *testing.B) {
	var src [1024]byte
	var dst [1024]byte
	benchmarkEncodeDecode(b, src, &dst)
}

func BenchmarkByteSliceUgorjiGoCodec(b *testing.B) {
	src := make([]byte, 1024)
	var dst []byte
	benchmarkEncodeDecodeUgorjiGoCodec(b, src, &dst)
}

func BenchmarkByteArrayUgorjiGoCodec(b *testing.B) {
	var src [1024]byte
	var dst [1024]byte
	benchmarkEncodeDecodeUgorjiGoCodec(b, src, &dst)
}

func BenchmarkMapStringString(b *testing.B) {
	src := map[string]string{
		"hello": "world",
		"foo":   "bar",
	}
	var dst map[string]string
	benchmarkEncodeDecode(b, src, &dst)
}

func BenchmarkMapStringStringPtr(b *testing.B) {
	src := map[string]string{
		"hello": "world",
		"foo":   "bar",
	}
	var dst map[string]string
	dstptr := &dst
	benchmarkEncodeDecode(b, src, &dstptr)
}

func BenchmarkMapIntInt(b *testing.B) {
	src := map[int]int{
		1: 10,
		2: 20,
	}
	var dst map[int]int
	benchmarkEncodeDecode(b, src, &dst)
}

func BenchmarkStringSlice(b *testing.B) {
	src := []string{"hello", "world"}
	var dst []string
	benchmarkEncodeDecode(b, src, &dst)
}

func BenchmarkStringSlicePtr(b *testing.B) {
	src := []string{"hello", "world"}
	var dst []string
	dstptr := &dst
	benchmarkEncodeDecode(b, src, &dstptr)
}

type benchmarkStruct struct {
	Name      string
	Age       int
	Colors    []string
	Data      []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

type benchmarkStruct2 struct {
	Name      string
	Age       int
	Colors    []string
	Data      []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	_ msgpack.CustomEncoder = &benchmarkStruct2{}
	_ msgpack.CustomDecoder = &benchmarkStruct2{}
)

func (s *benchmarkStruct2) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.Encode(
		s.Name,
		s.Colors,
		s.Age,
		s.Data,
		s.CreatedAt,
		s.UpdatedAt,
	)
}

func (s *benchmarkStruct2) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.Decode(
		&s.Name,
		&s.Colors,
		&s.Age,
		&s.Data,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
}

func structForBenchmark() *benchmarkStruct {
	return &benchmarkStruct{
		Name:      "Hello World",
		Colors:    []string{"red", "orange", "yellow", "green", "blue", "violet"},
		Age:       math.MaxInt32,
		Data:      make([]byte, 1024),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func structForBenchmark2() *benchmarkStruct2 {
	return &benchmarkStruct2{
		Name:      "Hello World",
		Colors:    []string{"red", "orange", "yellow", "green", "blue", "violet"},
		Age:       math.MaxInt32,
		Data:      make([]byte, 1024),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func BenchmarkStructVmihailencoMsgpack(b *testing.B) {
	in := structForBenchmark()
	out := &benchmarkStruct{}
	for i := 0; i < b.N; i++ {
		buf, err := msgpack.Marshal(in)
		if err != nil {
			b.Fatal(err)
		}

		err = msgpack.Unmarshal(buf, out)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStructMarshal(b *testing.B) {
	in := structForBenchmark()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := msgpack.Marshal(in)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStructUnmarshal(b *testing.B) {
	in := structForBenchmark()
	buf, err := msgpack.Marshal(in)
	if err != nil {
		b.Fatal(err)
	}
	out := &benchmarkStruct{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = msgpack.Unmarshal(buf, out)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStructManual(b *testing.B) {
	in := structForBenchmark2()
	out := &benchmarkStruct2{}
	for i := 0; i < b.N; i++ {
		buf, err := msgpack.Marshal(in)
		if err != nil {
			b.Fatal(err)
		}

		err = msgpack.Unmarshal(buf, out)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStructUgorjiGoMsgpack(b *testing.B) {
	in := structForBenchmark()
	out := &benchmarkStruct{}
	for i := 0; i < b.N; i++ {
		buf, err := gomsgpack.Marshal(in)
		if err != nil {
			b.Fatal(err)
		}

		err = gomsgpack.Unmarshal(buf, out, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStructUgorjiGoCodec(b *testing.B) {
	in := structForBenchmark()
	out := &benchmarkStruct{}
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		enc := gocodec.NewEncoder(buf, &gocodec.MsgpackHandle{})
		dec := gocodec.NewDecoder(buf, &gocodec.MsgpackHandle{})

		if err := enc.Encode(in); err != nil {
			b.Fatal(err)
		}
		if err := dec.Decode(out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStructJSON(b *testing.B) {
	in := structForBenchmark()
	out := &benchmarkStruct{}
	for i := 0; i < b.N; i++ {
		buf, err := json.Marshal(in)
		if err != nil {
			b.Fatal(err)
		}

		err = json.Unmarshal(buf, out)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStructGOB(b *testing.B) {
	in := structForBenchmark()
	out := &benchmarkStruct{}
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		enc := gob.NewEncoder(buf)
		dec := gob.NewDecoder(buf)

		if err := enc.Encode(in); err != nil {
			b.Fatal(err)
		}

		if err := dec.Decode(out); err != nil {
			b.Fatal(err)
		}
	}
}

type benchmarkSubStruct struct {
	Name string
	Age  int
}

func BenchmarkStructUnmarshalPartially(b *testing.B) {
	in := structForBenchmark()
	buf, err := msgpack.Marshal(in)
	if err != nil {
		b.Fatal(err)
	}
	out := &benchmarkSubStruct{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = msgpack.Unmarshal(buf, out)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCSV(b *testing.B) {
	for i := 0; i < b.N; i++ {
		record := []string{"1", "hello", "world"}

		buf := &bytes.Buffer{}
		r := csv.NewReader(buf)
		w := csv.NewWriter(buf)

		if err := w.Write(record); err != nil {
			b.Fatal(err)
		}
		w.Flush()
		if _, err := r.Read(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCSVMsgpack(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var num int
		var hello, world string

		buf := &bytes.Buffer{}
		enc := msgpack.NewEncoder(buf)
		dec := msgpack.NewDecoder(buf)

		if err := enc.Encode(1, "hello", "world"); err != nil {
			b.Fatal(err)
		}
		if err := dec.Decode(&num, &hello, &world); err != nil {
			b.Fatal(err)
		}
	}
}
