package msgpack

import "reflect"

func setBytesLen(b []byte, n int) []byte {
	if n <= cap(b) {
		return b[:n]
	}
	b = b[:cap(b)]
	b = append(b, make([]byte, n-cap(b))...)
	return b
}

func setStringsLen(s []string, n int) []string {
	if n <= cap(s) {
		return s[:n]
	}
	s = s[:cap(s)]
	s = append(s, make([]string, n-cap(s))...)
	return s
}

func setSliceValueLen(v reflect.Value, n int) reflect.Value {
	if n <= v.Cap() {
		return v.Slice(n, n)
	}
	v = v.Slice(v.Cap(), v.Cap())
	diff := n - v.Cap()
	return reflect.AppendSlice(v, reflect.MakeSlice(v.Type(), diff, diff))
}
