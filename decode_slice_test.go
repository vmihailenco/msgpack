package msgpack_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
)

func TestIssue401_UnboundedSliceAllocation(t *testing.T) {
	// Array32 header with length 4294967295 (0xffffffff) and no element data.
	maliciousPayload := []byte{0xdd, 0xff, 0xff, 0xff, 0xff}

	t.Run("int64 slice", func(t *testing.T) {
		var s []int64
		err := msgpack.Unmarshal(maliciousPayload, &s)
		require.Error(t, err)
	})

	t.Run("string slice", func(t *testing.T) {
		var s []string
		err := msgpack.Unmarshal(maliciousPayload, &s)
		require.Error(t, err)
	})

	t.Run("interface slice", func(t *testing.T) {
		var s []interface{}
		err := msgpack.Unmarshal(maliciousPayload, &s)
		require.Error(t, err)
	})

	t.Run("interface target", func(t *testing.T) {
		var v interface{}
		err := msgpack.Unmarshal(maliciousPayload, &v)
		require.Error(t, err)
	})

	t.Run("DecodeSlice direct", func(t *testing.T) {
		dec := msgpack.NewDecoder(bytes.NewReader(maliciousPayload))
		s, err := dec.DecodeSlice()
		require.Error(t, err)
		require.Nil(t, s)
	})
}

func TestTruncatedSliceHeader(t *testing.T) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)

	err := enc.EncodeArrayLen(50000)
	require.NoError(t, err)

	err = enc.EncodeInt(1)
	require.NoError(t, err)
	err = enc.EncodeInt(2)
	require.NoError(t, err)
	err = enc.EncodeInt(3)
	require.NoError(t, err)

	var s []int
	err = msgpack.Unmarshal(buf.Bytes(), &s)
	require.Error(t, err)
}

func TestSliceDecodeRoundTrip(t *testing.T) {
	t.Run("small slice", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		data, err := msgpack.Marshal(input)
		require.NoError(t, err)

		var output []int
		err = msgpack.Unmarshal(data, &output)
		require.NoError(t, err)
		require.Equal(t, input, output)
	})

	t.Run("string slice", func(t *testing.T) {
		input := []string{"alpha", "bravo", "charlie"}
		data, err := msgpack.Marshal(input)
		require.NoError(t, err)

		var output []string
		err = msgpack.Unmarshal(data, &output)
		require.NoError(t, err)
		require.Equal(t, input, output)
	})

	t.Run("interface slice", func(t *testing.T) {
		input := []interface{}{int64(42), "text", true}
		data, err := msgpack.Marshal(input)
		require.NoError(t, err)

		dec := msgpack.NewDecoder(bytes.NewReader(data))
		output, err := dec.DecodeSlice()
		require.NoError(t, err)
		require.Equal(t, input, output)
	})
}

func TestStreamingReaderSliceDecode(t *testing.T) {
	// Custom reader that does not implement lenReader (Len() int).
	payload := []byte{0xdd, 0xff, 0xff, 0xff, 0xff}
	r := &nonLenReader{r: bytes.NewReader(payload)}

	dec := msgpack.NewDecoder(r)
	var s []int64
	err := dec.Decode(&s)
	require.Error(t, err)
}

func TestUnboundedMapAllocation(t *testing.T) {
	// Map32 header with length 4294967295 (0xffffffff) and no key-value data.
	maliciousMapPayload := []byte{0xdf, 0xff, 0xff, 0xff, 0xff}

	t.Run("unmarshal map", func(t *testing.T) {
		var m map[string]interface{}
		err := msgpack.Unmarshal(maliciousMapPayload, &m)
		require.Error(t, err)
	})

	t.Run("DecodeMap direct", func(t *testing.T) {
		dec := msgpack.NewDecoder(bytes.NewReader(maliciousMapPayload))
		m, err := dec.DecodeMap()
		require.Error(t, err)
		require.Nil(t, m)
	})

	t.Run("DecodeUntypedMap direct", func(t *testing.T) {
		dec := msgpack.NewDecoder(bytes.NewReader(maliciousMapPayload))
		m, err := dec.DecodeUntypedMap()
		require.Error(t, err)
		require.Nil(t, m)
	})
}

type nonLenReader struct {
	r io.Reader
}

func (nlr *nonLenReader) Read(p []byte) (int, error) {
	return nlr.r.Read(p)
}
