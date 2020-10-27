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
	C interface{}
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

	for i := 0; i < 3; i++ {
		err := enc.EncodeString("hello")
		require.Nil(t, err)
	}

	for i := 0; i < 3; i++ {
		s, err := dec.DecodeString()
		require.Nil(t, err)
		require.Equal(t, "hello", s)
	}

	err := enc.Encode("hello")
	require.Nil(t, err)

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
	dict := []string{"hello world", "foo bar"}

	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	dec := msgpack.NewDecoder(&buf)

	{
		enc.ResetDict(&buf, dictMap(dict))
		err := enc.EncodeString("hello world")
		require.Nil(t, err)
		require.Equal(t, 3, buf.Len())

		dec.ResetDict(&buf, dict)
		s, err := dec.DecodeString()
		require.Nil(t, err)
		require.Equal(t, "hello world", s)
	}

	{
		enc.ResetDict(&buf, dictMap(dict))
		err := enc.Encode("foo bar")
		require.Nil(t, err)
		require.Equal(t, 3, buf.Len())

		dec.ResetDict(&buf, dict)
		s, err := dec.DecodeInterface()
		require.Nil(t, err)
		require.Equal(t, "foo bar", s)
	}

	dec.ResetDict(&buf, dict)
	_ = enc.EncodeString("xxxx")
	require.Equal(t, 5, buf.Len())
	_ = enc.Encode("xxxx")
	require.Equal(t, 10, buf.Len())
}

func TestMapWithInternedString(t *testing.T) {
	type M map[string]interface{}

	dict := []string{"hello world", "foo bar"}

	var buf bytes.Buffer

	enc := msgpack.NewEncoder(nil)
	enc.ResetDict(&buf, dictMap(dict))

	dec := msgpack.NewDecoder(nil)
	dec.ResetDict(&buf, dict)

	for i := 0; i < 100; i++ {
		in := M{
			"foo bar":     "hello world",
			"hello world": "foo bar",
			"foo":         "bar",
		}
		err := enc.Encode(in)
		require.Nil(t, err)

		_, err = dec.DecodeInterface()
		require.Nil(t, err)
	}
}

func dictMap(dict []string) map[string]int {
	m := make(map[string]int, len(dict))
	for i, s := range dict {
		m[s] = i
	}
	return m
}
