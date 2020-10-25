package msgpack_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
)

type NoIntern struct {
	A string
	B string
	C string
}

type Intern struct {
	A string      `msgpack:",intern"`
	B string      `msgpack:",intern"`
	C interface{} `msgpack:",intern"`
}

func TestInternedString(t *testing.T) {
	var buf bytes.Buffer

	enc := msgpack.NewEncoder(&buf)
	enc.UseInternedStrings(true)

	dec := msgpack.NewDecoder(&buf)
	dec.UseInternedStrings(true)

	for i := 0; i < 2; i++ {
		err := enc.EncodeString("hello")
		require.Nil(t, err)
	}
	err := enc.Encode("hello")
	require.Nil(t, err)

	s, err := dec.DecodeString()
	require.Nil(t, err)
	require.Equal(t, "hello", s)

	s, err = dec.DecodeString()
	require.Nil(t, err)
	require.Equal(t, "hello", s)

	v, err := dec.DecodeInterface()
	require.Nil(t, err)
	require.Equal(t, "hello", v)

	_, err = dec.DecodeInterface()
	require.Equal(t, io.EOF, err)
}

func TestInternedStringTag(t *testing.T) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	dec := msgpack.NewDecoder(&buf)

	in := []Intern{
		{"f", "f", "f"},
		{"fo", "fo", "fo"},
		{"foo", "foo", "foo"},
		{"f", "fo", "foo"},
	}
	err := enc.Encode(in)
	require.Nil(t, err)

	var out []Intern
	err = dec.Decode(&out)
	require.Nil(t, err)
	require.Equal(t, in, out)
}

func TestResetDict(t *testing.T) {
	dict := []string{"hello world"}

	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	dec := msgpack.NewDecoder(&buf)

	{
		enc.ResetDict(&buf, dict)
		err := enc.EncodeString("hello world")
		require.Nil(t, err)
		require.Equal(t, 3, buf.Len())

		dec.ResetDict(&buf, dict)
		s, err := dec.DecodeString()
		require.Nil(t, err)
		require.Equal(t, "hello world", s)
	}

	{
		enc.ResetDict(&buf, dict)
		err := enc.EncodeString("hello world")
		require.Nil(t, err)
		require.Equal(t, 3, buf.Len())

		dec.ResetDict(&buf, dict)
		s, err := dec.DecodeInterface()
		require.Nil(t, err)
		require.Equal(t, "hello world", s)
	}

	dec.ResetDict(&buf, dict)
	_ = enc.EncodeString("xxxx")
	require.Equal(t, 5, buf.Len())
	_ = enc.EncodeString("xxxx")
	require.Equal(t, 10, buf.Len())
}
